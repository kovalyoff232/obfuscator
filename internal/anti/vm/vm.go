package vm

import (
	"context"
)

// VMDetector — интерфейс детектора виртуальной среды.
// Спецификация: Name() string; Score(ctx) (float64, error); Signals() []string.
type VMDetector interface {
	Name() string
	Score(ctx context.Context) (float64, error)
	Signals() []string
}
