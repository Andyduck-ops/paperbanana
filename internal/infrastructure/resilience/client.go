package resilience

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/sony/gobreaker"
)

type ResilientClient struct {
	client  *http.Client
	breaker *gobreaker.CircuitBreaker
}

func NewResilientClient(name string, timeout time.Duration) *ResilientClient {
	breaker := gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:        name,
		MaxRequests: 1,
		Interval:    30 * time.Second,
		Timeout:     30 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures >= 5
		},
	})

	transport := &retryTransport{
		base:    http.DefaultTransport,
		breaker: breaker,
	}

	return &ResilientClient{
		client: &http.Client{
			Timeout:   defaultTimeout(timeout),
			Transport: transport,
		},
		breaker: breaker,
	}
}

func (rc *ResilientClient) Do(req *http.Request) (*http.Response, error) {
	return rc.client.Do(req)
}

func (rc *ResilientClient) HTTPClient() *http.Client {
	return rc.client
}

type retryTransport struct {
	base    http.RoundTripper
	breaker *gobreaker.CircuitBreaker
}

func (rt *retryTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	body, err := snapshotBody(req)
	if err != nil {
		return nil, err
	}

	var resp *http.Response
	operation := func() error {
		clonedReq, err := cloneRequest(req, body)
		if err != nil {
			return backoff.Permanent(err)
		}

		result, err := rt.breaker.Execute(func() (interface{}, error) {
			return rt.base.RoundTrip(clonedReq)
		})
		if err != nil {
			return err
		}

		resp = result.(*http.Response)
		if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= http.StatusInternalServerError {
			resp.Body.Close()
			return fmt.Errorf("retryable status: %d", resp.StatusCode)
		}

		return nil
	}

	expBackoff := backoff.NewExponentialBackOff()
	expBackoff.InitialInterval = 20 * time.Millisecond
	expBackoff.MaxInterval = 100 * time.Millisecond
	expBackoff.MaxElapsedTime = 800 * time.Millisecond

	if err := backoff.Retry(operation, expBackoff); err != nil {
		return nil, err
	}

	return resp, nil
}

func snapshotBody(req *http.Request) ([]byte, error) {
	if req.Body == nil {
		return nil, nil
	}

	body, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, fmt.Errorf("read request body: %w", err)
	}
	req.Body = io.NopCloser(bytes.NewReader(body))

	return body, nil
}

func cloneRequest(req *http.Request, body []byte) (*http.Request, error) {
	cloned := req.Clone(req.Context())
	if body != nil {
		cloned.Body = io.NopCloser(bytes.NewReader(body))
		cloned.ContentLength = int64(len(body))
	} else {
		cloned.Body = nil
	}
	return cloned, nil
}

func defaultTimeout(timeout time.Duration) time.Duration {
	if timeout > 0 {
		return timeout
	}
	return 60 * time.Second
}
