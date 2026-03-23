// Package config provides domain entities for LLM provider configuration.
package config

import "time"

// ProviderType represents predefined provider types.
type ProviderType string

const (
	ProviderTypeOpenAI           ProviderType = "openai"
	ProviderTypeAnthropic        ProviderType = "anthropic"
	ProviderTypeGemini           ProviderType = "gemini"
	ProviderTypeDeepSeek         ProviderType = "deepseek"
	ProviderTypeZhipu            ProviderType = "zhipu"
	ProviderTypeMoonshot         ProviderType = "moonshot"
	ProviderTypeQwen             ProviderType = "qwen"
	ProviderTypeDoubao           ProviderType = "doubao"
	ProviderTypeBaichuan         ProviderType = "baichuan"
	ProviderTypeMinimax          ProviderType = "minimax"
	ProviderTypeYi               ProviderType = "yi"
	ProviderTypeHunyuan          ProviderType = "hunyuan"
	ProviderTypeStepfun          ProviderType = "stepfun"
	ProviderTypeSilicon          ProviderType = "silicon"
	ProviderTypeOpenRouter       ProviderType = "openrouter"
	ProviderTypeOllama           ProviderType = "ollama"
	ProviderTypeOpenAICompatible ProviderType = "openai-compatible"
)

// ModelInfo represents a model configuration.
type ModelInfo struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	MaxTokens     int    `json:"max_tokens,omitempty"`
	SupportsVision bool   `json:"supports_vision,omitempty"`
	Enabled       bool   `json:"enabled"`
}

// SystemProviderPreset defines a predefined provider configuration.
type SystemProviderPreset struct {
	Type           ProviderType `json:"type"`
	Name           string       `json:"name"`
	DisplayName    string       `json:"display_name"`
	APIHost        string       `json:"api_host"`
	DocsURL        string       `json:"docs_url"`
	APIKeyURL      string       `json:"api_key_url"`
	DefaultModels  []ModelInfo  `json:"default_models"`
	SupportsVision bool         `json:"supports_vision"`
}

