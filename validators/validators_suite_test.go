package validators_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestValidators(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Validators Suite")
}
