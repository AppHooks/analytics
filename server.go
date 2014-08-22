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

func Analytics(m *martini.ClassicMartini) {
	db, err := gorm.Open("sqlite3", "/tmp/analytics.db")
	if err != nil {
		log.Println(err)
		panic(-1)
	}
	db.CreateTable(models.User{})

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

	m.Get("/", func(r acerender.Render, session Session) {
		r.AceOk("layout:index", session)
	})
	m.Group("/users", func(r martini.Router) {

		alreadyLoggedin := func(c martini.Context, res http.ResponseWriter, session Session) {
			if session.Get(SESSION_USER_KEY) != nil {
				res.Header().Set("Location", "/")
				res.WriteHeader(http.StatusFound)
				return
			}

			c.Next()
		}

		r.Get("/:page.html", func(params martini.Params, r acerender.Render) {
			r.AceOk("layout:"+params["page"], nil)
		})
		r.Post("/register", alreadyLoggedin, func(res http.ResponseWriter, req *http.Request, session Session) {
			email := req.PostFormValue("email")
			password := req.PostFormValue("password")

			if !models.IsUserExists(&db, email) {
				user := models.NewUser(&db, email, password)
				user.Save()

				res.Header().Set("Location", "/users/login.html")
			} else {
				res.Header().Set("Location", "/users/register.html")
			}
			res.WriteHeader(http.StatusFound)
		})
		r.Post("/login", alreadyLoggedin, func(res http.ResponseWriter, req *http.Request, session Session) {
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
	m := martini.Classic()
	Analytics(m)
	m.Run()
}
