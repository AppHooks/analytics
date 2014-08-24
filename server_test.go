package main_test

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/go-martini/martini"
	"github.com/jinzhu/gorm"
	"github.com/llun/analytics"
	"github.com/llun/analytics/models"

	. "github.com/martini-contrib/sessions"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const (
	REGISTER_URL = "/users/register"
	LOGIN_URL    = "/users/login"

	LOGIN_PAGE         = "/users/login.html"
	USER_REGISTER_PAGE = "/users/register.html"
	SERVICE_LIST_PAGE  = "/services/list.html"
)

func GenerateFixtures() *gorm.DB {
	os.Remove("/tmp/analytics.db")
	log.Println("Remove tmp database.")

	db, _ := gorm.Open("sqlite3", "/tmp/analytics.db")
	db.CreateTable(models.User{})
	user, _ := models.NewUser(&db, "admin@email.com", "password")
	user.Save()
	log.Printf("New user: %+v\n", user)

	return &db
}

func CreatePostFormRequest(url string, data *url.Values) *http.Request {
	req, _ := http.NewRequest("POST", url, strings.NewReader(data.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))
	return req
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
			Expect(res.Header().Get("Location")).To(Equal(LOGIN_PAGE))

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
			Expect(res.Header().Get("Location")).To(Equal(SERVICE_LIST_PAGE))

		})

	})

	Describe("User login", func() {
		It("should redirect to index page when success", func() {
			res := httptest.NewRecorder()
			req := CreatePostFormRequest(LOGIN_URL, &url.Values{
				"email":    []string{"admin@email.com"},
				"password": []string{"password"},
			})

			m.ServeHTTP(res, req)

			Expect(res.Code).To(Equal(http.StatusFound))
			Expect(res.Header().Get("Location")).To(Equal(SERVICE_LIST_PAGE))
		})
	})

	Describe("User register", func() {

		createRegisterData := func(email string, password string, confirm string) *url.Values {
			return &url.Values{
				"email":    []string{email},
				"password": []string{password},
				"confirm":  []string{confirm},
			}
		}

		It("should redirect to index page when success", func() {
			res := httptest.NewRecorder()
			req := CreatePostFormRequest(REGISTER_URL, createRegisterData("user@email.com", "password", "password"))

			m.ServeHTTP(res, req)

			Expect(res.Code).To(Equal(http.StatusFound))
			Expect(res.Header().Get("Location")).To(Equal(SERVICE_LIST_PAGE))
		})

		It("should redirect back to register page when validate email fail", func() {
			res := httptest.NewRecorder()
			req := CreatePostFormRequest(REGISTER_URL, createRegisterData("user", "password", "password"))

			m.ServeHTTP(res, req)

			Expect(res.Code).To(Equal(http.StatusFound))
			Expect(res.Header().Get("Location")).To(Equal("/users/register.html"))
		})

		It("should redirect back to register page when user is already exists", func() {
			res := httptest.NewRecorder()
			req := CreatePostFormRequest(REGISTER_URL, createRegisterData("admin@email.com", "password", "password"))

			m.ServeHTTP(res, req)

			Expect(res.Code).To(Equal(http.StatusFound))
			Expect(res.Header().Get("Location")).To(Equal(USER_REGISTER_PAGE))
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
