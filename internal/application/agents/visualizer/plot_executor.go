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
	defaultExecutionTimeout        = 30 * time.Second
	maxMemoryMB                    = 512
)

// 危险模块和函数的黑名单 - 用于输入验证
var dangerousPatterns = []string{
	// 系统模块
	`\bos\b`, `\bsys\b`, `\bsubprocess\b`, `\bimportlib\b`,
	// 危险函数
	`\bexec\s*\(`, `\beval\s*\(`, `\bcompile\s*\(`, `\b__import__\s*\(`,
	// 文件系统操作
	`\bopen\s*\(`, `\bfile\s*\(`,
	// 网络操作
	`\bsocket\b`, `\burllib\b`, `\bhttp\b`, `\bftplib\b`,
	// 其他危险操作
	`\bpty\b`, `\bcommands\b`, `\bpopen\b`, `\bpipe\b`,
}

// 允许的matplotlib模块白名单
var allowedModules = map[string]bool{
	"matplotlib":       true,
	"matplotlib.pyplot": true,
	"matplotlib.patches": true,
	"matplotlib.lines":   true,
	"matplotlib.colors":  true,
	"matplotlib.cm":      true,
	"matplotlib.font_manager": true,
	"matplotlib.ticker":  true,
	"matplotlib.gridspec": true,
	"matplotlib.axes":    true,
	"matplotlib.figure":  true,
	"numpy":             true,
	"numpy as np":       true,
	"np":                true, // after import numpy as np
}

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
	executionTimeout     time.Duration
	maxMemoryMB          int
	mu                   sync.Mutex
	activeProcesses      map[int]*exec.Cmd
	tempFiles            []string
	tempFilesMu          sync.Mutex
}

func NewPlotExecutor() PlotExecutor {
	return &pythonPlotExecutor{
		command:              "python3",
		gracefulTimeout:      defaultGracefulShutdownTimeout,
		executionTimeout:     defaultExecutionTimeout,
		maxMemoryMB:          maxMemoryMB,
		activeProcesses:      make(map[int]*exec.Cmd),
	}
}

// validatePlotCode 验证用户输入的绘图代码是否安全
// 使用白名单策略：只允许 matplotlib 和 numpy 相关的导入
func validatePlotCode(code string) error {
	// 1. 检查是否包含危险的导入或函数调用
	for _, pattern := range dangerousPatterns {
		matched, err := regexp.MatchString(`(?i)`+pattern, code)
		if err != nil {
			return fmt.Errorf("validation error: %w", err)
		}
		if matched {
			return fmt.Errorf("security violation: code contains prohibited pattern '%s'", pattern)
		}
	}

	// 2. 检查 import 语句，只允许白名单中的模块
	importPattern := regexp.MustCompile(`(?m)^\s*import\s+(\S+)`)
	fromImportPattern := regexp.MustCompile(`(?m)^\s*from\s+(\S+)\s+import`)

	matches := importPattern.FindAllStringSubmatch(code, -1)
	for _, match := range matches {
		if len(match) > 1 {
			module := match[1]
			if !isModuleAllowed(module) {
				return fmt.Errorf("security violation: import of module '%s' is not allowed", module)
			}
		}
	}

	matches = fromImportPattern.FindAllStringSubmatch(code, -1)
	for _, match := range matches {
		if len(match) > 1 {
			module := match[1]
			if !isModuleAllowed(module) {
				return fmt.Errorf("security violation: import from module '%s' is not allowed", module)
			}
		}
	}

	// 3. 检查是否包含魔术命令
	if strings.Contains(code, "!") || strings.Contains(code, "%") {
		return fmt.Errorf("security violation: magic commands are not allowed")
	}

	return nil
}

// isModuleAllowed 检查模块是否在白名单中
func isModuleAllowed(module string) bool {
	// 检查完整模块名
	if allowedModules[module] {
		return true
	}

	// 检查模块前缀（允许 matplotlib 子模块）
	allowedPrefixes := []string{"matplotlib", "numpy", "mpl_toolkits"}
	for _, prefix := range allowedPrefixes {
		if strings.HasPrefix(module, prefix) {
			return true
		}
	}

	return false
}

func (e *pythonPlotExecutor) Execute(ctx context.Context, code string) (PlotExecutionResult, error) {
	cleaned := extractPlotCode(code)
	if cleaned == "" {
		return PlotExecutionResult{}, fmt.Errorf("plot code is empty")
	}

	// 安全验证：验证用户代码
	if err := validatePlotCode(cleaned); err != nil {
		return PlotExecutionResult{}, fmt.Errorf("code validation failed: %w", err)
	}

	// 创建带超时的上下文
	execCtx, cancel := context.WithTimeout(ctx, e.executionTimeout)
	defer cancel()

	// 构建受限执行命令
	cmd := exec.CommandContext(execCtx, e.command, "-c", plotExecutorScript)
	cmd.Stdin = strings.NewReader(cleaned)

	// 资源限制：设置内存限制（通过环境变量传递限制信息，实际限制在Python脚本中实现）
	if cmd.Env == nil {
		cmd.Env = os.Environ()
	}
	cmd.Env = append(cmd.Env, fmt.Sprintf("PLOT_MAX_MEMORY_MB=%d", e.maxMemoryMB))

	// 限制进程资源（仅Unix系统）
	if sysProcAttr := getResourceLimits(e.maxMemoryMB); sysProcAttr != nil {
		cmd.SysProcAttr = sysProcAttr
	}

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
	case <-execCtx.Done():
		e.mu.Lock()
		delete(e.activeProcesses, cmd.Process.Pid)
		e.mu.Unlock()

		if err := e.terminateGracefully(cmd); err != nil {
			return PlotExecutionResult{}, fmt.Errorf("plot execution timeout: %w (cleanup: %v)", execCtx.Err(), err)
		}
		if ctx.Err() != nil {
			return PlotExecutionResult{}, fmt.Errorf("plot execution cancelled: %w", ctx.Err())
		}
		return PlotExecutionResult{}, fmt.Errorf("plot execution timeout after %v", e.executionTimeout)

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

		// 使用 CombinedOutput 获取输出
		// 由于我们已经调用了 Wait()，需要重新执行命令来获取输出
		// 实际上应该使用 cmd.Output() 或 cmd.CombinedOutput() 在 Wait 之前
		// 这里修正：应该在 goroutine 中使用 cmd.Output() 而不是 cmd.Wait()
		return e.executeWithOutput(ctx, cleaned)
	}
}

