// Package httplogger provides functionality to attach a logger to an http request's context.
package httplogger

import (
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync/atomic"

	"git.act3-ace.com/ace/go-common/pkg/logger"
	"git.act3-ace.com/ace/go-common/pkg/redact"
)

var requestNumber atomic.Int64

/*
 To extract the HTTP Request style log from the json log use
 jq -r -f cmd/ace-dt/internal/cli/log-http.jq log.jsonl > log.http
*/

// LoggingTransport logs to the request's context.
// The output can be processed by jq to format it nicely.
type LoggingTransport struct {
	Base http.RoundTripper
}

// RoundTrip logs http requests and reponses while redacting sensistive information.
func (s *LoggingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	ctx := req.Context()
	log := logger.V(logger.FromContext(ctx).WithGroup("http").With("requestID", requestNumber.Add(1)), 8)
	const maxSize = 10 * 1024

	enabled := log.Enabled(ctx, slog.LevelInfo)
	if enabled {
		req := req.Clone(ctx)
		// redact the URL credentials and query string (S3 signed URLs have credentials there)
		req.URL.User = nil
		req.URL.RawQuery = ""

		redactHTTPHeaders(req.Header)

		reqBytes, err := httputil.DumpRequestOut(req, req.ContentLength < maxSize)
		if err != nil {
			log.ErrorContext(ctx, "Failed to dump the HTTP request", "error", err.Error())
		} else {
			log.InfoContext(ctx, "HTTP Request", "contents", string(reqBytes))
		}
	}

	resp, err := s.Base.RoundTrip(req)
	// err is returned after dumping the response

	// need to check if response is nil so that go doesn't panic w/ segfault
	if resp != nil && enabled {
		savedHeaders := resp.Header.Clone()
		redactHTTPHeaders(resp.Header)
		// TODO redact the body of the auth response
		// for now we always omit the body to be conservative
		respBytes, err := httputil.DumpResponse(resp, false) // resp.ContentLength < maxSize)
		if err != nil {
			log.ErrorContext(ctx, "Failed to dump the HTTP response", "error", err.Error())
		} else {
			log.InfoContext(ctx, "HTTP Response", "contents", string(respBytes))
		}

		// restore then
		resp.Header = savedHeaders
	}

	return resp, err //nolint:wrapcheck
}

var redactedHeaders = []string{
	"Authorization",
	"Cookie", // probably not needed but why not
	"Set-Cookie",
}

// redact http headers in place.
func redactHTTPHeaders(hdrs http.Header) {
	// redact headers Authorization, Cookie, Set-Cookie
	// redact query params of Location headers
	for _, h := range redactedHeaders {
		values := hdrs.Values(h)
		for i, value := range values {
			values[i] = redact.String(value)
		}
	}

	values := hdrs.Values("Location")
	for i, value := range values {
		values[i] = redactURL(value)
	}
}

// redact the URL inplace removing user credentials and query string params.
func redactURL(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	u.User = nil
	u.RawQuery = ""
	return u.String()
}
