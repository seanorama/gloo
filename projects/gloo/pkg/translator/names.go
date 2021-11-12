package translator

import (
	"fmt"
	v1 "github.com/solo-io/gloo/projects/gloo/pkg/api/v1"
)

func routeConfigName(listener *v1.Listener) string {
	return listener.GetName() + "-routes"
}

func routeConfigNameWithIndex(listener *v1.Listener, index int) string {
	if hybridListener := listener.GetHybridListener(); hybridListener != nil && index > 0 {
		return fmt.Sprintf("%s-%s-routes", listener.GetName(), hybridListener.GetMatchedListeners()[index].GetMatcher().String())
	}
	return routeConfigName(listener)
}
