// Copyright 2018 Envoyproxy Authors
//
//   Licensed under the Apache License, Version 2.0 (the "License");
//   you may not use this file except in compliance with the License.
//   You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//   Unless required by applicable law or agreed to in writing, software
//   distributed under the License is distributed on an "AS IS" BASIS,
//   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//   See the License for the specific language governing permissions and
//   limitations under the License.

package xds

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/golang/protobuf/proto"
	"github.com/solo-io/solo-kit/pkg/api/v1/control-plane/cache"
	"github.com/solo-io/solo-kit/pkg/api/v1/control-plane/resource"
)

// Snapshot is an internally consistent snapshot of xDS resources.
// Consistently is important for the convergence as different resource types
// from the snapshot may be delivered to the proxy in arbitrary order.
type EnvoySnapshot struct {
	// Endpoints are items in the EDS V3 response payload.
	Endpoints resource.EnvoyResources

	// Clusters are items in the CDS response payload.
	Clusters resource.EnvoyResources

	// Routes are items in the RDS response payload.
	Routes resource.EnvoyResources

	// Listeners are items in the LDS response payload.
	Listeners resource.EnvoyResources
}

func (s *EnvoySnapshot) Serialize() ([]byte, error) {
	// convert to json then write out
	b, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (s *EnvoySnapshot) Deserialize(bytes []byte) error {
	err := json.Unmarshal(bytes, &s)
	if err != nil {
		return err
	}
	return nil
}

func (s *EnvoySnapshot) GetTypeUrl() string {
	return "envoysnapshot"
}

var _ cache.Snapshot = &EnvoySnapshot{}

// NewSnapshot creates a snapshot from response types and a version.
func NewSnapshot(
	version string,
	endpoints []*resource.EnvoyResource,
	clusters []*resource.EnvoyResource,
	routes []*resource.EnvoyResource,
	listeners []*resource.EnvoyResource,
) *EnvoySnapshot {
	// TODO: Copy resources
	return &EnvoySnapshot{
		Endpoints: resource.NewEnvoyResources(version, endpoints),
		Clusters:  resource.NewEnvoyResources(version, clusters),
		Routes:    resource.NewEnvoyResources(version, routes),
		Listeners: resource.NewEnvoyResources(version, listeners),
	}
}

func NewSnapshotFromResources(
	endpoints resource.EnvoyResources,
	clusters resource.EnvoyResources,
	routes resource.EnvoyResources,
	listeners resource.EnvoyResources,
) cache.Snapshot {
	// TODO: Copy resources and downgrade, maybe maintain hash to not do it too many times (https://github.com/solo-io/gloo/issues/4421)
	return &EnvoySnapshot{
		Endpoints: endpoints,
		Clusters:  clusters,
		Routes:    routes,
		Listeners: listeners,
	}
}

func NewEndpointsSnapshotFromResources(
	endpoints resource.EnvoyResources,
	clusters resource.EnvoyResources,
) cache.Snapshot {
	return &EnvoySnapshot{
		Endpoints: endpoints,
		Clusters:  clusters,
	}
}

// Consistent check verifies that the dependent resources are exactly listed in the
// snapshot:
// - all EDS resources are listed by name in CDS resources
// - all RDS resources are listed by name in LDS resources
//
// Note that clusters and listeners are requested without name references, so
// Envoy will accept the snapshot list of clusters as-is even if it does not match
// all references found in xDS.
func (s *EnvoySnapshot) Consistent() error {
	if s == nil {
		return errors.New("nil snapshot")
	}
	endpoints := resource.GetResourceReferences(s.Clusters.Items)
	if len(endpoints) != len(s.Endpoints.Items) {
		return fmt.Errorf("mismatched endpoint reference and resource lengths: length of %v does not equal length of %v", endpoints, s.Endpoints.Items)
	}
	if err := resource.Superset(endpoints, s.Endpoints.Items); err != nil {
		return err
	}

	routes := resource.GetResourceReferences(s.Listeners.Items)
	if len(routes) != len(s.Routes.Items) {
		return fmt.Errorf("mismatched route reference and resource lengths: length of %v does not equal length of %v", routes, s.Routes.Items)
	}
	return resource.Superset(routes, s.Routes.Items)
}

// GetResources selects snapshot resources by type.
func (s *EnvoySnapshot) GetResources(typ string) cache.Resources {
	if s == nil {
		return cache.Resources{}
	}
	switch typ {
	case resource.EndpointTypeV3:
		return cache.NewResources(s.Endpoints.Version, resource.GetEnvoyResources(s.Endpoints.Items))
	case resource.ClusterTypeV3:
		return cache.NewResources(s.Clusters.Version, resource.GetEnvoyResources(s.Clusters.Items))
	case resource.RouteTypeV3:
		return cache.NewResources(s.Routes.Version, resource.GetEnvoyResources(s.Routes.Items))
	case resource.ListenerTypeV3:
		return cache.NewResources(s.Listeners.Version, resource.GetEnvoyResources(s.Listeners.Items))
	}
	return cache.Resources{}
}

func (s *EnvoySnapshot) Clone() cache.Snapshot {
	snapshotClone := &EnvoySnapshot{}

	snapshotClone.Endpoints = resource.EnvoyResources{
		Version: s.Endpoints.Version,
		Items:   cloneEnvoyResourceItems(s.Endpoints.Items),
	}

	snapshotClone.Clusters = resource.EnvoyResources{
		Version: s.Clusters.Version,
		Items:   cloneEnvoyResourceItems(s.Clusters.Items),
	}

	snapshotClone.Routes = resource.EnvoyResources{
		Version: s.Routes.Version,
		Items:   cloneEnvoyResourceItems(s.Routes.Items),
	}

	snapshotClone.Listeners = resource.EnvoyResources{
		Version: s.Listeners.Version,
		Items:   cloneEnvoyResourceItems(s.Listeners.Items),
	}

	return snapshotClone
}

func cloneEnvoyResourceItems(items map[string]*resource.EnvoyResource) map[string]*resource.EnvoyResource {
	clonedItems := make(map[string]*resource.EnvoyResource, len(items))
	for k, v := range items {
		resProto := v.ResourceProto()
		resClone := proto.Clone(resProto)
		clonedItems[k] = resource.NewEnvoyResource(resClone)
	}
	return clonedItems
}

// Equal checks is 2 snapshots are equal, important since reflect.DeepEqual no longer works with proto4
func (this *EnvoySnapshot) Equal(that *EnvoySnapshot) bool {
	if len(this.Clusters.Items) != len(that.Clusters.Items) || this.Clusters.Version != that.Clusters.Version {
		return false
	}
	for key, thisVal := range this.Clusters.Items {
		thatVal, ok := that.Clusters.Items[key]
		if !ok {
			return false
		}
		if !proto.Equal(thisVal.ResourceProto(), thatVal.ResourceProto()) {
			return false
		}
	}
	if len(this.Endpoints.Items) != len(that.Endpoints.Items) || this.Endpoints.Version != that.Endpoints.Version {
		return false
	}
	for key, thisVal := range this.Endpoints.Items {
		thatVal, ok := that.Endpoints.Items[key]
		if !ok {
			return false
		}
		if !proto.Equal(thisVal.ResourceProto(), thatVal.ResourceProto()) {
			return false
		}
	}
	if len(this.Routes.Items) != len(that.Routes.Items) || this.Routes.Version != that.Routes.Version {
		return false
	}
	for key, thisVal := range this.Routes.Items {
		thatVal, ok := that.Routes.Items[key]
		if !ok {
			return false
		}
		if !proto.Equal(thisVal.ResourceProto(), thatVal.ResourceProto()) {
			return false
		}
	}
	if len(this.Endpoints.Items) != len(that.Endpoints.Items) || this.Endpoints.Version != that.Endpoints.Version {
		return false
	}
	for key, thisVal := range this.Endpoints.Items {
		thatVal, ok := that.Endpoints.Items[key]
		if !ok {
			return false
		}
		if !proto.Equal(thisVal.ResourceProto(), thatVal.ResourceProto()) {
			return false
		}
	}
	return true
}
