package llm

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateRequestSupportsSystemInstructionAndPromptVersion(t *testing.T) {
	req := GenerateRequest{
		SystemInstruction: "You are the planner.",
		PromptVersion:     "planner-v1",
		Messages: []Message{
			{
				Role:  RoleUser,
				Parts: []Part{TextPart("Describe the figure.")},
			},
		},
	}

	require.Len(t, req.Messages, 1)
	assert.Equal(t, "You are the planner.", req.SystemInstruction)
	assert.Equal(t, "planner-v1", req.PromptVersion)
	assert.Equal(t, RoleUser, req.Messages[0].Role)
	require.Len(t, req.Messages[0].Parts, 1)
	assert.Equal(t, PartTypeText, req.Messages[0].Parts[0].Type)
	assert.Equal(t, "Describe the figure.", req.Messages[0].Parts[0].Text)
}

func TestGenerateRequestSupportsTextAndImageParts(t *testing.T) {
	req := GenerateRequest{
		Messages: []Message{
			{
				Role: RoleUser,
				Parts: []Part{
					TextPart("Example figure"),
					InlineImagePart("image/png", []byte{0x01, 0x02, 0x03}),
				},
			},
		},
	}

	require.Len(t, req.Messages, 1)
	require.Len(t, req.Messages[0].Parts, 2)
	assert.Equal(t, PartTypeText, req.Messages[0].Parts[0].Type)
	assert.Equal(t, "Example figure", req.Messages[0].Parts[0].Text)
	assert.Equal(t, PartTypeImage, req.Messages[0].Parts[1].Type)
	assert.Equal(t, "image/png", req.Messages[0].Parts[1].MIMEType)
	assert.Equal(t, []byte{0x01, 0x02, 0x03}, req.Messages[0].Parts[1].Data)
}
