package config

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	yaml "go.yaml.in/yaml/v3"
)

type NodeCatalog struct {
	CustomNodes []NodeDefinition `yaml:"custom_nodes"`
}

type NodeDefinition struct {
	Name              string                `yaml:"name"`
	URL               string                `yaml:"url"`
	Method            string                `yaml:"method"`
	Headers           map[string]string     `yaml:"headers"`
	RequestTemplate   map[string]any        `yaml:"request_template"`
	ResponseParser    string                `yaml:"response_parser"`
	ResponseSelectors NodeResponseSelectors `yaml:"response_selectors"`
}

type NodeResponseSelectors struct {
	ImageBase64 string `yaml:"image_base64"`
	MIMEType    string `yaml:"mime_type"`
	Summary     string `yaml:"summary"`
	PlotCode    string `yaml:"plot_code"`
}

func LoadNodeConfig(path string) (*NodeCatalog, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read node config %s: %w", path, err)
	}

	decoder := yaml.NewDecoder(bytes.NewReader(raw))
	decoder.KnownFields(true)

	var cfg NodeCatalog
	if err := decoder.Decode(&cfg); err != nil {
		return nil, fmt.Errorf("decode node config %s: %w", path, err)
	}

	var extra any
	if err := decoder.Decode(&extra); err == nil {
		return nil, errors.New("node config must contain a single YAML document")
	} else if !errors.Is(err, io.EOF) {
		return nil, fmt.Errorf("decode trailing node config document %s: %w", path, err)
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func (c *NodeCatalog) NodeByName(name string) (NodeDefinition, bool) {
	for _, node := range c.CustomNodes {
		if node.Name == name {
			return node, true
		}
	}

	return NodeDefinition{}, false
}

func (c *NodeCatalog) validate() error {
	seen := make(map[string]struct{}, len(c.CustomNodes))

	for i := range c.CustomNodes {
		node := &c.CustomNodes[i]
		prefix := fmt.Sprintf("custom_nodes[%d]", i)
		node.expandEnv()

		if strings.TrimSpace(node.Name) == "" {
			return fmt.Errorf("%s.name must be set", prefix)
		}
		if _, exists := seen[node.Name]; exists {
			return fmt.Errorf("%s.name %q must be unique", prefix, node.Name)
		}
		seen[node.Name] = struct{}{}

		if strings.TrimSpace(node.URL) == "" {
			return fmt.Errorf("%s.url must be set", prefix)
		}
		if strings.TrimSpace(node.Method) == "" {
			return fmt.Errorf("%s.method must be set", prefix)
		}
		node.Method = strings.ToUpper(node.Method)

		if len(node.RequestTemplate) == 0 {
			return fmt.Errorf("%s.request_template must be set", prefix)
		}

		if strings.TrimSpace(node.ResponseParser) == "" {
			return fmt.Errorf("%s.response_parser must be set", prefix)
		}
		if node.ResponseParser != "json_path" {
			return fmt.Errorf("%s.response_parser %q is not supported", prefix, node.ResponseParser)
		}
		if strings.TrimSpace(node.ResponseSelectors.ImageBase64) == "" {
			return fmt.Errorf("%s.response_selectors.image_base64 must be set", prefix)
		}
		if strings.TrimSpace(node.ResponseSelectors.MIMEType) == "" {
			return fmt.Errorf("%s.response_selectors.mime_type must be set", prefix)
		}
		if strings.TrimSpace(node.ResponseSelectors.Summary) == "" {
			return fmt.Errorf("%s.response_selectors.summary must be set", prefix)
		}
		if err := validateSelectorPath(prefix+".response_selectors.image_base64", node.ResponseSelectors.ImageBase64); err != nil {
			return err
		}
		if err := validateSelectorPath(prefix+".response_selectors.mime_type", node.ResponseSelectors.MIMEType); err != nil {
			return err
		}
		if err := validateSelectorPath(prefix+".response_selectors.summary", node.ResponseSelectors.Summary); err != nil {
			return err
		}
		if strings.TrimSpace(node.ResponseSelectors.PlotCode) != "" {
			if err := validateSelectorPath(prefix+".response_selectors.plot_code", node.ResponseSelectors.PlotCode); err != nil {
				return err
			}
		}
	}

	return nil
}

func (n *NodeDefinition) expandEnv() {
	n.URL = os.ExpandEnv(strings.TrimSpace(n.URL))
	n.Method = strings.TrimSpace(n.Method)
	n.ResponseParser = strings.TrimSpace(n.ResponseParser)
	n.ResponseSelectors.ImageBase64 = os.ExpandEnv(strings.TrimSpace(n.ResponseSelectors.ImageBase64))
	n.ResponseSelectors.MIMEType = os.ExpandEnv(strings.TrimSpace(n.ResponseSelectors.MIMEType))
	n.ResponseSelectors.Summary = os.ExpandEnv(strings.TrimSpace(n.ResponseSelectors.Summary))
	n.ResponseSelectors.PlotCode = os.ExpandEnv(strings.TrimSpace(n.ResponseSelectors.PlotCode))

	if len(n.Headers) > 0 {
		expandedHeaders := make(map[string]string, len(n.Headers))
		for key, value := range n.Headers {
			expandedHeaders[os.ExpandEnv(strings.TrimSpace(key))] = os.ExpandEnv(value)
		}
		n.Headers = expandedHeaders
	}

	if len(n.RequestTemplate) > 0 {
		n.RequestTemplate = expandNodeTemplate(n.RequestTemplate).(map[string]any)
	}
}

func expandNodeTemplate(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		expanded := make(map[string]any, len(typed))
		for key, nested := range typed {
			expanded[key] = expandNodeTemplate(nested)
		}
		return expanded
	case []any:
		expanded := make([]any, len(typed))
		for i, item := range typed {
			expanded[i] = expandNodeTemplate(item)
		}
		return expanded
	case string:
		return os.ExpandEnv(typed)
	default:
		return value
	}
}

func validateSelectorPath(field, path string) error {
	path = strings.TrimSpace(path)
	if path == "$" || strings.HasPrefix(path, "$.") {
		return nil
	}

	return fmt.Errorf("%s must start with $.", field)
}
