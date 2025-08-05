package integrity

import (
	"context"
	"fmt"
)

// Wrapper — платф.-нейтральная обертка над текущей концепцией integrity weaving,
// реализует интерфейс IntegrityChecker без изменения алгоритмики исходного проекта.
// Mode() возвращает режим из конфигурации (пока фиксированный "weaving").
type Wrapper struct {
	name string
	mode string
}

func NewWrapper(mode string) *Wrapper {
	return &Wrapper{
		name: "integrity-weaving",
		mode: mode,
	}
}

func (w *Wrapper) Name() string { return w.name }

func (w *Wrapper) Mode() string { return w.mode }

// Verify — адаптер. В текущем проекте «weaving» выполняется на этапе обфускации,
// а рантайм-проверки отсутствуют. Для соблюдения интерфейса возвращаем (true, nil),
// не изменяя поведение.
func (w *Wrapper) Verify(ctx context.Context) (bool, error) {
	select {
	case <-ctx.Done():
		return false, ctx.Err()
	default:
	}
	return true, nil
}

func (w *Wrapper) String() string {
	return fmt.Sprintf("%s[%s]", w.name, w.mode)
}
