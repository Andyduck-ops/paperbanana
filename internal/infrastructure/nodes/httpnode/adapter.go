package httpnode

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	pbconfig "github.com/paperbanana/paperbanana/internal/config"
	"github.com/paperbanana/paperbanana/internal/infrastructure/resilience"
)

type Adapter struct {
	httpClient *http.Client
}

type Result struct {
	StatusCode int
	Headers    http.Header
	Body       map[string]any
	RawBody    []byte
}

type ErrorKind string

const (
	ErrorKindBuildRequest   ErrorKind = "build_request"
	ErrorKindTransport      ErrorKind = "transport"
	ErrorKindUnexpectedCode ErrorKind = "unexpected_status"
	ErrorKindParseResponse  ErrorKind = "parse_response"
)

type ExecutionError struct {
	NodeName   string
	Kind       ErrorKind
	StatusCode int
	Err        error
}

func NewAdapter(client *resilience.ResilientClient) *Adapter {
	if client == nil {
		return &Adapter{}
	}

	return &Adapter{httpClient: client.HTTPClient()}
}

func (e *ExecutionError) Error() string {
	if e.StatusCode > 0 {
		return fmt.Sprintf("execute node %s (%s, status=%d): %v", e.NodeName, e.Kind, e.StatusCode, e.Err)
	}

	return fmt.Sprintf("execute node %s (%s): %v", e.NodeName, e.Kind, e.Err)
}

func (e *ExecutionError) Unwrap() error {
	return e.Err
}

func (a *Adapter) Execute(ctx context.Context, node pbconfig.NodeDefinition) (Result, error) {
	if a.httpClient == nil {
		return Result{}, &ExecutionError{
			NodeName: node.Name,
			Kind:     ErrorKindTransport,
			Err:      fmt.Errorf("http client is required"),
		}
	}

	payload, err := json.Marshal(node.RequestTemplate)
	if err != nil {
		return Result{}, &ExecutionError{
			NodeName: node.Name,
			Kind:     ErrorKindBuildRequest,
			Err:      fmt.Errorf("marshal request template: %w", err),
		}
	}

	req, err := http.NewRequestWithContext(ctx, node.Method, node.URL, bytes.NewReader(payload))
	if err != nil {
		return Result{}, &ExecutionError{
			NodeName: node.Name,
			Kind:     ErrorKindBuildRequest,
			Err:      fmt.Errorf("build request: %w", err),
		}
	}

	for key, value := range node.Headers {
		req.Header.Set(key, value)
	}
	if req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return Result{}, &ExecutionError{
			NodeName: node.Name,
			Kind:     ErrorKindTransport,
			Err:      err,
		}
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return Result{}, &ExecutionError{
			NodeName:   node.Name,
			Kind:       ErrorKindTransport,
			StatusCode: resp.StatusCode,
			Err:        fmt.Errorf("read response body: %w", err),
		}
	}

	if resp.StatusCode >= http.StatusBadRequest {
		return Result{}, &ExecutionError{
			NodeName:   node.Name,
			Kind:       ErrorKindUnexpectedCode,
			StatusCode: resp.StatusCode,
			Err:        fmt.Errorf("unexpected response: %s", bytes.TrimSpace(raw)),
		}
	}

	switch node.ResponseParser {
	case "json_path":
		var payload map[string]any
		if err := json.Unmarshal(raw, &payload); err != nil {
			return Result{}, &ExecutionError{
				NodeName:   node.Name,
				Kind:       ErrorKindParseResponse,
				StatusCode: resp.StatusCode,
				Err:        fmt.Errorf("decode json response: %w", err),
			}
		}

		body, err := normalizeJSONPathBody(payload, node.ResponseSelectors)
		if err != nil {
			return Result{}, &ExecutionError{
				NodeName:   node.Name,
				Kind:       ErrorKindParseResponse,
				StatusCode: resp.StatusCode,
				Err:        err,
			}
		}

		return Result{
			StatusCode: resp.StatusCode,
			Headers:    resp.Header.Clone(),
			Body:       body,
			RawBody:    raw,
		}, nil
	default:
		return Result{}, &ExecutionError{
			NodeName: node.Name,
			Kind:     ErrorKindParseResponse,
			Err:      fmt.Errorf("unsupported response parser %q", node.ResponseParser),
		}
	}
}