// SystemProviderPresets returns all predefined provider presets.
func SystemProviderPresets() []SystemProviderPreset {
	return []SystemProviderPreset{
		{
			Type:        ProviderTypeOpenAI,
			Name:        "openai",
			DisplayName: "OpenAI",
			APIHost:     "https://api.openai.com/v1",
			DocsURL:     "https://platform.openai.com/docs",
			APIKeyURL:   "https://platform.openai.com/api-keys",
			DefaultModels: []ModelInfo{
				{ID: "gpt-4o", Name: "GPT-4o", SupportsVision: true, Enabled: true},
				{ID: "gpt-4o-mini", Name: "GPT-4o Mini", SupportsVision: true, Enabled: true},
				{ID: "gpt-4-turbo", Name: "GPT-4 Turbo", SupportsVision: true, Enabled: true},
			},
			SupportsVision: true,
		},
		{
			Type:        ProviderTypeAnthropic,
			Name:        "anthropic",
			DisplayName: "Anthropic",
			APIHost:     "https://api.anthropic.com/v1",
			DocsURL:     "https://docs.anthropic.com",
			APIKeyURL:   "https://console.anthropic.com/settings/keys",
			DefaultModels: []ModelInfo{
				{ID: "claude-sonnet-4-20250514", Name: "Claude Sonnet 4", SupportsVision: true, Enabled: true},
				{ID: "claude-3-5-sonnet-20241022", Name: "Claude 3.5 Sonnet", SupportsVision: true, Enabled: true},
				{ID: "claude-3-haiku-20240307", Name: "Claude 3 Haiku", Enabled: true},
			},
			SupportsVision: true,
		},
		{
			Type:        ProviderTypeGemini,
			Name:        "gemini",
			DisplayName: "Google Gemini",
			APIHost:     "https://generativelanguage.googleapis.com/v1beta",
			DocsURL:     "https://ai.google.dev/gemini-api/docs",
			APIKeyURL:   "https://aistudio.google.com/app/apikey",
			DefaultModels: []ModelInfo{
				{ID: "gemini-2.0-flash", Name: "Gemini 2.0 Flash", SupportsVision: true, Enabled: true},
				{ID: "gemini-1.5-pro", Name: "Gemini 1.5 Pro", SupportsVision: true, Enabled: true},
				{ID: "gemini-1.5-flash", Name: "Gemini 1.5 Flash", SupportsVision: true, Enabled: true},
			},
			SupportsVision: true,
		},
		{
			Type:        ProviderTypeDeepSeek,
			Name:        "deepseek",
			DisplayName: "DeepSeek (深度求索)",
			APIHost:     "https://api.deepseek.com",
			DocsURL:     "https://platform.deepseek.com/api-docs",
			APIKeyURL:   "https://platform.deepseek.com/api_keys",
			DefaultModels: []ModelInfo{
				{ID: "deepseek-chat", Name: "DeepSeek Chat", Enabled: true},
				{ID: "deepseek-reasoner", Name: "DeepSeek Reasoner", Enabled: true},
			},
		},
		{
			Type:        ProviderTypeZhipu,
			Name:        "zhipu",
			DisplayName: "智谱 GLM",
			APIHost:     "https://open.bigmodel.cn/api/paas/v4",
			DocsURL:     "https://open.bigmodel.cn/dev/api",
			APIKeyURL:   "https://open.bigmodel.cn/usercenter/apikeys",
			DefaultModels: []ModelInfo{
				{ID: "glm-4-flash", Name: "GLM-4 Flash", Enabled: true},
				{ID: "glm-4-plus", Name: "GLM-4 Plus", Enabled: true},
				{ID: "glm-4-air", Name: "GLM-4 Air", Enabled: true},
			},
			SupportsVision: true,
		},
		{
			Type:        ProviderTypeMoonshot,
			Name:        "moonshot",
			DisplayName: "月之暗面 (Moonshot)",
			APIHost:     "https://api.moonshot.cn/v1",
			DocsURL:     "https://platform.moonshot.cn/docs",
			APIKeyURL:   "https://platform.moonshot.cn/console/api-keys",
			DefaultModels: []ModelInfo{
				{ID: "moonshot-v1-8k", Name: "Moonshot V1 8K", Enabled: true},
				{ID: "moonshot-v1-32k", Name: "Moonshot V1 32K", Enabled: true},
				{ID: "moonshot-v1-128k", Name: "Moonshot V1 128K", Enabled: true},
			},
		},
		{
			Type:        ProviderTypeQwen,
			Name:        "qwen",
			DisplayName: "阿里云百炼 (Qwen)",
			APIHost:     "https://dashscope.aliyuncs.com/compatible-mode/v1",
			DocsURL:     "https://help.aliyun.com/zh/dashscope",
			APIKeyURL:   "https://dashscope.console.aliyun.com/apiKey",
			DefaultModels: []ModelInfo{
				{ID: "qwen-turbo", Name: "Qwen Turbo", Enabled: true},
				{ID: "qwen-plus", Name: "Qwen Plus", Enabled: true},
				{ID: "qwen-max", Name: "Qwen Max", Enabled: true},
			},
			SupportsVision: true,
		},
		{
			Type:        ProviderTypeDoubao,
			Name:        "doubao",
			DisplayName: "字节豆包 (Doubao)",
			APIHost:     "https://ark.cn-beijing.volces.com/api/v3",
			DocsURL:     "https://www.volcengine.com/docs/82379",
			APIKeyURL:   "https://console.volcengine.com/ark/region:ark+cn-beijing/apiKey",
			DefaultModels: []ModelInfo{
				{ID: "doubao-pro-32k", Name: "Doubao Pro 32K", Enabled: true},
				{ID: "doubao-pro-128k", Name: "Doubao Pro 128K", Enabled: true},
			},
		},
		{
			Type:        ProviderTypeBaichuan,
			Name:        "baichuan",
			DisplayName: "百川智能 (Baichuan)",
			APIHost:     "https://api.baichuan-ai.com/v1",
			DocsURL:     "https://platform.baichuan-ai.com/docs/api",
			APIKeyURL:   "https://platform.baichuan-ai.com/console/apikey",
			DefaultModels: []ModelInfo{
				{ID: "Baichuan4", Name: "Baichuan4", Enabled: true},
				{ID: "Baichuan3-Turbo", Name: "Baichuan3 Turbo", Enabled: true},
			},
		},
		{
			Type:        ProviderTypeMinimax,
			Name:        "minimax",
			DisplayName: "MiniMax",
			APIHost:     "https://api.minimax.chat/v1",
			DocsURL:     "https://www.minimax.io/docs",
			APIKeyURL:   "https://www.minimax.io/user-center/basic-information/interface-key",
			DefaultModels: []ModelInfo{
				{ID: "abab6.5s-chat", Name: "ABAB 6.5s Chat", Enabled: true},
				{ID: "abab6.5-chat", Name: "ABAB 6.5 Chat", Enabled: true},
			},
		},
		{
			Type:        ProviderTypeYi,
			Name:        "yi",
			DisplayName: "零一万物 (Yi)",
			APIHost:     "https://api.lingyiwanwu.com/v1",
			DocsURL:     "https://platform.lingyiwanwu.com/docs",
			APIKeyURL:   "https://platform.lingyiwanwu.com/apikeys",
			DefaultModels: []ModelInfo{
				{ID: "yi-lightning", Name: "Yi Lightning", Enabled: true},
				{ID: "yi-large", Name: "Yi Large", Enabled: true},
			},
		},
		{
			Type:        ProviderTypeHunyuan,
			Name:        "hunyuan",
			DisplayName: "腾讯混元 (Hunyuan)",
			APIHost:     "https://api.hunyuan.cloud.tencent.com/v1",
			DocsURL:     "https://cloud.tencent.com/document/product/1729",
			APIKeyURL:   "https://console.cloud.tencent.com/cam/capi",
			DefaultModels: []ModelInfo{
				{ID: "hunyuan-lite", Name: "Hunyuan Lite", Enabled: true},
				{ID: "hunyuan-standard", Name: "Hunyuan Standard", Enabled: true},
				{ID: "hunyuan-pro", Name: "Hunyuan Pro", Enabled: true},
			},
		},
		{
			Type:        ProviderTypeStepfun,
			Name:        "stepfun",
			DisplayName: "阶跃星辰 (StepFun)",
			APIHost:     "https://api.stepfun.com/v1",
			DocsURL:     "https://platform.stepfun.com/docs",
			APIKeyURL:   "https://platform.stepfun.com/console/api-key",
			DefaultModels: []ModelInfo{
				{ID: "step-1-8k", Name: "Step 1 8K", Enabled: true},
				{ID: "step-1-32k", Name: "Step 1 32K", Enabled: true},
			},
		},
		{
			Type:        ProviderTypeSilicon,
			Name:        "silicon",
			DisplayName: "硅基流动 (Silicon Cloud)",
			APIHost:     "https://api.siliconflow.cn/v1",
			DocsURL:     "https://docs.siliconflow.cn",
			APIKeyURL:   "https://cloud.siliconflow.cn/account/ak",
			DefaultModels: []ModelInfo{
				{ID: "Qwen/Qwen2.5-7B-Instruct", Name: "Qwen2.5 7B", Enabled: true},
				{ID: "deepseek-ai/DeepSeek-V2.5", Name: "DeepSeek V2.5", Enabled: true},
			},
		},
		{
			Type:        ProviderTypeOpenRouter,
			Name:        "openrouter",
			DisplayName: "OpenRouter",
			APIHost:     "https://openrouter.ai/api/v1",
			DocsURL:     "https://openrouter.ai/docs",
			APIKeyURL:   "https://openrouter.ai/keys",
			DefaultModels: []ModelInfo{
				{ID: "anthropic/claude-3.5-sonnet", Name: "Claude 3.5 Sonnet", SupportsVision: true, Enabled: true},
				{ID: "openai/gpt-4o", Name: "GPT-4o", SupportsVision: true, Enabled: true},
			},
			SupportsVision: true,
		},
		{
			Type:        ProviderTypeOllama,
			Name:        "ollama",
			DisplayName: "Ollama (本地)",
			APIHost:     "http://localhost:11434/v1",
			DocsURL:     "https://ollama.com/library",
			APIKeyURL:    "",
			DefaultModels: []ModelInfo{
				{ID: "llama3.2", Name: "Llama 3.2", Enabled: true},
				{ID: "qwen2.5", Name: "Qwen 2.5", Enabled: true},
			},
		},
	}
}

