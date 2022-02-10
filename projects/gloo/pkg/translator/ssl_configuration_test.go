package translator_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v1 "github.com/solo-io/gloo/projects/gloo/pkg/api/v1"
	"github.com/solo-io/gloo/projects/gloo/pkg/translator"
)

var _ = Describe("MergeSslConfig", func() {

	It("merges top-level SslConfig fields", func() {

		dst := &v1.SslConfig{
			SniDomains: []string{"dst"},
		}

		src := &v1.SslConfig{
			SniDomains:    []string{"src"},
			AlpnProtocols: []string{"src"},
		}

		expected := &v1.SslConfig{
			SniDomains: []string{"dst"},
			// Since this field is not defined on the dst, it should be overridden by the src
			AlpnProtocols: []string{"src"},
		}

		actual := translator.MergeSslConfig(dst, src)
		Expect(actual).To(Equal(expected))
	})

})
