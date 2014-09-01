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
	"strings"

	"github.com/go-martini/martini"
	"github.com/jinzhu/gorm"
	"github.com/llun/analytics/models"
	"github.com/llun/martini-acerender"

	. "github.com/llun/analytics"
	. "github.com/martini-contrib/sessions"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func GenerateFixtures() (*gorm.DB, *models.User) {
	os.Remove("/tmp/analytics.db")

	db, _ := gorm.Open("sqlite3", "/tmp/analytics.db")
	db.CreateTable(models.Service{})
	db.CreateTable(models.User{})

	user, _ := models.NewUser(&db, "admin@email.com", "password")
	user.Save()

	service := models.NewService(&db, "other", "other", map[string]interface{}{
		"key": "sample",
	})
	db.Model(user).Association("Services").Append(service)

	return &db, user
}

func CreatePostFormRequest(url string, data *url.Values) *http.Request {
	req, _ := http.NewRequest("POST", url, strings.NewReader(data.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))
	return req
}

var _ = Describe("Server", func() {

	var (
		m         *martini.ClassicMartini
		db        *gorm.DB
		firstUser *models.User
	)

	Describe("Not logged in path", func() {
		BeforeEach(func() {
			m = martini.Classic()

			db, firstUser = GenerateFixtures()
			Analytics(db, m, map[string]Factory{})
		})

		Describe("Index", func() {

			It("should redirect to login when user doesn't login yet", func() {

				res := httptest.NewRecorder()
				req, _ := http.NewRequest("GET", "/", nil)

				m.ServeHTTP(res, req)

				Expect(res.Code).To(Equal(http.StatusFound))
				Expect(res.Header().Get(HEADER_LOCATION)).To(Equal(USER_LOGIN_PAGE))

			})

			It("should redirect to list services when user is already login", func() {

				m.Use(func(c martini.Context, session Session) {
					// Emulate user is already logged in.
					session.Set(SESSION_USER_KEY, "user")
					c.Next()
				})
				res := httptest.NewRecorder()
				req, _ := http.NewRequest("GET", "/", nil)

				m.ServeHTTP(res, req)

				Expect(res.Code).To(Equal(http.StatusFound))
				Expect(res.Header().Get(HEADER_LOCATION)).To(Equal(SERVICE_LIST_PAGE))

			})

		})

		Describe("User login", func() {
			It("should redirect to index page when success", func() {
				res := httptest.NewRecorder()
				req := CreatePostFormRequest(USER_LOGIN_URL, &url.Values{
					"email":    []string{"admin@email.com"},
					"password": []string{"password"},
				})

				m.ServeHTTP(res, req)

				Expect(res.Code).To(Equal(http.StatusFound))
				Expect(res.Header().Get(HEADER_LOCATION)).To(Equal(SERVICE_LIST_PAGE))
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
				req := CreatePostFormRequest(USER_REGISTER_URL, createRegisterData("user@email.com", "password", "password"))

				m.ServeHTTP(res, req)

				Expect(res.Code).To(Equal(http.StatusFound))
				Expect(res.Header().Get(HEADER_LOCATION)).To(Equal(SERVICE_LIST_PAGE))
			})

			It("should redirect back to register page when validate email fail", func() {
				res := httptest.NewRecorder()
				req := CreatePostFormRequest(USER_REGISTER_URL, createRegisterData("user", "password", "password"))

				m.ServeHTTP(res, req)

				Expect(res.Code).To(Equal(http.StatusFound))
				Expect(res.Header().Get(HEADER_LOCATION)).To(Equal("/users/register.html"))
			})

			It("should redirect back to register page when user is already exists", func() {
				res := httptest.NewRecorder()
				req := CreatePostFormRequest(USER_REGISTER_URL, createRegisterData("admin@email.com", "password", "password"))

				m.ServeHTTP(res, req)

				Expect(res.Code).To(Equal(http.StatusFound))
				Expect(res.Header().Get(HEADER_LOCATION)).To(Equal(USER_REGISTER_PAGE))
			})

		})
	})

	Describe("Logged in path", func() {
		BeforeEach(func() {
			m = martini.Classic()

			db, firstUser = GenerateFixtures()
			Analytics(db, m, map[string]Factory{})
			m.Use(func(s Session, d acerender.TemplateData) {
				s.Set(SESSION_USER_KEY, "0")
				user := models.GetUserFromId(db, firstUser.Id)
				d.Set(SESSION_USER_KEY, user)
			})
		})

		Context("Services", func() {

			It("should add service to user object", func() {
				res := httptest.NewRecorder()
				req := CreatePostFormRequest(SERVICE_ADD_URL, &url.Values{
					"name": []string{"Service1"},
					"type": []string{"other"},
					"key":  []string{"key"},
				})

				m.ServeHTTP(res, req)

				body, _ := ioutil.ReadAll(res.Body)

				expectedMap := map[string]interface{}{
					"success": true,
				}
				expectedBytes, _ := json.Marshal(expectedMap)

				Expect(res.Code).To(Equal(http.StatusFound))
				Expect(res.Header().Get("Location")).To(Equal(SERVICE_LIST_PAGE))
				Expect(string(body)).To(Equal(string(expectedBytes)))

				var services []models.Service
				user := models.GetUserFromId(db, firstUser.Id)
				db.Model(user).Related(&services)
				Expect(len(services)).To(Equal(2))

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
				req, _ := http.NewRequest("POST", "/services/send/"+firstUser.Key, bytes.NewBufferString(data))
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

})
