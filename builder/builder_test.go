package main

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"testing"
)

func TestBuilder(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Builder Specs")
}

var _ = Describe("Build", func() {
	var ()

	BeforeEach(func() {
	})

	Context("when", func() {
		It("", func() {
		})
	})
})