func normalizeJSONPathBody(payload map[string]any, selectors pbconfig.NodeResponseSelectors) (map[string]any, error) {
	body := make(map[string]any, 4)

	imageBase64, err := jsonPathString(payload, selectors.ImageBase64)
	if err != nil {
		return nil, fmt.Errorf("extract image_base64 via %q: %w", selectors.ImageBase64, err)
	}
	mimeType, err := jsonPathString(payload, selectors.MIMEType)
	if err != nil {
		return nil, fmt.Errorf("extract mime_type via %q: %w", selectors.MIMEType, err)
	}
	summary, err := jsonPathString(payload, selectors.Summary)
	if err != nil {
		return nil, fmt.Errorf("extract summary via %q: %w", selectors.Summary, err)
	}

	body["image_base64"] = imageBase64
	body["mime_type"] = mimeType
	body["summary"] = summary

	if strings.TrimSpace(selectors.PlotCode) != "" {
		plotCode, err := jsonPathString(payload, selectors.PlotCode)
		if err != nil {
			return nil, fmt.Errorf("extract plot_code via %q: %w", selectors.PlotCode, err)
		}
		body["plot_code"] = plotCode
	}

	return body, nil
}

func jsonPathString(payload map[string]any, path string) (string, error) {
	value, err := jsonPathValue(payload, path)
	if err != nil {
		return "", err
	}

	text, ok := value.(string)
	if !ok {
		return "", fmt.Errorf("resolved value has type %T, want string", value)
	}

	return strings.TrimSpace(text), nil
}

func jsonPathValue(payload map[string]any, path string) (any, error) {
	segments, err := parseJSONPath(path)
	if err != nil {
		return nil, err
	}

	var current any = payload
	for _, segment := range segments {
		switch step := segment.(type) {
		case string:
			object, ok := current.(map[string]any)
			if !ok {
				return nil, fmt.Errorf("segment %q resolved against %T, want object", step, current)
			}
			next, ok := object[step]
			if !ok {
				return nil, fmt.Errorf("segment %q not found", step)
			}
			current = next
		case int:
			items, ok := current.([]any)
			if !ok {
				return nil, fmt.Errorf("index %d resolved against %T, want array", step, current)
			}
			if step < 0 || step >= len(items) {
				return nil, fmt.Errorf("index %d out of range", step)
			}
			current = items[step]
		default:
			return nil, fmt.Errorf("unsupported json path segment %v", segment)
		}
	}

	return current, nil
}

func parseJSONPath(path string) ([]any, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return nil, fmt.Errorf("path must be set")
	}
	if path == "$" {
		return nil, nil
	}
	if !strings.HasPrefix(path, "$.") {
		return nil, fmt.Errorf("path must start with $.")
	}

	path = strings.TrimPrefix(path, "$.")
	segments := make([]any, 0, 4)
	var token strings.Builder

	flushToken := func() {
		if token.Len() == 0 {
			return
		}
		segments = append(segments, token.String())
		token.Reset()
	}

	for i := 0; i < len(path); i++ {
		switch path[i] {
		case '.':
			flushToken()
		case '[':
			flushToken()
			end := strings.IndexByte(path[i:], ']')
			if end < 0 {
				return nil, fmt.Errorf("unclosed index in path %q", path)
			}
			end += i
			index, err := strconv.Atoi(path[i+1 : end])
			if err != nil {
				return nil, fmt.Errorf("invalid index %q", path[i+1:end])
			}
			segments = append(segments, index)
			i = end
		default:
			token.WriteByte(path[i])
		}
	}

	flushToken()
	if len(segments) == 0 {
		return nil, fmt.Errorf("path %q must contain at least one selector segment", path)
	}

	return segments, nil
}
