package reconciler

import (
	gloov1 "github.com/solo-io/gloo/projects/gloo/pkg/api/v1"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	resources "github.com/solo-io/solo-kit/pkg/api/v1/resources"
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

type memoryProxyClient struct {
	proxies    map[string]*gloov1.Proxy
	baseClient memoryResourceClient
}

func (s *memoryProxyClient) Watch(namespace string, opts clients.WatchOpts) (<-chan gloov1.ProxyList, <-chan error, error) {
	panic("implement me")
}

func NewMemoryProxyClient() MemoryProxyClient {
	proxies := make(map[string]*gloov1.Proxy)
	c := memoryProxyClient{
		proxies: proxies,
	}
	baseClient := memoryResourceClient{
		proxyClient: &c,
	}
	c.baseClient = baseClient
	return &c
}
func (s *memoryProxyClient) BaseClient() clients.ResourceClient {
	return s.baseClient
}

func (s *memoryProxyClient) Register() error {
	return nil
}

func (s *memoryProxyClient) Write(resource *gloov1.Proxy, opts clients.WriteOpts) (*gloov1.Proxy, error) {
	//TODO respect opts
	name := resource.GetMetadata().GetName()
	s.proxies[name] = resource
	return resource, nil
}

func (s *memoryProxyClient) Read(namespace, name string, opts clients.ReadOpts) (*gloov1.Proxy, error) {
	//TODO error handling if necessary
	return s.proxies[name], nil
}

func (s *memoryProxyClient) Delete(namespace, name string, opts clients.DeleteOpts) error {
	//TODO error handling if necessary, handle DeleteOpts
	delete(s.proxies, name)
	return nil
}
func (s *memoryProxyClient) List(namespace string, opts clients.ListOpts) (gloov1.ProxyList, error) {
	proxyList := make([]*gloov1.Proxy, 0, len(s.proxies))
	for _, proxy := range s.proxies {
		proxyList = append(proxyList, proxy)
	}
	return proxyList, nil
}

type memoryResourceClient struct {
	proxyClient *memoryProxyClient
}

func (m memoryResourceClient) Kind() string {
	return "*v1.Proxy"
}

func (m memoryResourceClient) NewResource() resources.Resource {
	panic("implement me")
}

func (m memoryResourceClient) Register() error {
	return nil
}

func (m memoryResourceClient) Read(namespace, name string, opts clients.ReadOpts) (resources.Resource, error) {
	return m.proxyClient.Read(namespace, name, opts)
}

func (m memoryResourceClient) Write(resource resources.Resource, opts clients.WriteOpts) (resources.Resource, error) {
	return m.proxyClient.Write(resource.(*gloov1.Proxy), opts)
}

func (m memoryResourceClient) Delete(namespace, name string, opts clients.DeleteOpts) error {
	return m.proxyClient.Delete(namespace, name, opts)
}

func (m memoryResourceClient) List(namespace string, opts clients.ListOpts) (resources.ResourceList, error) {
	proxyList, _ := m.proxyClient.List(namespace, opts)
	return convertToResource(proxyList), nil
}

func (m memoryResourceClient) Watch(namespace string, opts clients.WatchOpts) (<-chan resources.ResourceList, <-chan error, error) {
	panic("implement me")
}
func convertToResource(proxies gloov1.ProxyList) resources.ResourceList {
	var resourceList resources.ResourceList
	for _, proxy := range proxies {
		resourceList = append(resourceList, proxy)
	}
	return resourceList
}
