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

		It("should return error when email is invalid", func() {
			user, err := NewUser(nil, "email", "password")
			Expect(user).To(BeNil())
			Expect(err).To(Equal(errors.New("Invalid Email")))
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

		Context("#UpdatePassword", func() {

			var user *User

			BeforeEach(func() {
				user, _ = NewUser(nil, "email@address.com", "password")
			})

			It("should hash password and return true when success", func() {
				result := user.UpdatePassword("newpassword", "newpassword")
				Expect(result).To(BeTrue())
				Expect(user.Password).ToNot(Equal("newpassword"))

				Expect(user.Authenticate("password")).To(BeFalse())
				Expect(user.Authenticate("newpassword")).To(BeTrue())
			})

			It("should not update password when password and confirm is not match", func() {
				result := user.UpdatePassword("newpassword", "wrongpassword")
				Expect(result).To(BeFalse())

				Expect(user.Authenticate("password")).To(BeTrue())
				Expect(user.Authenticate("newpassword")).To(BeFalse())
			})

		})

	})

})
