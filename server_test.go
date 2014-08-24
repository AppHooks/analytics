package main_test

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strconv"

	"github.com/go-martini/martini"
	"github.com/jinzhu/gorm"
	"github.com/llun/analytics"
	"github.com/llun/analytics/models"

	. "github.com/martini-contrib/sessions"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func GenerateFixtures() (db gorm.DB) {
	os.Remove("/tmp/analytics.db")

	db, _ = gorm.Open("sqlite3", "/tmp/analytics.db")
	models.NewUser(&db, "admin@email.com", "password")
	models.NewService(&db, "")

	return
}

var _ = Describe("Server", func() {

	var m *martini.ClassicMartini

	BeforeEach(func() {
		m = martini.Classic()

		db := GenerateFixtures()
		main.Analytics(db, m)
	})

	Describe("Index", func() {

		It("should redirect to login when user doesn't login yet", func() {

			res := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/", nil)

			m.ServeHTTP(res, req)

			Expect(res.Code).To(Equal(http.StatusFound))
			Expect(res.Header().Get("Location")).To(Equal("/users/login.html"))

		})

		It("should redirect to list services when user is already login", func() {

			m.Use(func(c martini.Context, session Session) {
				// Emulate user is already logged in.
				session.Set(main.SESSION_USER_KEY, "user")
				c.Next()
			})
			res := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/", nil)

			m.ServeHTTP(res, req)

			Expect(res.Code).To(Equal(http.StatusFound))
			Expect(res.Header().Get("Location")).To(Equal("/services/list.html"))

		})

	})

	Describe("User login", func() {
		It("should redirect to index page when success", func() {
			data := url.Values{}
			data.Set("email", "admin@email.com")
			data.Set("password", "password")

			res := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/users/login", bytes.NewBufferString(data.Encode()))
			req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
			req.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))

			m.ServeHTTP(res, req)

			Expect(res.Code).To(Equal(http.StatusFound))
			Expect(res.Header().Get("Location")).To(Equal("/services/list.html"))
		})
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

		It("should redirect back to register page when validate email fail", func() {
			data := url.Values{}
			data.Set("email", "user")
			data.Set("password", "password")
			data.Set("confirm", "password")

			res := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/users/register", bytes.NewBufferString(data.Encode()))
			req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
			req.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))

			m.ServeHTTP(res, req)

			Expect(res.Code).To(Equal(http.StatusFound))
			Expect(res.Header().Get("Location")).To(Equal("/users/register.html"))

		})

	})

	Context("Services", func() {

		It("should add service to user object", func() {
		})

		It("should send data to user services", func() {
			prepare := map[string]interface{}{
				"event": "name",
				"data": map[string]interface{}{
					"key1": "value1",
					"key2": "value2",
				},
			}

			b, _ := json.Marshal(prepare)
			data := string(b)

			res := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/services/send/key", bytes.NewBufferString(data))
			req.Header.Add("Content-Type", "application/json")
			req.Header.Add("Content-Length", strconv.Itoa(len(data)))

			m.ServeHTTP(res, req)

			Expect(res.Code).To(Equal(http.StatusFound))

			body, _ := ioutil.ReadAll(res.Body)
			output := map[string]interface{}{
				"success": true,
				"services": map[string]interface{}{
					"service1": true,
					"service2": true,
				},
			}
			outputBytes, _ := json.Marshal(output)
			Expect(string(body)).To(Equal(string(outputBytes)))
		})

	})

})
