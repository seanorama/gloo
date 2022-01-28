package securityscanutils

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"k8s.io/apimachinery/pkg/util/version"
)

const GlooProjectName = "gloo"
const SoloProjectsProjectName = "solo-projects"

// WriteSecurityScanReportForProject works by performing the following steps:
// 	1. Given a list of tags (ie v1.8.4) and a project (ie Gloo)
//  2. For each tag:
// 		A - Determine what images were published for that tag
//		B - Pull down the report for that image/tag combination
//		C - Aggregate those details into a consumable format
//		D - Write those reports to a file that our docs templates can render
func WriteSecurityScanReportForProject(project string, tags []string) error {
	if project != GlooProjectName && project != SoloProjectsProjectName {
		panic("Only supported for gloo and solo-projects")
	}

	// We assume that tags are sorted in descending order
	latestTag := tags[0]
	prevMinorVersion, _ := version.ParseSemantic(latestTag)

	reportWriter := GetReportWriter(project, prevMinorVersion)

	for ix, tag := range tags {
		taggedVersion, err := version.ParseSemantic(tag)
		if err != nil {
			return err
		}

		// To make the docs clearer, we have some formatting to distinguish the most recent tag
		// for each minor version
		isLatestTagForMinorVersion := ix == 0 || taggedVersion.Minor() != prevMinorVersion.Minor()
		if isLatestTagForMinorVersion {
			prevMinorVersion = taggedVersion
		}

		// The report builder handles the formatting logic for combining a report for a single image
		// into a report for a set of images within a tag
		versionedReportBuilder := GetVersionedReportBuilder(project, taggedVersion, isLatestTagForMinorVersion)

		// Aggregate all image reports into a single report
		if err := addProjectImagesToReport(project, versionedReportBuilder); err != nil {
			return err
		}

		report := versionedReportBuilder.Build()

		if err := reportWriter.Write(taggedVersion, report); err != nil {
			return err
		}
	}

	return reportWriter.Flush()
}

func addProjectImagesToReport(project string, reportBuilder *VersionedReportBuilder) error {
	if project == GlooProjectName {
		return addGlooImagesToReport(reportBuilder)
	}

	if project == SoloProjectsProjectName {
		return addGlooEImagesToReport(reportBuilder)
	}

	return nil
}

func addGlooImagesToReport(reportBuilder *VersionedReportBuilder) error {
	publishedReportUrlTemplate := "https://storage.googleapis.com/solo-gloo-security-scans/gloo/%s/%s_cve_report.docgen"
	for _, image := range OpenSourceImages() {
		url := fmt.Sprintf(publishedReportUrlTemplate, reportBuilder.Version.String(), image)
		imageReport, err := GetSecurityScanReport(url)
		if err != nil {
			return err
		}
		reportBuilder.AddImageReport(image, imageReport)
	}
	return nil
}

func addGlooEImagesToReport(reportBuilder *VersionedReportBuilder) error {
	publishedReportUrlTemplate := "https://storage.googleapis.com/solo-gloo-security-scans/solo-projects/%s/%s_cve_report.docgen"
	hasFedVersion, _ := reportBuilder.Version.Compare("1.7.0")

	for _, image := range EnterpriseImages(hasFedVersion < 0) {
		url := fmt.Sprintf(publishedReportUrlTemplate, reportBuilder.Version.String(), image)
		imageReport, err := GetSecurityScanReport(url)
		if err != nil {
			return err
		}
		reportBuilder.AddImageReport(image, imageReport)
	}
	return nil
}

// List of images included in gloo edge open source
func OpenSourceImages() []string {
	return []string{"access-logger", "certgen", "discovery", "gateway", "gloo", "gloo-envoy-wrapper", "ingress", "sds"}
}

// List of images only included in gloo edge enterprise
// In 1.7, we replaced the grpcserver images with gloo-fed images.
// For images before 1.7, set before17 to true.
func EnterpriseImages(before17 bool) []string {
	extraImages := []string{"gloo-fed", "gloo-fed-apiserver", "gloo-fed-apiserver-envoy", "gloo-federation-console", "gloo-fed-rbac-validating-webhook"}
	if before17 {
		extraImages = []string{"grpcserver-ui", "grpcserver-envoy", "grpcserver-ee"}
	}
	return append([]string{"rate-limit-ee", "gloo-ee", "gloo-ee-envoy-wrapper", "observability-ee", "extauth-ee"}, extraImages...)
}

func GetSecurityScanReport(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}

	var report string
	if resp.StatusCode == http.StatusOK {
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		report = string(bodyBytes)
	} else if resp.StatusCode == http.StatusNotFound {
		// Older releases may be missing scan results
		report = "No scan found\n"
	}
	resp.Body.Close()

	return report, nil
}
