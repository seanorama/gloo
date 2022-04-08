package setup
import (
	"context"
	. "github.com/onsi/ginkgo"
	v1 "github.com/solo-io/gloo/projects/gloo/pkg/api/v1"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/factory"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/memory"
)
var _ = Describe("Clean up proxies", func() {
	var(
		proxyClient v1.ProxyClient
		ctx         context.Context
	)
	BeforeEach(func() {
		resourceClientFactory := &factory.MemoryResourceClientFactory{
			Cache: memory.NewInMemoryResourceCache(),
		}

		proxyClient, _ = v1.NewProxyClient(ctx, resourceClientFactory)
	})

})