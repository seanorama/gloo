package e2e_test

import (
	"context"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	gatewayv1 "github.com/solo-io/gloo/projects/gateway/pkg/api/v1"
	gatewaydefaults "github.com/solo-io/gloo/projects/gateway/pkg/defaults"
	v1 "github.com/solo-io/gloo/projects/gloo/pkg/api/v1"
	"github.com/solo-io/gloo/projects/gloo/pkg/defaults"
	"github.com/solo-io/gloo/test/v1helpers"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"

	"github.com/solo-io/gloo/test/services"
)

var _ = Describe("xDS Sanitizers", func() {

	var (
		ctx           context.Context
		cancel        context.CancelFunc
	)

	BeforeEach(func() {
		ctx, cancel = context.WithCancel(context.Background())

		defaults.HttpPort = services.NextBindPort()
		defaults.HttpsPort = services.NextBindPort()

	})

	AfterEach(func() {
		cancel()
	})

	Context("Route Replacing Sanitizer", func() {

		var (
			err error
			testClients   services.TestClients
			envoyInstance *services.EnvoyInstance

			proxy *v1.Proxy

			gateway *gatewayv1.Gateway
			testUpstream *v1helpers.TestUpstream

			invalidConfigPolicy *v1.GlooOptions_InvalidConfigPolicy
		)

		JustBeforeEach(func() {
			runOptions := &services.RunOptions{
				// Each test can define invalid config policy to inject into Gloo
				Settings: &v1.Settings{
					Gloo: &v1.GlooOptions{
						InvalidConfigPolicy: invalidConfigPolicy,
					},
				},
				NsToWrite: defaults.GlooSystem,
				NsToWatch: []string{
					"default",
					defaults.GlooSystem,
				},
				WhatToRun: services.What{
					DisableFds: true,
					DisableUds: true,
				},
			}
			testClients = services.RunGlooGatewayUdsFds(ctx, runOptions)

			envoyInstance, err = envoyFactory.NewEnvoyInstance()
			Expect(err).NotTo(HaveOccurred())
			err = envoyInstance.RunWithRole(defaults.GlooSystem+"~"+gatewaydefaults.GatewayProxyName, testClients.GlooPort)
			Expect(err).NotTo(HaveOccurred())


			testUpstream = v1helpers.NewTestHttpUpstream(ctx, envoyInstance.LocalAddr())


			gateway = gatewaydefaults.DefaultGateway(defaults.GlooSystem)

			_, err := testClients.GatewayClient.Write(gateway, clients.WriteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())

			proxy = createProxyWithValidAndInvalidRoute()

			_, err = testClients.ProxyClient.Write(proxy, clients.WriteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())


		})

		AfterEach(func() {
			err := testClients.GatewayClient.Delete(gateway.GetMetadata().GetNamespace(), gateway.GetMetadata().GetName(), clients.DeleteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())

			err = testClients.ProxyClient.Delete(proxy.GetMetadata().GetNamespace(), proxy.GetMetadata().GetName(), clients.DeleteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())

			envoyInstance.Clean()
		})


		When("InvalidConfigPolicy.ReplaceInvalidRoutes is True", func() {

			BeforeEach(func() {
				invalidConfigPolicy = &v1.GlooOptions_InvalidConfigPolicy{
					ReplaceInvalidRoutes:     false,
					InvalidRouteResponseCode: 418,
					InvalidRouteResponseBody: "",
				}
			})

			It("replaces invalid routes", func() {
				v1helpers.TestUpstreamReachable(defaults.HttpPort, testUpstream, nil)
			})
		})

		When("InvalidConfigPolicy.ReplaceInvalidRoutes is False", func() {

			BeforeEach(func() {
				invalidConfigPolicy = &v1.GlooOptions_InvalidConfigPolicy{
					ReplaceInvalidRoutes:     true,
					InvalidRouteResponseCode: 418,
					InvalidRouteResponseBody: "",
				}
			})

			It("does not replace invalid routes", func() {

			})
		})

		When("InvalidConfigPolicy.ReplaceInvalidRoutes is not defined", func() {
			// This test validates that our e2e tests by default have invalid route replacement configured

			BeforeEach(func() {
				invalidConfigPolicy = nil
			})

			It("replaces invalid routes", func() {

			})
		})
	})

})


func createProxyWithValidAndInvalidRoute() *v1.Proxy {
	return nil
}