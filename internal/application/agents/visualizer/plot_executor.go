package visualizer

import (
	"context"
	"encoding/base64"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

var plotCodeFencePattern = regexp.MustCompile("(?s)```(?:python)?\\s*(.*?)```")

type PlotExecutionResult struct {
	Bytes    []byte
	MIMEType string
}

type PlotExecutor interface {
	Execute(ctx context.Context, code string) (PlotExecutionResult, error)
}

type pythonPlotExecutor struct {
	command string
}

func NewPlotExecutor() PlotExecutor {
	return pythonPlotExecutor{command: "python3"}
}

func (e pythonPlotExecutor) Execute(ctx context.Context, code string) (PlotExecutionResult, error) {
	cleaned := extractPlotCode(code)
	if cleaned == "" {
		return PlotExecutionResult{}, fmt.Errorf("plot code is empty")
	}

	cmd := exec.CommandContext(ctx, e.command, "-c", plotExecutorScript)
	cmd.Stdin = strings.NewReader(cleaned)

	output, err := cmd.CombinedOutput()
	if err != nil {
		message := strings.TrimSpace(string(output))
		if message == "" {
			message = err.Error()
		}
		return PlotExecutionResult{}, fmt.Errorf("execute plot code: %s", message)
	}

	bytes, err := base64.StdEncoding.DecodeString(strings.TrimSpace(string(output)))
	if err != nil {
		return PlotExecutionResult{}, fmt.Errorf("decode rendered plot output: %w", err)
	}

	return PlotExecutionResult{
		Bytes:    bytes,
		MIMEType: "image/jpeg",
	}, nil
}

func extractPlotCode(code string) string {
	matches := plotCodeFencePattern.FindStringSubmatch(code)
	if len(matches) == 2 {
		return strings.TrimSpace(matches[1])
	}
	return strings.TrimSpace(code)
}

const plotExecutorScript = `
import base64
import io
import sys

import matplotlib.pyplot as plt

code = sys.stdin.read()

plt.switch_backend("Agg")
plt.close("all")
plt.rcdefaults()

namespace = {}
exec(code, namespace)

if not plt.get_fignums():
    raise SystemExit("plot code did not create a figure")

buffer = io.BytesIO()
plt.savefig(buffer, format="jpeg", bbox_inches="tight", dpi=300)
plt.close("all")
sys.stdout.write(base64.b64encode(buffer.getvalue()).decode("utf-8"))
`
