package main_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"

	"github.com/go-martini/martini"
	"github.com/llun/analytics"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Server", func() {

	var m *martini.ClassicMartini

	BeforeEach(func() {
		m = martini.Classic()
		main.Analytics(m)
	})

	Describe("User register", func() {

		It("should redirect to index page when success", func() {
			data := url.Values{}
			data.Set("email", "user@email.com")
			data.Set("password", "password")
			data.Set("confirm", "password")

			res := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/users/register", bytes.NewBufferString(data.Encode()))
			req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
			req.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))

			m.ServeHTTP(res, req)

			Expect(res.Code).To(Equal(http.StatusFound))
			Expect(res.Header().Get("Location")).To(Equal("/services/list.html"))
		})

	})

})
