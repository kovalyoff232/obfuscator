//go:build !linux && !windows
// +build !linux,!windows

package debug

import "context"

// noopDebuggerDetector — заглушка для не-Linux и не-Windows платформ.
type noopDebuggerDetector struct{}

func NewDetector() DebuggerDetector { return &noopDebuggerDetector{} }

func (d *noopDebuggerDetector) Name() string { return "noop-debugger-detector" }
func (d *noopDebuggerDetector) Detected(ctx context.Context) (bool, string, error) {
	return false, "", nil
}
