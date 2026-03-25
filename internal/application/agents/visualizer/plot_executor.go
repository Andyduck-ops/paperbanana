package visualizer

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"sync"
	"syscall"
	"time"
)

var plotCodeFencePattern = regexp.MustCompile("(?s)```(?:python)?\\s*(.*?)```")

const (
	defaultGracefulShutdownTimeout = 5 * time.Second
)

type PlotExecutionResult struct {
	Bytes    []byte
	MIMEType string
}

type PlotExecutor interface {
	Execute(ctx context.Context, code string) (PlotExecutionResult, error)
	Cleanup(ctx context.Context) error
}

type pythonPlotExecutor struct {
	command              string
	gracefulTimeout      time.Duration
	mu                   sync.Mutex
	activeProcesses      map[int]*exec.Cmd
	tempFiles            []string
	tempFilesMu          sync.Mutex
}

func NewPlotExecutor() PlotExecutor {
	return &pythonPlotExecutor{
		command:         "python3",
		gracefulTimeout: defaultGracefulShutdownTimeout,
		activeProcesses: make(map[int]*exec.Cmd),
	}
}

func (e *pythonPlotExecutor) Execute(ctx context.Context, code string) (PlotExecutionResult, error) {
	cleaned := extractPlotCode(code)
	if cleaned == "" {
		return PlotExecutionResult{}, fmt.Errorf("plot code is empty")
	}

	cmd := exec.Command(e.command, "-c", plotExecutorScript)
	cmd.Stdin = strings.NewReader(cleaned)

	if err := cmd.Start(); err != nil {
		return PlotExecutionResult{}, fmt.Errorf("start plot process: %w", err)
	}

	e.mu.Lock()
	e.activeProcesses[cmd.Process.Pid] = cmd
	e.mu.Unlock()

	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case <-ctx.Done():
		e.mu.Lock()
		delete(e.activeProcesses, cmd.Process.Pid)
		e.mu.Unlock()

		if err := e.terminateGracefully(cmd); err != nil {
			return PlotExecutionResult{}, fmt.Errorf("plot execution cancelled: %w (cleanup: %v)", ctx.Err(), err)
		}
		return PlotExecutionResult{}, fmt.Errorf("plot execution cancelled: %w", ctx.Err())

	case err := <-done:
		e.mu.Lock()
		delete(e.activeProcesses, cmd.Process.Pid)
		e.mu.Unlock()

		if err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				message := strings.TrimSpace(string(exitErr.Stderr))
				if message == "" {
					message = err.Error()
				}
				return PlotExecutionResult{}, fmt.Errorf("execute plot code: %s", message)
			}
			return PlotExecutionResult{}, fmt.Errorf("execute plot code: %w", err)
		}

		output, err := cmd.Output()
		if err != nil {
			message := strings.TrimSpace(string(output))
			if message == "" {
				message = err.Error()
			}
			return PlotExecutionResult{}, fmt.Errorf("read plot output: %s", message)
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
}

func (e *pythonPlotExecutor) terminateGracefully(cmd *exec.Cmd) error {
	if cmd == nil || cmd.Process == nil {
		return nil
	}

	if err := cmd.Process.Signal(syscall.SIGTERM); err != nil {
		if err := cmd.Process.Kill(); err != nil {
			return fmt.Errorf("failed to kill process: %w", err)
		}
		return nil
	}

	timeout := time.NewTimer(e.gracefulTimeout)
	defer timeout.Stop()

	done := make(chan error, 1)
	go func() {
		_, err := cmd.Process.Wait()
		done <- err
	}()

	select {
	case <-timeout.C:
		if err := cmd.Process.Kill(); err != nil {
			return fmt.Errorf("process did not exit gracefully and kill failed: %w", err)
		}
		return fmt.Errorf("process did not exit gracefully within %v, killed", e.gracefulTimeout)
	case <-done:
		return nil
	}
}

func (e *pythonPlotExecutor) Cleanup(ctx context.Context) error {
	var errs []error

	e.mu.Lock()
	for pid, cmd := range e.activeProcesses {
		if cmd != nil && cmd.Process != nil {
			if err := e.terminateGracefully(cmd); err != nil {
				errs = append(errs, fmt.Errorf("terminate process %d: %w", pid, err))
			}
		}
		delete(e.activeProcesses, pid)
	}
	e.mu.Unlock()

	e.tempFilesMu.Lock()
	for _, path := range e.tempFiles {
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			errs = append(errs, fmt.Errorf("remove temp file %s: %w", path, err))
		}
	}
	e.tempFiles = nil
	e.tempFilesMu.Unlock()

	if len(errs) > 0 {
		return fmt.Errorf("cleanup encountered %d error(s): %v", len(errs), errs)
	}
	return nil
}

func (e *pythonPlotExecutor) trackTempFile(path string) {
	e.tempFilesMu.Lock()
	e.tempFiles = append(e.tempFiles, path)
	e.tempFilesMu.Unlock()
}

func (e *pythonPlotExecutor) untrackTempFile(path string) {
	e.tempFilesMu.Lock()
	for i, p := range e.tempFiles {
		if p == path {
			e.tempFiles = append(e.tempFiles[:i], e.tempFiles[i+1:]...)
			break
		}
	}
	e.tempFilesMu.Unlock()
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
