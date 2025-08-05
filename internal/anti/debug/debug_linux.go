//go:build linux
// +build linux

package debug

import (
	"context"
	"os"
	"syscall"
)

// linuxDebuggerDetector переносит смысл из pkg/obfuscator/anti_debug_linux.go:
// обнаружение отладки через ptrace Syscall; алгоритмика детектора не меняется.
type linuxDebuggerDetector struct{}

func NewDetector() DebuggerDetector { return &linuxDebuggerDetector{} }

func (d *linuxDebuggerDetector) Name() string { return "linux-debugger-detector" }

// Detected возвращает (true,"ptrace",nil) если ptrace-тест сигнализирует отладку.
// Иначе (false,"",nil). Ошибки не фатальны для поведения.
func (d *linuxDebuggerDetector) Detected(ctx context.Context) (bool, string, error) {
	if os.Getenv("OBF_DISABLE_ANTI_DEBUG") != "" {
		return false, "", nil
	}
	// ptrace check via syscall.Syscall(SYS_PTRACE, PTRACE_TRACEME, 0, 0)
	pres, _, _ := syscall.Syscall(syscall.SYS_PTRACE, uintptr(0), uintptr(0), uintptr(0))
	if pres != 0 {
		return true, "ptrace", nil
	}
	return false, "", nil
}
