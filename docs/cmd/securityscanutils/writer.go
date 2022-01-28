package securityscanutils

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"

	"github.com/hashicorp/go-multierror"
	"k8s.io/apimachinery/pkg/util/version"
)

const docsDir = "content/static/content"
const glooFilename = "gloo-security-scan.docgen"
const soloProjectsFilename = "glooe-security-scan.docgen"

func GetReportWriter(project string, latestVersion *version.Version) ReportWriter {
	filename := glooFilename
	if project == SoloProjectsProjectName {
		filename = soloProjectsFilename
	}

	return NewPartitionedFileReportWriter(docsDir, filename, latestVersion)
}

type ReportWriter interface {
	Write(version *version.Version, report string) error
	Flush() error
}

type FileReportWriter struct {
	dir      string
	filename string

	file       *os.File
	fileBuffer *bufio.Writer
}

func NewFileReportWriter(dir, filename string) *FileReportWriter {
	return &FileReportWriter{
		dir:      dir,
		filename: filename,
	}
}

func (f *FileReportWriter) Write(version *version.Version, report string) error {
	if f.fileBuffer == nil {
		// Create the file and fileBuffer once
		if err := os.MkdirAll(f.dir, os.ModePerm); err != nil {
			return err
		}
		outputFile, err := os.Create(filepath.Join(f.dir, f.filename))
		if err != nil {
			return err
		}

		f.file = outputFile
		f.fileBuffer = bufio.NewWriter(outputFile)
	}

	_, err := f.fileBuffer.WriteString(report)
	return err
}

func (f *FileReportWriter) Flush() error {
	defer f.file.Close()

	// flush buffered data to the file
	return f.fileBuffer.Flush()
}

// A PartitionedFileReportWriter writes reports to separate files, grouped by minor version
// We needed this type of ReportWriter since aggregating reports in a single file caused
// our docs to not load.
// NOTE: We build our docs for a number of LTS branches. The templates for those branches still
// look for the old aggregate file name. Therefore, to maintain backwards compatibility with those
// versions, we also write some of the reports to the old file format
type PartitionedFileReportWriter struct {
	dir           string
	filename      string
	latestVersion *version.Version

	writers         map[uint]*FileReportWriter
	aggregateWriter *FileReportWriter
}

func NewPartitionedFileReportWriter(dir string, filename string, latestVersion *version.Version) *PartitionedFileReportWriter {
	return &PartitionedFileReportWriter{
		dir:             dir,
		filename:        filename,
		latestVersion:   latestVersion,
		writers:         make(map[uint]*FileReportWriter),
		aggregateWriter: NewFileReportWriter(dir, filename),
	}
}

func (p *PartitionedFileReportWriter) Write(version *version.Version, report string) error {
	// Choose the appropriate partitioned writer
	w, ok := p.writers[version.Minor()]
	if !ok {
		// Create the partitioned writer if necessary
		w = NewFileReportWriter(fmt.Sprintf("%s/%d", p.dir, version.Minor()), p.filename)
		p.writers[version.Minor()] = w
	}

	if err := w.Write(version, report); err != nil {
		return err
	}

	if p.latestVersion.Minor()-version.Minor() <= 2 {
		// Only include latest minor version and one before that
		// The purpose of this writer is to support older versions of the docs that fail to
		// render when the file is too large.
		return p.aggregateWriter.Write(version, report)
	}

	return nil
}

func (p *PartitionedFileReportWriter) Flush() error {
	var multiErr *multierror.Error

	// Flush each of the writers
	for _, w := range p.writers {
		multiErr = multierror.Append(multiErr, w.Flush())
	}

	// Flush the aggregate writer
	multiErr = multierror.Append(multiErr, p.aggregateWriter.Flush())

	return multiErr.ErrorOrNil()
}
