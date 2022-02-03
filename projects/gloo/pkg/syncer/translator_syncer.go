package syncer

import (
	"context"

	"github.com/rotisserie/eris"
	gatewayv1 "github.com/solo-io/gloo/projects/gateway/pkg/api/v1"
	"github.com/solo-io/gloo/projects/gateway/pkg/reconciler"
	"github.com/solo-io/gloo/projects/gateway/pkg/utils"
	"github.com/solo-io/gloo/projects/gateway/pkg/utils/metrics"
	"github.com/solo-io/gloo/projects/gloo/pkg/api/v1/enterprise/options/ratelimit"
	"github.com/solo-io/gloo/projects/gloo/pkg/syncer/sanitizer"
	"github.com/solo-io/go-utils/contextutils"

	"github.com/hashicorp/go-multierror"
	gwsyncer "github.com/solo-io/gloo/projects/gateway/pkg/syncer"
	gwtranslator "github.com/solo-io/gloo/projects/gateway/pkg/translator"
	v1 "github.com/solo-io/gloo/projects/gloo/pkg/api/v1"
	v1snap "github.com/solo-io/gloo/projects/gloo/pkg/api/v1/gloosnapshot"
	"github.com/solo-io/gloo/projects/gloo/pkg/translator"
	"github.com/solo-io/gloo/projects/gloo/pkg/xds"
	envoycache "github.com/solo-io/solo-kit/pkg/api/v1/control-plane/cache"
	"github.com/solo-io/solo-kit/pkg/api/v2/reporter"
)

type translatorSyncer struct {
	translator translator.Translator
	sanitizer  sanitizer.XdsSanitizer
	xdsCache   envoycache.SnapshotCache
	xdsHasher  *xds.ProxyKeyHasher
	reporter   reporter.StatusReporter
	// used for debugging purposes only
	latestSnap *v1snap.ApiSnapshot
	extensions []TranslatorSyncerExtension
	// used to track which envoy node IDs exist without belonging to a proxy
	extensionKeys     map[string]struct{}
	settings          *v1.Settings
	statusMetrics     metrics.ConfigStatusMetrics
	gatewayTranslator gwtranslator.Translator
	gatewaySyncer     *gwsyncer.TranslatorSyncer
}

type TranslatorSyncerExtensionParams struct {
	RateLimitServiceSettings ratelimit.ServiceSettings
}

type TranslatorSyncerExtensionFactory func(context.Context, TranslatorSyncerExtensionParams) (TranslatorSyncerExtension, error)

type UpgradeableTranslatorSyncerExtension interface {
	ExtensionName() string
	IsUpgrade() bool
}

type TranslatorSyncerExtension interface {
	Sync(
		ctx context.Context,
		snap *v1snap.ApiSnapshot,
		settings *v1.Settings,
		xdsCache envoycache.SnapshotCache,
		reports reporter.ResourceReports,
	) (string, error)
}

func NewTranslatorSyncer(
	translator translator.Translator,
	xdsCache envoycache.SnapshotCache,
	xdsHasher *xds.ProxyKeyHasher,
	sanitizer sanitizer.XdsSanitizer,
	reporter reporter.StatusReporter,
	devMode bool,
	extensions []TranslatorSyncerExtension,
	settings *v1.Settings,
	statusMetrics metrics.ConfigStatusMetrics,
	gatewayTranslator gwtranslator.Translator,
	gatewaySyncer *gwsyncer.TranslatorSyncer,
) v1snap.ApiSyncer {
	s := &translatorSyncer{
		translator:        translator,
		xdsCache:          xdsCache,
		xdsHasher:         xdsHasher,
		reporter:          reporter,
		extensions:        extensions,
		sanitizer:         sanitizer,
		settings:          settings,
		statusMetrics:     statusMetrics,
		gatewayTranslator: gatewayTranslator,
		gatewaySyncer:     gatewaySyncer,
	}
	if devMode {
		// TODO(ilackarms): move this somewhere else?
		go func() {
			_ = s.ServeXdsSnapshots()
		}()
	}
	return s
}

func (s *translatorSyncer) Sync(ctx context.Context, snap *v1snap.ApiSnapshot) error {
	logger := contextutils.LoggerFrom(ctx)
	reports := make(reporter.ResourceReports)

	//generate proxies
	// TODO: check whether we are running in gateway mode
	// TODO: only run if there was an update to a gw type
	s.setProxies(ctx, snap, reports)
	var multiErr *multierror.Error
	err := s.syncEnvoy(ctx, snap, reports)
	if err != nil {
		multiErr = multierror.Append(multiErr, err)
	}
	s.extensionKeys = map[string]struct{}{}
	for _, extension := range s.extensions {
		intermediateReports := make(reporter.ResourceReports)
		nodeID, err := extension.Sync(ctx, snap, s.settings, s.xdsCache, intermediateReports)
		if err != nil {
			multiErr = multierror.Append(multiErr, err)
		}
		reports.Merge(intermediateReports)
		s.extensionKeys[nodeID] = struct{}{}
	}

	if err := s.reporter.WriteReports(ctx, reports, nil); err != nil {
		logger.Debugf("Failed writing report for proxies: %v", err)
		multiErr = multierror.Append(multiErr, eris.Wrapf(err, "writing reports"))
	}
	// Update resource status metrics
	for resource, report := range reports {
		status := s.reporter.StatusFromReport(report, nil)
		s.statusMetrics.SetResourceStatus(ctx, resource, status)
	}

	return multiErr.ErrorOrNil()
}

func (s *translatorSyncer) setProxies(ctx context.Context, snap *v1snap.ApiSnapshot, allReports reporter.ResourceReports) {
	gwSnap := gatewayv1.ApiSnapshot{
		VirtualServices:    snap.VirtualServices,
		Gateways:           snap.Gateways,
		RouteTables:        snap.RouteTables,
		RouteOptions:       snap.RouteOptions,
		VirtualHostOptions: snap.VirtualHostOptions,
	}
	logger := contextutils.LoggerFrom(ctx)
	gatewaysByProxy := utils.GatewaysByProxyName(snap.Gateways)

	desiredProxies := make(reconciler.GeneratedProxies)
	// TODO: make this an instance variable or const
	managedProxyLabels := map[string]string{
		"created_by": "gateway",
	}
	for proxyName, gatewayList := range gatewaysByProxy {
		//TODO writeNamespace
		proxy, reports := s.gatewayTranslator.Translate(ctx, proxyName, "gloo-system", &gwSnap, gatewayList)
		if proxy != nil {

			//TODO: I assume we can remove compression if these stay in memory
			// Otherwise, implement shouldCompress
			//if s.shouldCompresss(ctx) {
			//	compress.SetShouldCompressed(proxy)
			//}

			logger.Infof("desired proxy %v", proxy.GetMetadata().Ref())
			proxy.GetMetadata().Labels = managedProxyLabels
			desiredProxies[proxy] = reports

			//TODO: remove this log
			logger.Infof("Generated proxy %s", proxy.String())
		}
	}
	//TODO handle reports
	//TODO stripInvalidListenersAndVirtualHosts - will generate a list of proxies
	finalProxies :=make(v1.ProxyList, len(desiredProxies))
	i := 0
	for proxy, _ := range(desiredProxies) {
		finalProxies[i] = proxy
		i++
	}
	snap.Proxies = finalProxies
}
