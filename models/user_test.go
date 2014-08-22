package models_test

import (
	"errors"

	. "github.com/llun/analytics/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Models/User", func() {

	Describe("#NewUser", func() {

		It("should hash password and generate random key", func() {
			user, _ := NewUser(nil, "email@address.com", "password")
			Expect(user.Password).ToNot(Equal("password"))
			Expect(user.Key).NotTo(BeEmpty())
		})

		Context("Validate Email", func() {

			It("should return error when email is not email", func() {
				user, err := NewUser(nil, "email", "password")
				Expect(user).To(BeNil())
				Expect(err).ToNot(BeNil())
				Expect(err).To(Equal(errors.New("Invalid Email")))
			})

			It("should return user when email is correct", func() {
				user, err := NewUser(nil, "email@admin.com", "password")
				Expect(user).ToNot(BeNil())
				Expect(err).To(BeNil())
			})

		})

	})

	Describe("User", func() {

		Context("#Authenticate", func() {

			var user *User

			BeforeEach(func() {
				user, _ = NewUser(nil, "email@address.com", "password")
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
				user, _ := NewUser(nil, "email@address.com", "password")
				Expect(user.ToMap()).To(Equal(map[string]interface{}{
					"email": "email@address.com",
				}))
			})

		})

	})

})
