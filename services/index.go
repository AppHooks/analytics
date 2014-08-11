package services

import (
	"io/ioutil"
	"net/http"
)

type Input struct {
	Event string
	Data  map[string]interface{}
	IP    string
}

type Output struct {
	Success bool
}

type Service interface {
	Send(in Input) Output
	GetName() string
}

type Network interface {
	Request(url string, data string) (status int, body string, err error)
}

type NetworkWrapper struct {
}

func (n NetworkWrapper) Request(url string, data string) (int, string, error) {
	resp, err := http.Get(url + "?" + data)
	defer resp.Body.Close()

	if err != nil {
		return 0, "", err
	}

	body, err := ioutil.ReadAll(resp.Body)
	return resp.StatusCode, string(body), err
}
