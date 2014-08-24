package validators_test

import (
	. "github.com/llun/analytics/validators"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("EmailValidator", func() {

	Describe("#IsEmail", func() {

		It("should return true for normal email", func() {
			Expect(IsEmail("normal@mail.com")).To(BeTrue())
		})

		It("should return false for no domain email", func() {
			Expect(IsEmail("nodomain")).To(BeFalse())
		})

	})

})
