package debug

import "context"

// DebuggerDetector — интерфейс детектора отладчика.
// Спецификация: Name() string; Detected(ctx) (bool, string, error).
type DebuggerDetector interface {
	Name() string
	Detected(ctx context.Context) (bool, string, error)
}