// GetPresetByType returns a preset by provider type.
func GetPresetByType(t ProviderType) *SystemProviderPreset {
	for _, p := range SystemProviderPresets() {
		if p.Type == t {
			return &p
		}
	}
	return nil
}

// BuiltInPresets returns all system provider presets (alias for SystemProviderPresets).
func BuiltInPresets() []SystemProviderPreset {
	return SystemProviderPresets()
}

// GetPresetByName returns a preset by provider name.
func GetPresetByName(name string) *SystemProviderPreset {
	for _, p := range SystemProviderPresets() {
		if p.Name == name {
			return &p
		}
	}
	return nil
}

// Provider represents an LLM provider configuration.
type Provider struct {
	ID          string       `json:"id"`
	Type        ProviderType `json:"type"`
	Name        string       `json:"name"`
	DisplayName string       `json:"display_name"`
	APIHost     string       `json:"api_host"`
	APIKey      string       `json:"-"` // Never expose API key in JSON
	Models      []ModelInfo  `json:"models"`
	Enabled     bool         `json:"enabled"`
	IsSystem    bool         `json:"is_system"`
	IsDefault   bool         `json:"is_default"`
	TimeoutMs   int          `json:"timeout_ms"`
	// Task-specific model selection
	QueryModel string `json:"query_model"` // Model for retrieval/planning/critique
	GenModel   string `json:"gen_model"`   // Model for visualization generation
	CreatedAt  time.Time    `json:"created_at"`
	UpdatedAt  time.Time    `json:"updated_at"`
}

// BaseURL returns the API host (alias for backward compatibility).
func (p *Provider) BaseURL() string {
	return p.APIHost
}

// GetDefaultModel returns the first enabled model, or a fallback.
func (p *Provider) GetDefaultModel() string {
	for _, m := range p.Models {
		if m.Enabled {
			return m.ID
		}
	}
	return "default"
}

// ProviderRepository defines the interface for provider persistence.
type ProviderRepository interface {
	Create(provider *Provider) error
	GetByID(id string) (*Provider, error)
	GetByName(name string) (*Provider, error)
	List() ([]*Provider, error)
	ListEnabled() ([]*Provider, error)
	Update(provider *Provider) error
	Delete(id string) error
	SetDefault(id string) error
	GetDefault() (*Provider, error)
	InitializeSystemProviders() error
}
