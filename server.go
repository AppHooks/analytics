package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-martini/martini"
	"github.com/jinzhu/gorm"
	"github.com/llun/analytics/models"
	"github.com/llun/martini-acerender"
	"github.com/yosssi/ace"

	. "github.com/llun/analytics/services"
	. "github.com/martini-contrib/sessions"

	_ "github.com/mattn/go-sqlite3"
)

const (
	SESSION_USER_KEY          = "user"
	SESSION_INTERNAL_USER_KEY = "_user"
)

func Analytics(db gorm.DB, m *martini.ClassicMartini) {

	network := NetworkWrapper{}
	services := []Service{}

	if len(os.Getenv("Mixpanel")) > 0 {
		mixpanel := Mixpanel{network, os.Getenv("Mixpanel")}
		services = append(services, &mixpanel)
	}

	aggregator := Aggregator{services}
	m.Use(acerender.Renderer(&ace.Options{BaseDir: "public/templates"}))

	store := NewCookieStore([]byte("secret"))
	store.Options(Options{
		Path:   "/",
		MaxAge: 86400000,
	})
	m.Use(Sessions("analytics", store))
	m.Use(func(c martini.Context, session Session) {
		if session.Get(SESSION_USER_KEY) != nil {
			user := models.GetUserFromId(&db, session.Get(SESSION_USER_KEY).(int64))
			session.Set(SESSION_INTERNAL_USER_KEY, user)
		}

		c.Next()
	})

	alreadyLoggedIn := func(c martini.Context, res http.ResponseWriter, session Session) {
		if session.Get(SESSION_USER_KEY) != nil {
			res.Header().Set("Location", "/services/list.html")
			res.WriteHeader(http.StatusFound)
			return
		}

		c.Next()
	}

	m.Get("/", alreadyLoggedIn, func(res http.ResponseWriter) {
		res.Header().Set("Location", "/users/login.html")
		res.WriteHeader(http.StatusFound)
	})
	m.Group("/users", func(r martini.Router) {

		r.Get("/:page.html", func(params martini.Params, r acerender.Render) {
			r.AceOk("layout:"+params["page"], nil)
		})
		r.Post("/register", alreadyLoggedIn, func(res http.ResponseWriter, req *http.Request, session Session) {
			email := req.PostFormValue("email")
			password := req.PostFormValue("password")

			if !models.IsUserExists(&db, email) {
				user, err := models.NewUser(&db, email, password)
				if err == nil {
					user.Save()
					res.Header().Set("Location", "/services/list.html")
				} else {
					res.Header().Set("Location", "/users/register.html")
				}
			} else {
				res.Header().Set("Location", "/users/register.html")
			}
			res.WriteHeader(http.StatusFound)
		})
		r.Post("/login", alreadyLoggedIn, func(res http.ResponseWriter, req *http.Request, session Session) {
			user := models.GetUserFromEmail(&db, req.PostFormValue("email"))
			if user.Authenticate(req.PostFormValue("password")) {
				session.Set(SESSION_USER_KEY, user.Id)
				res.Header().Set("Location", "/")
			} else {
				res.Header().Set("Location", "/users/login.html")
			}
			res.WriteHeader(http.StatusFound)

		})
		r.Get("/logout", func(res http.ResponseWriter, session Session) {
			session.Clear()
			res.Header().Set("Location", "/users/login.html")
			res.WriteHeader(http.StatusFound)
		})
	})

	m.Post("/analytics/send", func(res http.ResponseWriter, req *http.Request) {
		header := res.Header()
		header.Add("Content-Type", "application/json")
		res.WriteHeader(http.StatusOK)

		decoder := json.NewDecoder(req.Body)
		var input map[string]interface{}
		err := decoder.Decode(&input)
		if err != nil {
			panic(-1)
		}

		fmt.Fprintf(res, aggregator.Send(input, req.RemoteAddr))
	})
}

func main() {
	var (
		db  gorm.DB
		err error
	)
	if martini.Env == martini.Dev {
		db, err = gorm.Open("sqlite3", "/tmp/analytics.db")
	} else {
		// Use postgresql
	}

	if err != nil {
		log.Println(err)
		panic(-1)
	}

	db.CreateTable(models.User{})

	m := martini.Classic()
	Analytics(db, m)
	m.Run()
}
