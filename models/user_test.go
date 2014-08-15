package models_test

import (
	. "github.com/llun/analytics/models"
	_ "github.com/mattn/go-sqlite3"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Models/User", func() {

	Describe("User", func() {

		Context("#NewUser", func() {

			It("should hash password", func() {
				user := NewUser(nil, "newuser", "password")
				Expect(user.Hash).ToNot(Equal("password"))
			})

		})

		Context("#Authenticate", func() {

			var user *User

			BeforeEach(func() {
				user = NewUser(nil, "user", "password")
			})

			It("should authenticate success with correct password", func() {
				Expect(user.Authenticate("password")).To(BeTrue())
			})

			It("should authenticate fail with wrong password", func() {
				Expect(user.Authenticate("wrong_password")).To(BeFalse())
			})

		})

	})

})
