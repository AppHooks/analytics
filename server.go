package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

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
	HEADER_LOCATION     = "Location"
	HEADER_CONTENT_TYPE = "Content-Type"

	SESSION_USER_KEY = "user"

	USER_REGISTER_PAGE = "/users/register.html"
	USER_LOGIN_PAGE    = "/users/login.html"
	SERVICE_LIST_PAGE  = "/services/list.html"
	SERVICE_ADD_PAGE   = "/services/add.html"

	SERVICE_ADD_URL   = "/services/add"
	USER_REGISTER_URL = "/users/register"
	USER_LOGIN_URL    = "/users/login"
)

type Factory func(configuration map[string]interface{}) Service

func Analytics(db *gorm.DB, m *martini.ClassicMartini, services map[string]Factory) {
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
			res.Header().Set(HEADER_LOCATION, SERVICE_LIST_PAGE)
			res.WriteHeader(http.StatusFound)
			return
		}

		c.Next()
	}

	requiredLoggedIn := func(c martini.Context, res http.ResponseWriter, session Session) {
		if session.Get(SESSION_USER_KEY) == nil {
			res.Header().Set(HEADER_LOCATION, USER_LOGIN_PAGE)
			res.WriteHeader(http.StatusFound)
			return
		}

		c.Next()
	}

	m.Get("/", alreadyLoggedIn, func(res http.ResponseWriter) {
		res.Header().Set(HEADER_LOCATION, USER_LOGIN_PAGE)
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
				res.Header().Set(HEADER_LOCATION, SERVICE_LIST_PAGE)
			} else {
				res.Header().Set(HEADER_LOCATION, USER_REGISTER_PAGE)
			}
			res.WriteHeader(http.StatusFound)
		})
		r.Post("/login", alreadyLoggedIn, func(res http.ResponseWriter, req *http.Request, session Session) {
			user := models.GetUserFromEmail(db, req.FormValue("email"))
			if user != nil && user.Authenticate(req.FormValue("password")) {
				session.Set(SESSION_USER_KEY, user.Id)
				res.Header().Set(HEADER_LOCATION, SERVICE_LIST_PAGE)
			} else {
				res.Header().Set(HEADER_LOCATION, USER_LOGIN_PAGE)
			}
			res.WriteHeader(http.StatusFound)

		})
		r.Get("/logout", func(res http.ResponseWriter, session Session, data acerender.TemplateData) {
			session.Clear()
			data.Clear()
			res.Header().Set(HEADER_LOCATION, USER_LOGIN_PAGE)
			res.WriteHeader(http.StatusFound)
		})
	})
	m.Group("/services", func(r martini.Router) {

		r.Get("/:page.html", func(params martini.Params, r acerender.Render) {
			r.AceOk("layout:services_"+params["page"], nil)
		})

		r.Post("/add", func(res http.ResponseWriter, req *http.Request, data acerender.TemplateData) {
			user := data.Get(SESSION_USER_KEY).(*models.User)
			service := models.NewService(db, req.FormValue("name"), req.FormValue("type"), map[string]interface{}{
				"key": req.FormValue("key"),
			})
			db.Model(user).Association("Services").Append(service)

			header := res.Header()
			header.Set(HEADER_LOCATION, SERVICE_LIST_PAGE)
			header.Set(HEADER_CONTENT_TYPE, "application/json")
			res.WriteHeader(http.StatusFound)

			output, _ := json.Marshal(map[string]interface{}{"success": true})
			fmt.Fprintf(res, string(output))
		})

		r.Post("/send/:key", func(params martini.Params, res http.ResponseWriter, req *http.Request) {
			user := models.GetUserFromKey(db, params["key"])
			userServices := models.GetServicesForUser(db, user)

			decoder := json.NewDecoder(req.Body)
			var input map[string]interface{}
			decoder.Decode(&input)

			targets := []Service{}
			for _, userService := range userServices {
				data, exists := services[userService.Type]
				if exists {
					targets = append(targets, data(userService.GetConfiguration()))
				}
			}

			aggregator := &Aggregator{targets}

			fmt.Fprintf(res, aggregator.Send(input, req.RemoteAddr))
		})

	}, requiredLoggedIn)

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
	db.CreateTable(models.Service{})

	m := martini.Classic()

	network := NetworkWrapper{}
	Analytics(&db, m, map[string]Factory{
		"mixpanel": func(configuration map[string]interface{}) Service {
			mixpanel := &Mixpanel{
				Network: network,
			}
			mixpanel.LoadConfiguration(configuration)
			return mixpanel
		},
	})
	m.Run()
}
