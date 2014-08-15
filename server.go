package main

import (
	"encoding/json"
	"fmt"
	"github.com/go-martini/martini"
	"github.com/jinzhu/gorm"
	"github.com/llun/analytics/models"
	. "github.com/llun/analytics/services"
	"github.com/llun/martini-acerender"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"net/http"
	"os"
)

func main() {

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
		services = append(services, mixpanel)
	}

	aggregator := Aggregator{services}

	m := martini.Classic()
	m.Use(acerender.Renderer(acerender.Options{BaseDir: "public/templates"}))
	m.Get("/", func(r acerender.Render) {
		r.AceOk("layout:index", nil)
	})
	m.Get("/users/register.html", func(r acerender.Render) {
		r.AceOk("layout:register", nil)
	})

	m.Post("/users/register", func(res http.ResponseWriter, req *http.Request) {
		username := req.PostFormValue("username")
		password := req.PostFormValue("password")

		if models.IsUserExists(&db, username, "") == false {
			user := models.NewUser(&db, username, password, "")
			user.Save()

			res.Header().Set("Location", "/")
		} else {
			res.Header().Set("Location", "/users/register.html")
		}
		res.WriteHeader(302)
	})
	m.Post("/users/login", func(req *http.Request) string {
		user := models.GetUserFromUsername(&db, req.PostFormValue("username"))
		log.Println(user)
		return "Hello"
	})
	m.Get("/users/logout", func() {
	})
	m.Post("/analytics/send", func(res http.ResponseWriter, req *http.Request) {
		header := res.Header()
		header.Add("Content-Type", "application/json")
		res.WriteHeader(200)

		decoder := json.NewDecoder(req.Body)
		var input map[string]interface{}
		err := decoder.Decode(&input)
		if err != nil {
			panic(-1)
		}

		fmt.Fprintf(res, aggregator.Send(input, req.RemoteAddr))
	})
	m.Run()
}
