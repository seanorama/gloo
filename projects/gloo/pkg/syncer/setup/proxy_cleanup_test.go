package setup
import (
	"context"
	"github.com/golang/protobuf/ptypes/wrappers"
	. "github.com/onsi/ginkgo"
	v1 "github.com/solo-io/gloo/projects/gloo/pkg/api/v1"
	"github.com/solo-io/gloo/projects/gloo/pkg/defaults"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/factory"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/memory"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
)
var _ = Describe("Clean up proxies", func() {
	var(
		proxyClient v1.ProxyClient
		ctx         context.Context
		settings    *v1.Settings
		managedProxyLabels map[string]string
		unmanagedProxyLabels map[string]string
	)
	BeforeEach(func() {
		resourceClientFactory := &factory.MemoryResourceClientFactory{
			Cache: memory.NewInMemoryResourceCache(),
		}

		proxyClient, _ = v1.NewProxyClient(ctx, resourceClientFactory)
		settings = &v1.Settings{
			ConfigSource: &v1.Settings_KubernetesConfigSource{
				KubernetesConfigSource: &v1.Settings_KubernetesCrds{},
			},
			Gateway: &v1.GatewayOptions{
				EnableGatewayController: &wrappers.BoolValue{Value: true},
				PersistProxySpec: &wrappers.BoolValue{Value: false},
			},
		}
	    managedProxyLabels = map[string]string {
		  "created_by": "gloo-gateway",
	    }
	    unmanagedProxyLabels = map[string] string {
	    	"created_by" : "other-controller",
		}
	})
	It("Deletes proxies with the gateway label", func() {
		gatewayProxy := &v1.Proxy{
			Metadata: &core.Metadata{
				Name:            "test-proxy",
				Namespace:       defaults.GlooSystem,
				Labels:          managedProxyLabels,
			},
		}
		otherProxy := &v1.Proxy {
			Metadata: &core.Metadata{
				Name:            "test-proxy",
				Namespace:       defaults.GlooSystem,
				Labels:          nil,
			},
		}
		if settings.Gateway.PersistProxySpec.Value {
			proxyClient.Write(otherProxy, clients.WriteOpts{})
			proxyClient.Write(gatewayProxy, clients.WriteOpts{})
			deleteUnusedProxies(ctx, defaults.GlooSystem, proxyClient)
		}

	})
	It("Does not delete proxies when persisting proxies is enabled", func() {

	})
})