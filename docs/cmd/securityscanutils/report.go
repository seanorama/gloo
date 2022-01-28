package securityscanutils

import (
	"bytes"
	"fmt"

	"k8s.io/apimachinery/pkg/util/version"
)

func GetVersionedReportBuilder(project string, version *version.Version, isMostRecentVersionInMinorVersion bool) *VersionedReportBuilder {
	projectHeader := "Gloo Open Source"
	if project == SoloProjectsProjectName {
		projectHeader = "Gloo Enterprise"
	}

	reportBuilder := &VersionedReportBuilder{
		Version:            version,
		Header:             fmt.Sprintf("<details><summary> Release %s </summary>\n\n", version.String()),
		Trailer:            fmt.Sprintln("</details>"),
		ImageHeaderFormat:  "**" + projectHeader + " %s image**\n\n",
		ImageTrailerFormat: "\n\n",
	}

	if isMostRecentVersionInMinorVersion {
		reportBuilder.Header = fmt.Sprintf("\n***Latest %d.%d.x %s Release: %s***\n\n", version.Major(), version.Minor(), projectHeader, version.String())
		reportBuilder.Trailer = ""
	}

	return reportBuilder
}

type VersionedReportBuilder struct {
	// The version of the report we will be building
	Version *version.Version

	// The header and trailer that wraps the entire report
	Header, Trailer string

	// The header and trailer that wraps each image within the report
	ImageHeaderFormat, ImageTrailerFormat string

	reportBuffer bytes.Buffer
}

func (v *VersionedReportBuilder) AddImageReport(imageName, imageReport string) {
	imageHeader := fmt.Sprintf(v.ImageHeaderFormat, imageName)
	imageTrailer := fmt.Sprintf(v.ImageTrailerFormat, imageName)

	v.reportBuffer.WriteString(imageHeader)
	v.reportBuffer.WriteString(imageReport)
	v.reportBuffer.WriteString(imageTrailer)
}

func (v *VersionedReportBuilder) Build() string {
	return fmt.Sprintf("%s%s%s", v.Header, v.reportBuffer.String(), v.Trailer)
}
