package manager

import (
	"context"
	"time"
)

// Config по спецификации анти-аналитики.
// Поля: VMThreshold float64; EnableVM bool; EnableDebug bool; IntegrityMode string; Debounce time.Duration; Whitelist []string.
type Config struct {
	VMThreshold   float64
	EnableVM      bool
	EnableDebug   bool
	IntegrityMode string
	Debounce      time.Duration
	Whitelist     []string
}

// Manager интерфейс фасада по спецификации.
// Методы: CheckVM(ctx) (float64, []string, error); CheckDebugger(ctx) (bool, string, error); CheckIntegrity(ctx) (bool, error).
type Manager interface {
	CheckVM(ctx context.Context) (float64, []string, error)
	CheckDebugger(ctx context.Context) (bool, string, error)
	CheckIntegrity(ctx context.Context) (bool, error)
}
