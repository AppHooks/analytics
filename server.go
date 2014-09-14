package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

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
	USER_PROFILE_PAGE  = "/users/profile.html"
	SERVICE_LIST_PAGE  = "/services/list.html"
	SERVICE_ADD_PAGE   = "/services/add.html"

	SERVICE_ADD_URL    = "/services/add"
	SERVICE_REMOVE_URL = "/services/remove"
	USER_REGISTER_URL  = "/users/register"
	USER_LOGIN_URL     = "/users/login"
	USER_UPDATE_URL    = "/users/update"
)

type Factory func(name string, configuration map[string]interface{}) Service

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
	m.Use(func(res http.ResponseWriter, c martini.Context, session Session, data acerender.TemplateData) {
		if userid := session.Get(SESSION_USER_KEY); userid != nil {
			user := models.GetUserFromId(db, userid.(int64))
			if user == nil {
				session.Delete(SESSION_USER_KEY)
				res.Header().Set(HEADER_LOCATION, SERVICE_LIST_PAGE)
				res.WriteHeader(http.StatusFound)
				return
			}

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

		r.Post("/update", requiredLoggedIn, func(res http.ResponseWriter, req *http.Request, data acerender.TemplateData) {
			email := strings.TrimSpace(req.FormValue("email"))

			current := req.FormValue("current")

			password := req.FormValue("password")
			confirm := req.FormValue("confirm")

			currentUser := data.Get(SESSION_USER_KEY).(*models.User)
			if len(email) > 0 && currentUser.Authenticate(current) && currentUser.UpdatePassword(password, confirm) {
				currentUser.Email = email
				currentUser.Save()
			}

			res.Header().Set(HEADER_LOCATION, USER_PROFILE_PAGE)
			res.WriteHeader(http.StatusFound)
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

		r.Get("/list.html", func(r acerender.Render, data acerender.TemplateData) {
			services := models.GetServicesForUser(db, data.Get(SESSION_USER_KEY).(*models.User))
			r.AceOk("layout:services_list", &acerender.AceData{
				map[string]interface{}{
					"services": services,
				},
			})
		})
		r.Get("/:page.html", func(params martini.Params, r acerender.Render, req *http.Request) {
			r.AceOk("layout:services_"+params["page"], nil)
		})

		r.Post("/add", func(res http.ResponseWriter, req *http.Request, data acerender.TemplateData) {
			body, _ := ioutil.ReadAll(req.Body)
			var input map[string]interface{}
			json.Unmarshal(body, &input)

			user := data.Get(SESSION_USER_KEY).(*models.User)
			service := models.NewService(db, input["name"].(string), input["type"].(string),
				input["config"].(map[string]interface{}))
			db.Model(user).Association("Services").Append(service)

			header := res.Header()
			header.Set(HEADER_LOCATION, SERVICE_LIST_PAGE)
			header.Set(HEADER_CONTENT_TYPE, "application/json")
			res.WriteHeader(http.StatusFound)

			output, _ := json.Marshal(map[string]interface{}{"success": true})
			fmt.Fprintf(res, string(output))
		})

		r.Get("/remove/:service", func(res http.ResponseWriter, params martini.Params) {
			service := models.GetServiceFromId(db, params["service"])
			service.Delete()

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
					targets = append(targets, data(userService.Name, userService.GetConfiguration()))
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
		"mixpanel": func(name string, configuration map[string]interface{}) Service {
			mixpanel := &Mixpanel{
				Name:    name,
				Network: network,
			}
			mixpanel.LoadConfiguration(configuration)
			return mixpanel
		},
	})
	m.Run()
}
