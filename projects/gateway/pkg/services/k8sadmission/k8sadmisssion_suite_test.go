package k8sadmission_test

import (
	"os"
	"testing"

	"github.com/solo-io/gloo/projects/gloo/pkg/defaults"
	"github.com/solo-io/solo-kit/pkg/utils/statusutils"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestK8sAdmission(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "K8sAdmission Suite")
}

var _ = BeforeSuite(func() {
	err := os.Setenv(statusutils.PodNamespaceEnvName, defaults.GlooSystem)
	Expect(err).NotTo(HaveOccurred())
})

var _ = AfterSuite(func() {
	err := os.Unsetenv(statusutils.PodNamespaceEnvName)
	Expect(err).NotTo(HaveOccurred())
})
