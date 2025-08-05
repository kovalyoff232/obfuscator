//go:build windows
// +build windows

package debug

import (
	"context"
	"os"
	"syscall"
	"unsafe"
)

// windowsDebuggerDetector — использует kernel32!IsDebuggerPresent и kernel32!CheckRemoteDebuggerPresent.
type windowsDebuggerDetector struct{}

func NewDetector() DebuggerDetector { return &windowsDebuggerDetector{} }

func (d *windowsDebuggerDetector) Name() string { return "windows-debugger-detector" }

func (d *windowsDebuggerDetector) Detected(ctx context.Context) (bool, string, error) {
	if os.Getenv("OBF_DISABLE_ANTI_DEBUG") != "" {
		return false, "", nil
	}

	// 1) IsDebuggerPresent
	if r1, _, _ := procIsDebuggerPresent.Call(); r1 != 0 {
		return true, "IsDebuggerPresent", nil
	}

	// 2) CheckRemoteDebuggerPresent(GetCurrentProcess(), &present)
	hProc, _, _ := procGetCurrentProcess.Call()
	var present uint32
	if _, _, _ = procCheckRemoteDebuggerPresent.Call(
		hProc,
		uintptr(unsafe.Pointer(&present)),
	); present != 0 {
		return true, "CheckRemoteDebuggerPresent", nil
	}

	return false, "", nil
}

var (
	modKernel32                    = syscall.NewLazyDLL("kernel32.dll")
	procIsDebuggerPresent          = modKernel32.NewProc("IsDebuggerPresent")
	procCheckRemoteDebuggerPresent = modKernel32.NewProc("CheckRemoteDebuggerPresent")
	procGetCurrentProcess          = modKernel32.NewProc("GetCurrentProcess")
)
