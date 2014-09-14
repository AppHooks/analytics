package main_test

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/go-martini/martini"
	"github.com/jinzhu/gorm"
	"github.com/llun/analytics/models"
	"github.com/llun/martini-acerender"

	. "github.com/llun/analytics"
	. "github.com/llun/analytics/services"
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

	db.Model(user).Association("Services").Append(models.NewService(&db, "other1", "mock", map[string]interface{}{
		"key": "sample",
	}))
	db.Model(user).Association("Services").Append(models.NewService(&db, "other2", "mock", map[string]interface{}{
		"key": "sample",
	}))

	return &db, user
}

func CreatePostFormRequest(url string, data *url.Values) *http.Request {
	req, _ := http.NewRequest("POST", url, strings.NewReader(data.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))
	req.RemoteAddr = "127.0.0.1"
	return req
}

func CreateJSONFormRequest(url string, data map[string]interface{}) *http.Request {
	b, _ := json.Marshal(data)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(b))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Content-Length", strconv.Itoa(len(b)))
	req.RemoteAddr = "127.0.0.1"

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
			Analytics(db, m, map[string]Factory{
				"mock": func(name string, configuration map[string]interface{}) Service {
					return &MockService{Name: name}
				},
			})
			m.Use(func(s Session, d acerender.TemplateData) {
				s.Set(SESSION_USER_KEY, "0")
				user := models.GetUserFromId(db, firstUser.Id)
				d.Set(SESSION_USER_KEY, user)
			})
		})

		Context("Services", func() {

			It("should add service to user object", func() {
				res := httptest.NewRecorder()
				req := CreateJSONFormRequest(SERVICE_ADD_URL, map[string]interface{}{
					"name": "Service1",
					"type": "other",
					"config": map[string]interface{}{
						"key": "value",
					},
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
				Expect(len(services)).To(Equal(3))

				Expect(services[2].Name).To(Equal("Service1"))
				Expect(services[2].Type).To(Equal("other"))

				configBytes, _ := json.Marshal(map[string]interface{}{
					"key": "value",
				})
				Expect(services[2].Configuration).To(Equal(string(configBytes)))

			})

			It("should remove service", func() {
				res := httptest.NewRecorder()
				req, _ := http.NewRequest("GET", SERVICE_REMOVE_URL+"/1", nil)

				m.ServeHTTP(res, req)

				Expect(res.Code).To(Equal(http.StatusFound))
				Expect(res.Header().Get("Location")).To(Equal(SERVICE_LIST_PAGE))

				var services []models.Service
				user := models.GetUserFromId(db, firstUser.Id)
				db.Model(user).Related(&services)
				Expect(len(services)).To(Equal(1))
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
				req.RemoteAddr = "127.0.0.1"
				req.Header.Add("Content-Type", "application/json")
				req.Header.Add("Content-Length", strconv.Itoa(len(data)))

				m.ServeHTTP(res, req)

				Expect(res.Code).To(Equal(http.StatusOK))

				body, _ := ioutil.ReadAll(res.Body)
				output := map[string]interface{}{
					"success": true,
					"services": map[string]interface{}{
						"other1": true,
						"other2": true,
					},
				}
				outputBytes, _ := json.Marshal(output)
				Expect(string(body)).To(Equal(string(outputBytes)))
			})

			It("should render list of services that user own", func() {
				res := httptest.NewRecorder()
				req, _ := http.NewRequest("GET", "/services/list.html", nil)

				m.ServeHTTP(res, req)

				bodyBytes, _ := ioutil.ReadAll(res.Body)

				Expect(res.Code).To(Equal(http.StatusOK))

				re := regexp.MustCompile("<td>(\\w+)</td>")
				matches := re.FindAllStringSubmatch(string(bodyBytes), -1)

				founds := []string{}
				for _, list := range matches {
					founds = append(founds, list[1])
				}
				Expect(founds).To(Equal([]string{"other1", "other2"}))
			})

		})

		Context("Update user profile", func() {

			It("should update user profile when current password is match", func() {
				res := httptest.NewRecorder()
				req := CreatePostFormRequest(USER_UPDATE_URL, &url.Values{
					"current":  []string{"password"},
					"email":    []string{"user@newemail.com"},
					"password": []string{"newpassword"},
					"confirm":  []string{"newpassword"},
				})

				m.ServeHTTP(res, req)

				Expect(res.Code).To(Equal(http.StatusFound))
				Expect(res.Header().Get("Location")).To(Equal(USER_PROFILE_PAGE))

				user := models.GetUserFromId(db, 1)
				Expect(user.Email).ToNot(Equal(firstUser.Email))
				Expect(user.Password).ToNot(Equal(firstUser.Password))
				Expect(user.Email).To(Equal("user@newemail.com"))
			})

			It("should not update user profile when current password is not match", func() {
				res := httptest.NewRecorder()
				req := CreatePostFormRequest(USER_UPDATE_URL, &url.Values{
					"current":  []string{"wrongpassword"},
					"email":    []string{"user@newemail.com"},
					"password": []string{"newpassword"},
					"confirm":  []string{"newpassword"},
				})

				m.ServeHTTP(res, req)

				Expect(res.Code).To(Equal(http.StatusFound))
				Expect(res.Header().Get("Location")).To(Equal(USER_PROFILE_PAGE))

				user := models.GetUserFromId(db, 1)
				Expect(user.Email).To(Equal(firstUser.Email))
				Expect(user.Password).To(Equal(firstUser.Password))
			})

			It("should not update password when password and confirm is not match", func() {
				res := httptest.NewRecorder()
				req := CreatePostFormRequest(USER_UPDATE_URL, &url.Values{
					"current":  []string{"password"},
					"email":    []string{"user@newemail.com"},
					"password": []string{"newpassword"},
					"confirm":  []string{"notmatchpassword"},
				})

				m.ServeHTTP(res, req)

				Expect(res.Code).To(Equal(http.StatusFound))
				Expect(res.Header().Get("Location")).To(Equal(USER_PROFILE_PAGE))

				user := models.GetUserFromId(db, 1)
				Expect(user.Email).To(Equal(firstUser.Email))
				Expect(user.Password).To(Equal(firstUser.Password))
			})

		})

	})

})
