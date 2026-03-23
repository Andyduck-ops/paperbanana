package modelselection

import "strings"

const (
	QueryModelMetadataKey    = "config.query_model"
	GenerationModelMetadataKey = "config.gen_model"
	RetrievalModeMetadataKey = "config.retrieval_mode"
)

func QueryModel(metadata map[string]string, fallback string) string {
	return resolve(metadata, QueryModelMetadataKey, fallback)
}

func GenerationModel(metadata map[string]string, fallback string) string {
	return resolve(metadata, GenerationModelMetadataKey, fallback)
}

func RetrievalMode(metadata map[string]string, fallback string) string {
	if value := strings.TrimSpace(metadata[RetrievalModeMetadataKey]); value != "" {
		return value
	}
	if value := strings.TrimSpace(metadata["retrieval_setting"]); value != "" {
		return value
	}
	return fallback
}

func resolve(metadata map[string]string, key, fallback string) string {
	if value := strings.TrimSpace(metadata[key]); value != "" {
		return value
	}
	return fallback
}
