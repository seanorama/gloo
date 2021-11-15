package translator

import (
	"fmt"
	v1 "github.com/solo-io/gloo/projects/gloo/pkg/api/v1"
)

func routeConfigName(listener *v1.Listener) string {
	return listener.GetName() + "-routes"
}

func routeConfigNameWithMatchedListener(listener *v1.Listener, matchedListener *v1.MatchedListener) string {
	if matchedListener != nil {
		return fmt.Sprintf("%s-%s-routes", listener.GetName(), matchedListener.GetMatcher().String())
	}
	return routeConfigName(listener)
}
