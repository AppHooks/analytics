package main_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestAnalytics(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Analytics Suite")
}
