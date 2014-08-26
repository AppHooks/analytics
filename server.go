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
	SESSION_USER_KEY = "user"
)

type TemplateData map[string]interface{}

func Analytics(db *gorm.DB, m *martini.ClassicMartini) {

	network := NetworkWrapper{}
	services := []Service{}

	if len(os.Getenv("Mixpanel")) > 0 {
		mixpanel := Mixpanel{network, os.Getenv("Mixpanel")}
		services = append(services, &mixpanel)
	}

	aggregator := Aggregator{services}

	templateOptions := &ace.Options{BaseDir: "public/templates"}
	if martini.Env == martini.Dev {
		templateOptions.DynamicReload = true
	}
	m.Use(acerender.Ace(templateOptions))

	store := NewCookieStore([]byte("secret"))
	store.Options(Options{
		Path:   "/",
		MaxAge: 86400000,
	})
	m.Use(Sessions("analytics", store))
	m.Use(func(c martini.Context, session Session, data acerender.TemplateData) {
		if userid := session.Get(SESSION_USER_KEY); userid != nil {
			user := models.GetUserFromId(db, userid.(int64))
			data.Set(SESSION_USER_KEY, user)
		}

		c.Next()
	})

	alreadyLoggedIn := func(c martini.Context, res http.ResponseWriter, session Session) {
		if userid := session.Get(SESSION_USER_KEY); userid != nil {
			res.Header().Set("Location", "/services/list.html")
			res.WriteHeader(http.StatusFound)
			return
		}

		c.Next()
	}

	requiredLoggedIn := func(c martini.Context, res http.ResponseWriter, session Session) {
		if session.Get(SESSION_USER_KEY) == nil {
			res.Header().Set("Location", "/users/login.html")
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

		r.Get("/profile.html", requiredLoggedIn, func(r acerender.Render) {
			r.AceOk("layout:users_profile", nil)
		})
		r.Get("/:page.html", alreadyLoggedIn, func(params martini.Params, r acerender.Render) {
			r.AceOk("layout:users_"+params["page"], nil)
		})
		r.Post("/register", alreadyLoggedIn, func(res http.ResponseWriter, req *http.Request, session Session) {
			email := req.FormValue("email")
			password := req.FormValue("password")

			if user, err := models.NewUser(db, email, password); !models.IsUserExists(db, email) && err == nil {
				user.Save()
				res.Header().Set("Location", "/services/list.html")
			} else {
				res.Header().Set("Location", "/users/register.html")
			}
			res.WriteHeader(http.StatusFound)
		})
		r.Post("/login", alreadyLoggedIn, func(res http.ResponseWriter, req *http.Request, session Session) {
			user := models.GetUserFromEmail(db, req.FormValue("email"))
			if user != nil && user.Authenticate(req.FormValue("password")) {
				log.Printf("User %v is logged in\n", user.Email)
				session.Set(SESSION_USER_KEY, user.Id)
				res.Header().Set("Location", "/services/list.html")
			} else {
				res.Header().Set("Location", "/users/login.html")
			}
			res.WriteHeader(http.StatusFound)

		})
		r.Get("/logout", func(res http.ResponseWriter, session Session, data acerender.TemplateData) {
			session.Clear()
			data.Clear()
			res.Header().Set("Location", "/users/login.html")
			res.WriteHeader(http.StatusFound)
		})
	})
	m.Group("/services", func(r martini.Router) {

		r.Get("/:page.html", func(params martini.Params, r acerender.Render) {
			r.AceOk("layout:services_"+params["page"], nil)
		})

	}, requiredLoggedIn)

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
	Analytics(&db, m)
	m.Run()
}
