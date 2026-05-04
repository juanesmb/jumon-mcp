package observability

import (
	stdhttp "net/http"
	"time"

	infrahttppkg "jumon-mcp/internal/infrastructure/http"
)

// HTTPTransport wraps outbound traffic with structured logs and metrics layered on otel instrumentation.
func (r *Recorder) HTTPTransport(next stdhttp.RoundTripper) stdhttp.RoundTripper {
	return roundTripObserve{r: r, next: next}
}

type roundTripObserve struct {
	r    *Recorder
	next stdhttp.RoundTripper
}

func (rt roundTripObserve) RoundTrip(req *stdhttp.Request) (*stdhttp.Response, error) {
	if rt.next == nil {
		rt.next = stdhttp.DefaultTransport
	}
	start := time.Now()
	attempt := infrahttppkg.AttemptFromContext(req.Context())

	resp, err := rt.next.RoundTrip(req)

	ms := float64(time.Since(start).Milliseconds())

	method := NormalizeHTTPMethod(req.Method)
	routePattern := GatewayRoutePattern(req.URL.String())
	provider := ProviderFromGatewayURL(req.URL.String())

	status := 0
	retry := attempt > 1
	errSummary := ""
	switch {
	case err != nil:
		errSummary = err.Error()
	case resp != nil:
		status = resp.StatusCode
	default:
		status = 0
		errSummary = "nil_response"
	}

	rt.r.RecordUpstreamHTTP(req.Context(), method, routePattern, provider, status, retry, attempt, ms, errSummary)
	return resp, err
}
