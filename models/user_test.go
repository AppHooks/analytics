package models_test

import (
	. "github.com/llun/analytics/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Models/User", func() {

	Context("#NewUser", func() {

		It("should hash password and generate random key", func() {
			user := NewUser(nil, "email@address.com", "password")
			Expect(user.Password).ToNot(Equal("password"))
			Expect(user.Key).NotTo(BeEmpty())
		})

	})

	Describe("User", func() {

		Context("#Authenticate", func() {

			var user *User

			BeforeEach(func() {
				user = NewUser(nil, "email@address.com", "password")
			})

			It("should authenticate success with correct password", func() {
				Expect(user.Authenticate("password")).To(BeTrue())
			})

			It("should authenticate fail with wrong password", func() {
				Expect(user.Authenticate("wrong_password")).To(BeFalse())
			})

		})

		Context("#ToMap", func() {

			It("should export public properties", func() {
				user := NewUser(nil, "email@address.com", "password")
				Expect(user.ToMap()).To(Equal(map[string]interface{}{
					"email": "email@address.com",
				}))
			})

		})

	})

})
