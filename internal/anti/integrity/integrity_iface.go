package integrity

import "context"

// IntegrityChecker — интерфейс проверки целостности.
// Спецификация: Name() string; Mode() string; Verify(ctx) (bool, error).
type IntegrityChecker interface {
	Name() string
	Mode() string
	Verify(ctx context.Context) (bool, error)
}
