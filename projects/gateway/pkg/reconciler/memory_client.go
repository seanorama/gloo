package reconciler

import (
	gloov1 "github.com/solo-io/gloo/projects/gloo/pkg/api/v1"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
)

type MemoryProxyClient interface {
	BaseClient() clients.ResourceClient
	Register() error
	Read(namespace, name string, opts clients.ReadOpts) (*gloov1.Proxy, error)
	Write(resource *gloov1.Proxy, opts clients.WriteOpts) (*gloov1.Proxy, error)
	Delete(namespace, name string, opts clients.DeleteOpts) error
	List(namespace string, opts clients.ListOpts) (gloov1.ProxyList, error)
	gloov1.ProxyWatcher
}
