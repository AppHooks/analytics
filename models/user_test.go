package models_test

import (
	. "github.com/llun/analytics/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Models/User", func() {

	Describe("User", func() {

		Context("#Authenticate", func() {

			var user User

			BeforeEach(func() {
				user = User{
					Username: "user",
					Hash:     "5baa61e4c9b93f3f0682250b6cf8331b7ee68fd8",
				}
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
