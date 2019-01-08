# sentry-event-exporter

This is a tiny utility written in Go that retrieves submitted events from [Sentry](https://sentry.io) and exports as CSV.

## Building

```
$ go get github.com/opencollector/sentry-event-exporter/cmd/sentry-event-exporter
```

## Synopsis

```
$ sentry-event-exporter -help
Usage of sentry-event-exporter:
  -authtoken string
        Sentry authentication token (get it at https://sentry.io/settings/account/api/auth-tokens/ )
  -endpoint string
        Sentry endpoint
  -events
        include events in the result
  -organization string
        organization
  -project string
        project
```

* `-authtoken` (defaults to the value of `SENTRY_AUTHTOKEN` environment variable)

	Specifies the authentication token that can be created and retrieved at https://sentry.io/settings/account/api/auth-tokens/ .

* `-endpoint` (defaults to the endpoint managed by sentry.io)

	Specifies the alternate endpoint URL. (Supply it if you want to use your own Sentry instance)

* `-events`

	When unspecified, only issues are dumped.  Give this option if you want to get the events for those too.

* `-organization` (mandatory)

	Specifies the organization.

* `-project` (mandatory)

	Specifies the project.


The rest of the arguments are the query.  Refer to the [Search](https://docs.sentry.io/workflow/search/) section of the document for details.

```
sentry-event-exporter -authtoken xxx -organization yyy -project zzz is:unresolved 'browser:IE 11.0'
```


## License

```
Copyright (c) 2019 Open Collector, Inc.
                   Moriyoshi Koizumi

Permission is hereby granted, free of charge, to any person obtaining a copy of
this software and associated documentation files (the "Software"), to deal in
the Software without restriction, including without limitation the rights to
use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies
of the Software, and to permit persons to whom the Software is furnished to do
so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
```
