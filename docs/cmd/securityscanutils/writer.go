package securityscanutils

import (
    "github.com/hashicorp/go-multierror"
    "k8s.io/apimachinery/pkg/util/version"
)


const docsDir = "content/static/content"
const filename = "gloo-security-scan.docgen"

func GetReportWriter(latestVersion *version.Version) ReportWriter {

    
    return &FileReportWriter{
        dir: docsDir,
    }
}

type ReportWriter interface {
    Write(version *version.Version, report string) error
    Flush() error
}

type FileReportWriter struct {
    dir string
    
}

func (a *FileReportWriter) Write(version *version.Version, report string) error {
    panic("implement me")
}

func (a *FileReportWriter) Flush() error {
    panic("implement me")
}


// A PartitionedFileReportWriter writes reports to separate files, grouped by minor version
// We needed this type of ReportWriter since aggregating reports in a single file caused
// our docs to not load.
// NOTE: We build our docs for a number of LTS branches. The templates for those branches still
// look for the old aggregate file name. Therefore, to maintain backwards compatibility with those
// versions, we also write some of the reports to the old file format
type PartitionedFileReportWriter struct {
    latestVersion *version.Version

    writers map[uint]*FileReportWriter

    aggregateWriter *FileReportWriter
}

func (p *PartitionedFileReportWriter) Write(version *version.Version, report string) error {
    // Choose the appropriate partitioned writer
    w, ok := p.writers[version.Minor()]
    if ok {
        if err := w.Write(version, report); err != nil {
            return err
        }
    }

    if p.latestVersion.Minor() - version.Minor() > 1 {
        // If the version of the report is not n-1, don't bother aggregating it
        return nil
    }
    return p.aggregateWriter.Write(version, report)
}

func (p *PartitionedFileReportWriter) Flush() error {
    var multiErr *multierror.Error

    for _, w := range p.writers {
         multiErr = multierror.Append(multiErr, w.Flush())
    }
    multiErr = multierror.Append(multiErr, p.aggregateWriter.Flush())

    return multiErr.ErrorOrNil()
}