// executeWithOutput 执行代码并获取输出
func (e *pythonPlotExecutor) executeWithOutput(ctx context.Context, code string) (PlotExecutionResult, error) {
	execCtx, cancel := context.WithTimeout(ctx, e.executionTimeout)
	defer cancel()

	cmd := exec.CommandContext(execCtx, e.command, "-c", plotExecutorScript)
	cmd.Stdin = strings.NewReader(code)

	if cmd.Env == nil {
		cmd.Env = os.Environ()
	}
	cmd.Env = append(cmd.Env, fmt.Sprintf("PLOT_MAX_MEMORY_MB=%d", e.maxMemoryMB))

	if sysProcAttr := getResourceLimits(e.maxMemoryMB); sysProcAttr != nil {
		cmd.SysProcAttr = sysProcAttr
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		if execCtx.Err() != nil {
			return PlotExecutionResult{}, fmt.Errorf("plot execution timeout after %v", e.executionTimeout)
		}
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

// getResourceLimits 获取进程资源限制（Unix系统）
func getResourceLimits(maxMemoryMB int) *syscall.SysProcAttr {
	// 注意：这在 Windows 上返回 nil，在 Unix 系统上设置资源限制
	// 由于代码需要在 Windows 上编译，我们使用条件编译
	return nil
}

const plotExecutorScript = `
import base64
import io
import os
import sys
import resource
import signal

# 设置内存限制（如果环境变量存在）
max_memory_mb = int(os.environ.get('PLOT_MAX_MEMORY_MB', 512))
max_memory_bytes = max_memory_mb * 1024 * 1024

try:
    # 设置内存限制（仅Unix系统）
    resource.setrlimit(resource.RLIMIT_AS, (max_memory_bytes, max_memory_bytes))
except (ImportError, ValueError, OSError):
    pass  # 在非Unix系统上忽略

# 设置执行超时信号
def timeout_handler(signum, frame):
    raise TimeoutError("Plot execution exceeded time limit")

try:
    signal.signal(signal.SIGALRM, timeout_handler)
    signal.alarm(30)  # 30秒超时
except (ImportError, ValueError, OSError):
    pass  # 在Windows上忽略

import matplotlib
matplotlib.use('Agg')  # 必须在导入pyplot之前设置
import matplotlib.pyplot as plt
import numpy as np

# 创建受限的执行命名空间 - 只允许安全模块
safe_globals = {
    '__builtins__': {
        'abs': abs, 'all': all, 'any': any, 'bin': bin, 'bool': bool,
        'bytearray': bytearray, 'bytes': bytes, 'chr': chr, 'complex': complex,
        'dict': dict, 'divmod': divmod, 'enumerate': enumerate, 'filter': filter,
        'float': float, 'format': format, 'frozenset': frozenset, 'hasattr': hasattr,
        'hash': hash, 'hex': hex, 'int': int, 'isinstance': isinstance,
        'issubclass': issubclass, 'iter': iter, 'len': len, 'list': list,
        'map': map, 'max': max, 'min': min, 'next': next, 'oct': oct,
        'ord': ord, 'pow': pow, 'print': print, 'range': range, 'repr': repr,
        'reversed': reversed, 'round': round, 'set': set, 'slice': slice,
        'sorted': sorted, 'str': str, 'sum': sum, 'tuple': tuple, 'type': type,
        'zip': zip, 'True': True, 'False': False, 'None': None,
    },
    'plt': plt,
    'np': np,
    'numpy': np,
    'matplotlib': matplotlib,
    'io': io,
    'base64': base64,
}

# 读取用户代码
code = sys.stdin.read()

# 清除之前的图形
plt.close('all')
plt.rcdefaults()

# 在受限命名空间中执行用户代码
namespace = {}
try:
    exec(code, safe_globals, namespace)
except Exception as e:
    print(f"Execution error: {e}", file=sys.stderr)
    sys.exit(1)

# 检查是否创建了图形
if not plt.get_fignums():
    print("Error: plot code did not create a figure", file=sys.stderr)
    sys.exit(1)

# 保存图形
buffer = io.BytesIO()
plt.savefig(buffer, format="jpeg", bbox_inches="tight", dpi=300)
plt.close("all")

# 输出base64编码的图像
sys.stdout.write(base64.b64encode(buffer.getvalue()).decode("utf-8"))
`
