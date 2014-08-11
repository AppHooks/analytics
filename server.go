package main

import (
	"encoding/json"
	"fmt"
	"github.com/go-martini/martini"
	. "github.com/llun/analytics/services"
	"github.com/llun/martini-acerender"
	"net/http"
	"os"
)

func main() {

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
	m.Get("/login", func() {
	})
	m.Get("/logout", func() {
	})
	m.Post("/send", func(res http.ResponseWriter, req *http.Request) {
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
