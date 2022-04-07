package setup

import (
	"context"
	"errors"

	v1 "github.com/solo-io/gloo/projects/gloo/pkg/api/v1"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
)
func deleteUnusedProxies(ctx context.Context, namespace string, expectedCreatedBy string, proxyClient v1.ProxyClient) error {
	currentProxies, err := proxyClient.List(namespace, clients.ListOpts{Ctx: ctx})
	if err != nil {
		return err
	}
	deleteErrs := make([]error, 0)
	for _, proxy := range currentProxies {
		if val, ok := proxy.GetMetadata().GetLabels()["created-by"]; ok && val == expectedCreatedBy {
			err = proxyClient.Delete(namespace, proxy.GetMetadata().GetName(), clients.DeleteOpts{Ctx: ctx})
			// continue to clean up other proxies
			if err != nil {
				deleteErrs = append(deleteErrs, err)
			}
		}
	}
	// Concatenate error messages from all the failed deletes
	if len(deleteErrs) > 0 {
		allErrs := ""
		for _, err := range deleteErrs {
			allErrs += err.Error()
		}
		return errors.New(allErrs)
	}
	return nil
}
