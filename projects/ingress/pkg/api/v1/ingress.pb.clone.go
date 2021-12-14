// Code generated by protoc-gen-ext. DO NOT EDIT.
// source: github.com/solo-io/gloo/projects/ingress/api/v1/ingress.proto

package v1

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"strings"

	"github.com/solo-io/protoc-gen-ext/pkg/clone"
	"google.golang.org/protobuf/proto"

	github_com_golang_protobuf_ptypes_any "github.com/golang/protobuf/ptypes/any"

	github_com_solo_io_solo_kit_pkg_api_v1_resources_core "github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
)

// ensure the imports are used
var (
	_ = errors.New("")
	_ = fmt.Print
	_ = binary.LittleEndian
	_ = bytes.Compare
	_ = strings.Compare
	_ = clone.Cloner(nil)
	_ = proto.Message(nil)
)

// Clone function
func (m *Ingress) Clone() proto.Message {
	var target *Ingress
	if m == nil {
		return target
	}
	target = &Ingress{}

	if h, ok := interface{}(m.GetKubeIngressSpec()).(clone.Cloner); ok {
		target.KubeIngressSpec = h.Clone().(*github_com_golang_protobuf_ptypes_any.Any)
	} else {
		target.KubeIngressSpec = proto.Clone(m.GetKubeIngressSpec()).(*github_com_golang_protobuf_ptypes_any.Any)
	}

	if h, ok := interface{}(m.GetKubeIngressStatus()).(clone.Cloner); ok {
		target.KubeIngressStatus = h.Clone().(*github_com_golang_protobuf_ptypes_any.Any)
	} else {
		target.KubeIngressStatus = proto.Clone(m.GetKubeIngressStatus()).(*github_com_golang_protobuf_ptypes_any.Any)
	}

	if h, ok := interface{}(m.GetMetadata()).(clone.Cloner); ok {
		target.Metadata = h.Clone().(*github_com_solo_io_solo_kit_pkg_api_v1_resources_core.Metadata)
	} else {
		target.Metadata = proto.Clone(m.GetMetadata()).(*github_com_solo_io_solo_kit_pkg_api_v1_resources_core.Metadata)
	}

	return target
}