package manager

import (
	"context"
	"errors"
	"time"

	"obfuscator/internal/anti/debug"
	"obfuscator/internal/anti/integrity"
	"obfuscator/internal/anti/vm"
)

// New создаёт фасад менеджера с заданными детекторами.
// Минимально инвазивно: логика порогов/режимов остаётся простой (на уровне агрегации),
// алгоритмика самих детекторов не изменяется.
func New(cfg Config, vmDetectors []vm.VMDetector, dbgDetectors []debug.DebuggerDetector, integ integrity.IntegrityChecker) Manager {
	return &managerImpl{
		cfg:          cfg,
		vmDetectors:  vmDetectors,
		dbgDetectors: dbgDetectors,
		integrity:    integ,
	}
}

type managerImpl struct {
	cfg          Config
	vmDetectors  []vm.VMDetector
	dbgDetectors []debug.DebuggerDetector
	integrity    integrity.IntegrityChecker
}

func (m *managerImpl) CheckVM(ctx context.Context) (float64, []string, error) {
	if !m.cfg.EnableVM {
		return 0.0, nil, nil
	}
	type res struct {
		score   float64
		signals []string
		err     error
	}
	var total float64
	var signals []string
	deadline := time.Now().Add(m.cfg.Debounce)
	for _, d := range m.vmDetectors {
		// простой последовательный вызов; debounce трактуем как soft-deadline
		if ctx.Err() != nil {
			return 0, signals, ctx.Err()
		}
		s, err := d.Score(ctx)
		if err != nil {
			// собираем первую ошибку, но не прерываем агрегацию
			if ctx.Err() != nil {
				return total, signals, ctx.Err()
			}
		}
		total += s
		signals = append(signals, d.Signals()...)
		if time.Now().After(deadline) {
			// мягкий выход, если вышли за debounce-тайм
			break
		}
	}
	return total, signals, nil
}

func (m *managerImpl) CheckDebugger(ctx context.Context) (bool, string, error) {
	if !m.cfg.EnableDebug {
		return false, "", nil
	}
	var detected bool
	var reason string
	var firstErr error
	for _, d := range m.dbgDetectors {
		ok, why, err := d.Detected(ctx)
		if err != nil && firstErr == nil {
			firstErr = err
		}
		if ok {
			detected = true
			if reason == "" {
				reason = d.Name() + ": " + why
			}
			// продолжаем, чтобы собрать потенциальные причины (минимально инвазивно можно вернуть первую)
			break
		}
	}
	return detected, reason, firstErr
}

func (m *managerImpl) CheckIntegrity(ctx context.Context) (bool, error) {
	if m.integrity == nil {
		return false, errors.New("integrity checker not configured")
	}
	return m.integrity.Verify(ctx)
}
