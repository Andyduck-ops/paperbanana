package modelselection

import "testing"

func TestQueryModel_UsesRuntimeManagedProviderDefaultWhenNoOverride(t *testing.T) {
	metadata := map[string]string{
		RuntimeManagedMetadataKey: "true",
	}

	got := QueryModel(metadata, "grok-4.1-fast")
	if got != "" {
		t.Fatalf("expected empty query model so runtime provider default can take over, got %q", got)
	}
}

func TestGenerationModel_UsesRuntimeManagedProviderDefaultWhenNoOverride(t *testing.T) {
	metadata := map[string]string{
		RuntimeManagedMetadataKey: "true",
	}

	got := GenerationModel(metadata, "grok-4.1-fast")
	if got != "" {
		t.Fatalf("expected empty generation model so runtime provider gen_model can take over, got %q", got)
	}
}

func TestGenerationModel_PreservesExplicitOverrideWhenRuntimeManaged(t *testing.T) {
	metadata := map[string]string{
		RuntimeManagedMetadataKey:  "true",
		GenerationModelMetadataKey: "grok:grok-imagine-1.0",
	}

	got := GenerationModel(metadata, "grok-4.1-fast")
	if got != "grok:grok-imagine-1.0" {
		t.Fatalf("expected explicit generation override to win, got %q", got)
	}
}
