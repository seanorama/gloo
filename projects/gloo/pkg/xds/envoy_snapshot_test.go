package xds_test

import (
	"io/ioutil"
	"syscall"

	envoy_config_cluster_v3 "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	envoy_config_endpoint_v3 "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	envoy_config_listener_v3 "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	envoy_config_route_v3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/gloo/projects/gloo/pkg/xds"
	"github.com/solo-io/solo-kit/pkg/api/v1/control-plane/cache"
	"github.com/solo-io/solo-kit/pkg/api/v1/control-plane/resource"
	"github.com/solo-io/solo-kit/test/matchers"
)

var _ = Describe("EnvoySnapshot", func() {

	It("clones correctly", func() {

		toBeCloned := xds.NewSnapshot("1234",
			[]cache.Resource{resource.NewEnvoyResource(&envoy_config_endpoint_v3.ClusterLoadAssignment{})},
			[]cache.Resource{resource.NewEnvoyResource(&envoy_config_cluster_v3.Cluster{})},
			[]cache.Resource{resource.NewEnvoyResource(&envoy_config_route_v3.Route{})},
			[]cache.Resource{resource.NewEnvoyResource(&envoy_config_listener_v3.Listener{})},
		)

		// Create an identical struct which is guaranteed not to have been touched to compare against
		untouched := xds.NewSnapshot("1234",
			[]cache.Resource{resource.NewEnvoyResource(&envoy_config_endpoint_v3.ClusterLoadAssignment{})},
			[]cache.Resource{resource.NewEnvoyResource(&envoy_config_cluster_v3.Cluster{})},
			[]cache.Resource{resource.NewEnvoyResource(&envoy_config_route_v3.Route{})},
			[]cache.Resource{resource.NewEnvoyResource(&envoy_config_listener_v3.Listener{})},
		)

		clone := toBeCloned.Clone()

		// Verify that original snapshot and clone are identical
		Expect(toBeCloned.Equal(clone.(*xds.EnvoySnapshot))).To(BeTrue())
		Expect(untouched.Equal(clone.(*xds.EnvoySnapshot))).To(BeTrue())

		// Mutate the clone
		clone.GetResources(
			resource.EndpointTypeV3,
		).Items[""].(*resource.EnvoyResource).ResourceProto().(*envoy_config_endpoint_v3.ClusterLoadAssignment).ClusterName = "new_endpoint"

		// Verify that original snapshot was not mutated
		Expect(toBeCloned.Equal(clone.(*xds.EnvoySnapshot))).NotTo(BeTrue())
		Expect(toBeCloned.Equal(untouched)).To(BeTrue())
	})

	FIt("writes to persisted file and restores from file", func() {

		f, err := ioutil.TempFile("", "snapshotcache")
		Expect(err).NotTo(HaveOccurred())
		defer syscall.Unlink(f.Name())

		empty := xds.NewSnapshot("1234", nil, nil, nil, nil)
		registeredTypes := map[string]cache.Snapshot{empty.GetTypeUrl(): empty}

		c := cache.NewSnapshotCacheFromBackup(true, &xds.ProxyKeyHasher{}, nil, f.Name(), registeredTypes)
		key := "test"

		_, err = c.GetSnapshot(key)
		Expect(err).To(MatchError("no snapshot found for node test"))

		snapshot := xds.NewSnapshot("1234",
			[]cache.Resource{resource.NewEnvoyResource(&envoy_config_endpoint_v3.ClusterLoadAssignment{
				ClusterName: "endpointClusterName",
			})},
			[]cache.Resource{resource.NewEnvoyResource(&envoy_config_cluster_v3.Cluster{
				Name: "clusterName",
			})},
			[]cache.Resource{resource.NewEnvoyResource(&envoy_config_route_v3.RouteConfiguration{ // TODO(kdorosh) change other erroneous uses of this
				Name: "routeName",
			})},
			[]cache.Resource{resource.NewEnvoyResource(&envoy_config_listener_v3.Listener{
				Name: "listenerName",
			})},
		)
		err = c.SetSnapshot(key, snapshot) // should write to temp file too
		Expect(err).ToNot(HaveOccurred())

		// simulate restart, need to create new cache
		restoredCache := cache.NewSnapshotCacheFromBackup(true, &xds.ProxyKeyHasher{}, nil, f.Name(), registeredTypes)

		snap, err := restoredCache.GetSnapshot(key)
		Expect(err).ToNot(HaveOccurred())

		rs := snap.GetResources(resource.ClusterTypeV3)
		Expect(rs.Items).To(HaveLen(1))
		Expect(snapshot.Clusters.Items).To(HaveLen(1))
		Expect(rs.Items["clusterName"].ResourceProto()).To(matchers.MatchProto(snapshot.Clusters.Items["clusterName"].ResourceProto()))
		Expect(rs.Version).To(Equal(snapshot.Clusters.Version))

		rs = snap.GetResources(resource.EndpointTypeV3)
		Expect(rs.Items).To(HaveLen(1))
		Expect(snapshot.Endpoints.Items).To(HaveLen(1))
		Expect(rs.Items["endpointClusterName"].ResourceProto()).To(matchers.MatchProto(snapshot.Endpoints.Items["endpointClusterName"].ResourceProto()))
		Expect(rs.Version).To(Equal(snapshot.Endpoints.Version))

		rs = snap.GetResources(resource.RouteTypeV3)
		Expect(rs.Items).To(HaveLen(1))
		Expect(snapshot.Routes.Items).To(HaveLen(1))
		Expect(rs.Items["routeName"].ResourceProto()).To(matchers.MatchProto(snapshot.Routes.Items["routeName"].ResourceProto()))
		Expect(rs.Version).To(Equal(snapshot.Routes.Version))

		rs = snap.GetResources(resource.ListenerTypeV3)
		Expect(rs.Items).To(HaveLen(1))
		Expect(snapshot.Listeners.Items).To(HaveLen(1))
		Expect(rs.Items["listenerName"].ResourceProto()).To(matchers.MatchProto(snapshot.Listeners.Items["listenerName"].ResourceProto()))
		Expect(rs.Version).To(Equal(snapshot.Listeners.Version))
	})
})
