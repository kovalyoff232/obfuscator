//go:build !linux
// +build !linux

package vm

import "context"

// noopVMDetector — заглушка для не-Linux платформ.
type noopVMDetector struct{}

func NewNoopDetector() VMDetector { return &noopVMDetector{} } // заглушка для не-Linux

func (d *noopVMDetector) Name() string                               { return "noop-vm-detector" }
func (d *noopVMDetector) Score(ctx context.Context) (float64, error) { return 0.0, nil }
func (d *noopVMDetector) Signals() []string                          { return nil }
