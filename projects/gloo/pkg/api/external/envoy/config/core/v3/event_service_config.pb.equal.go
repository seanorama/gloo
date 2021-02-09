// Code generated by protoc-gen-ext. DO NOT EDIT.
// source: github.com/solo-io/gloo/projects/gloo/api/external/envoy/config/core/v3/event_service_config.proto

package v3

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"strings"

	"github.com/golang/protobuf/proto"
	equality "github.com/solo-io/protoc-gen-ext/pkg/equality"
)

// ensure the imports are used
var (
	_ = errors.New("")
	_ = fmt.Print
	_ = binary.LittleEndian
	_ = bytes.Compare
	_ = strings.Compare
	_ = equality.Equalizer(nil)
	_ = proto.Message(nil)
)

// Equal function
func (m *EventServiceConfig) Equal(that interface{}) bool {
	if that == nil {
		return m == nil
	}

	target, ok := that.(*EventServiceConfig)
	if !ok {
		that2, ok := that.(EventServiceConfig)
		if ok {
			target = &that2
		} else {
			return false
		}
	}
	if target == nil {
		return m == nil
	} else if m == nil {
		return false
	}

	switch m.ConfigSourceSpecifier.(type) {

	case *EventServiceConfig_GrpcService:
		if _, ok := target.ConfigSourceSpecifier.(*EventServiceConfig_GrpcService); !ok {
			return false
		}

		if h, ok := interface{}(m.GetGrpcService()).(equality.Equalizer); ok {
			if !h.Equal(target.GetGrpcService()) {
				return false
			}
		} else {
			if !proto.Equal(m.GetGrpcService(), target.GetGrpcService()) {
				return false
			}
		}

	default:
		// m is nil but target is not nil
		if m.ConfigSourceSpecifier != target.ConfigSourceSpecifier {
			return false
		}
	}

	return true
}