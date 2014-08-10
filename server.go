package main

import (
	"encoding/json"
	"fmt"
	"github.com/go-martini/martini"
	. "github.com/llun/analytics/services"
	"github.com/llun/martini-amber"
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
	m.Use(martini_amber.Renderer())
	m.Get("/", func(r martini_amber.Render) {
		r.AmberOK("index", nil)
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
