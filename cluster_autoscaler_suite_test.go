package main_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestClusterAutoscaler(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ClusterAutoscaler Suite")
}
