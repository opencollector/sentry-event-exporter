//
// Copyright (c) 2019 Open Collector, Inc.
//                    Moriyoshi Koizumi
//
// Permission is hereby granted, free of charge, to any person obtaining a copy of
// this software and associated documentation files (the "Software"), to deal in
// the Software without restriction, including without limitation the rights to
// use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies
// of the Software, and to permit persons to whom the Software is furnished to do
// so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	isatty "github.com/mattn/go-isatty"
	export "github.com/opencollector/sentry-event-exporter"
)

var progName string
var config export.ExporterConfig
var tagRegexp = regexp.MustCompile("^\\w+:")

func init() {
	progName = filepath.Base(os.Args[0])
	flag.StringVar(&config.AuthToken, "authtoken", "", "Sentry authentication token (get it at https://sentry.io/settings/account/api/auth-tokens/ )")
	config.Endpoint = flag.String("endpoint", "", "Sentry endpoint")
	flag.StringVar(&config.Organization, "organization", "", "organization")
	flag.StringVar(&config.Project, "project", "", "project")
	flag.BoolVar(&config.IncludeEvents, "events", false, "include events in the result")
	flag.Parse()
	criteria := os.Args[len(os.Args)-flag.NArg():]
	if len(criteria) > 0 {
		l := 0
		// roughly estimate the resulting string's length
		for _, c := range criteria {
			l += len(c) + 1
		}
		buf := make([]byte, 0, l)
		for i, q := range criteria {
			if i > 0 {
				buf = append(buf, ' ')
			}
			// if the criterion is a tag, only the value part has to be quoted.
			l := tagRegexp.FindStringIndex(q)
			if l != nil {
				buf = append(buf, q[l[0]:l[1]]...)
				q = q[l[1]:]
			}
			// if the criterion contains a whitespace, then enclose it in double quotes.
			if strings.ContainsAny(q, " \t\n\r") {
				q = strings.Replace(q, "\"", "\\\"", -1)
				buf = append(buf, '"')
				buf = append(buf, q...)
				buf = append(buf, '"')
			} else {
				buf = append(buf, q...)
			}
		}
		query := string(buf)
		config.Query = &query
	}
	if config.Endpoint != nil && *config.Endpoint == "" {
		config.Endpoint = nil
	}
	if config.AuthToken == "" {
		config.AuthToken = os.Getenv("SENTRY_AUTHTOKEN")
	}
	if config.AuthToken == "" {
		bail("authtoken is not specified")
	}
	if config.Organization == "" {
		bail("organization is not specified")
	}
	if config.Project == "" {
		bail("project is not specified")
	}
}

func bail(msg string) {
	fmt.Fprintf(os.Stderr, "%s: %s\n", progName, msg)
	os.Exit(1)
}

func bufferIfNotTTY(w io.Writer) io.Writer {
	fdBacked, ok := w.(interface{ Fd() uintptr })
	if ok && (isatty.IsTerminal(fdBacked.Fd()) || isatty.IsCygwinTerminal(fdBacked.Fd())) {
		return w
	}
	return bufio.NewWriter(w)
}

func main() {
	w := bufferIfNotTTY(os.Stdout)
	defer func() {
		flushable, ok := w.(interface{ Flush() error })
		if ok {
			flushable.Flush()
		}
	}()
	exporter, err := export.NewExporter(config, export.NewExcelCSVRenderer(w))
	if err != nil {
		bail(err.Error())
	}
	err = exporter.Export()
	if err != nil {
		bail(err.Error())
	}
}
