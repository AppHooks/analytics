package validators_test

import (
	. "github.com/llun/analytics/validators"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("EmailValidator", func() {

	Describe("#ValidateEmail", func() {

		It("should return true for normal email", func() {
			Expect(ValidateEmail("normal@mail.com")).To(BeTrue())
		})

		It("should return false for no domain email", func() {
			Expect(ValidateEmail("nodomain")).To(BeFalse())
		})

	})

})
