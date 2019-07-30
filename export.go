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

package export

import (
	sentry "github.com/atlassian/go-sentry-api"
	"github.com/pkg/errors"
	"time"
)

type ExporterConfig struct {
	AuthToken     string
	Endpoint      *string
	Organization  string
	Project       string
	StatsPeriod   *string
	Query         *string
	IncludeEvents bool
}

type ExporterResult struct {
	IssueID         string
	AssignedTo      string
	Count           string
	Culprit         string
	FirstSeen       time.Time
	LastSeen        time.Time
	Level           string
	Logger          string
	Permalink       string
	Project         string
	ShareID         string
	ShortID         string
	Status          string
	Title           string
	IssueType       string
	UserCount       int
	UserReportCount int
	EventID         string
	EventType       string
	Release         string
	Message         string
	EventCreated    time.Time
	EventReceived   time.Time
	Platform        string
	GroupID         string
}

type ExporterResultRenderer interface {
	RenderHeader() error
	RenderPartialResults([]ExporterResult) error
	RenderFooter() error
	Fini()
}

type Exporter struct {
	ExporterConfig
	Renderer ExporterResultRenderer
	Client   *sentry.Client
}

func NewExporter(config ExporterConfig, renderer ExporterResultRenderer) (*Exporter, error) {
	client, err := sentry.NewClient(config.AuthToken, config.Endpoint, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create client")
	}
	return &Exporter{
		ExporterConfig: config,
		Client:         client,
		Renderer:       renderer,
	}, nil
}

func (*ExporterConfig) StringizeInternalUser(u *sentry.InternalUser) string {
	if u == nil {
		return ""
	}
	if u.Username != nil {
		return *u.Username
	} else {
		return "(unknown)"
	}
}

func (*ExporterConfig) StringizeProject(p *sentry.Project) string {
	if p == nil {
		return ""
	} else {
		return p.Name
	}
}

func (*ExporterConfig) StringizeTime(t time.Time) string {
	return t.Format(time.RFC3339Nano)
}

func (*ExporterConfig) StringizeStatus(s *sentry.Status) string {
	if s == nil {
		return ""
	} else {
		return string(*s)
	}
}

func (*ExporterConfig) StringizeRelease(r *sentry.Release) string {
	if r == nil {
		return ""
	} else {
		return r.Version
	}
}

func emptyIfNull(s *string) string {
	if s == nil {
		return ""
	} else {
		return *s
	}
}

func zeroIfNull(i *int) int {
	if i == nil {
		return 0
	} else {
		return *i
	}
}

func (config *ExporterConfig) PopulateWithEvent(r *ExporterResult, event sentry.Event) {
	r.EventID = event.EventID
	r.EventType = emptyIfNull(event.Type)
	r.Message = emptyIfNull(event.Message)
	r.Release = config.StringizeRelease(event.Release)
	r.EventCreated = *event.DateCreated
	if event.DateReceived != nil {
		r.EventReceived = *event.DateReceived
	}
	r.Platform = emptyIfNull(event.Platform)
	r.GroupID = emptyIfNull(event.GroupID)
}

func (config *ExporterConfig) BuildExporterResultsForSingleIssue(issue sentry.Issue, events []sentry.Event) []ExporterResult {
	results := []ExporterResult{
		{
			AssignedTo:      config.StringizeInternalUser(issue.AssignedTo),
			Count:           emptyIfNull(issue.Count),
			Culprit:         emptyIfNull(issue.Culprit),
			FirstSeen:       *issue.FirstSeen,
			IssueID:         emptyIfNull(issue.ID),
			LastSeen:        *issue.LastSeen,
			Level:           emptyIfNull(issue.Level),
			Logger:          emptyIfNull(issue.Logger),
			Permalink:       emptyIfNull(issue.Permalink),
			Project:         config.StringizeProject(issue.Project),
			ShareID:         emptyIfNull(issue.ShareID),
			ShortID:         emptyIfNull(issue.ShortID),
			Status:          config.StringizeStatus(issue.Status),
			Title:           emptyIfNull(issue.Title),
			IssueType:       emptyIfNull(issue.Type),
			UserCount:       zeroIfNull(issue.UserCount),
			UserReportCount: zeroIfNull(issue.UserReportCount),
		},
	}
	if events != nil {
		for _, event := range events {
			r := results[0]
			config.PopulateWithEvent(&r, event)
			results = append(results, r)
		}
		results = results[1:]
	}
	return results
}

func (exp *Exporter) RenderIssues(issues []sentry.Issue) error {
	if exp.IncludeEvents {
		for _, issue := range issues {
			events, link, err := exp.Client.GetIssueEvents(issue)
			if err != nil {
				return errors.Wrapf(err, "failed to retrieve events for issue %s", *issue.ID)
			}
			partialResults := exp.BuildExporterResultsForSingleIssue(issue, events)
			err = exp.Renderer.RenderPartialResults(partialResults)
			if err != nil {
				return errors.Wrapf(err, "failed to render results")
			}
			for link != nil && link.Next.Results {
				events = []sentry.Event{}
				nextLink, err := exp.Client.GetPage(link.Next, &events)
				if err != nil {
					return errors.Wrapf(err, "failed to retrieve events for issue %s", *issue.ID)
				}
				partialResults := exp.BuildExporterResultsForSingleIssue(issue, events)
				err = exp.Renderer.RenderPartialResults(partialResults)
				if err != nil {
					return errors.Wrapf(err, "failed to render results")
				}
				link = nextLink
			}
		}
	} else {
		for _, issue := range issues {
			partialResults := exp.BuildExporterResultsForSingleIssue(issue, nil)
			err := exp.Renderer.RenderPartialResults(partialResults)
			if err != nil {
				return errors.Wrapf(err, "failed to render results")
			}
		}
	}
	return nil
}

func (exp *Exporter) Export() error {
	defer exp.Renderer.Fini()

	err := exp.Renderer.RenderHeader()
	if err != nil {
		return errors.Wrapf(err, "failed to render a header")
	}
	org, err := exp.Client.GetOrganization(exp.Organization)
	if err != nil {
		return errors.Wrapf(err, "failed to retrieve organization %s", exp.Organization)
	}
	proj, err := exp.Client.GetProject(org, exp.Project)
	if err != nil {
		return errors.Wrapf(err, "failed to retrieve project %s in organization %s", exp.Project, exp.Organization)
	}
	issues, link, err := exp.Client.GetIssues(org, proj, exp.StatsPeriod, nil, exp.Query)
	if err != nil {
		return errors.Wrapf(err, "failed to retrieve issues")
	}
	err = exp.RenderIssues(issues)
	if err != nil {
		return err
	}
	for link != nil && link.Next.Results {
		issues = []sentry.Issue{}
		nextLink, err := exp.Client.GetPage(link.Next, &issues)
		if err != nil {
			return errors.Wrapf(err, "failed to retrieve issues")
		}
		err = exp.RenderIssues(issues)
		if err != nil {
			return err
		}
		link = nextLink
	}

	err = exp.Renderer.RenderFooter()
	if err != nil {
		return errors.Wrapf(err, "failed to render a footer")
	}
	return nil
}
