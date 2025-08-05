//go:build linux
// +build linux

package vm

import (
	"context"
	"errors"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// linuxVMDetector переносит логику из pkg/obfuscator/anti_vm_linux.go (алгоритмика не меняется),
// но в форме детектора: суммарный скор и список сигналов.
type linuxVMDetector struct{}

func NewDetector() VMDetector { return &linuxVMDetector{} }

func (d *linuxVMDetector) Name() string { return "linux-vm-detector" }

func (d *linuxVMDetector) Signals() []string {
	// Список строковых индикаторов, как в исходной реализации
	return []string{
		"mac:00:05:69", "mac:00:0c:29", "mac:00:50:56",
		"mac:08:00:27",
		"mac:00:1c:42",
		"mac:00:16:3e",
		"dmi:product_name", "dmi:sys_vendor", "dmi:board_vendor", "dmi:board_name",
	}
}

func (d *linuxVMDetector) Score(ctx context.Context) (float64, error) {
	// Поведение должно сохраниться: проверяем MAC-префиксы и DMI содержимое строк.
	// Возвращаем суммарный «1» при первых найденных признаках (минимально: 0 или 1).
	if os.Getenv("OBF_DISABLE_ANTI_VM") != "" {
		return 0, nil
	}
	// MAC prefixes
	vmPrefixes := []string{
		"00:05:69", "00:0c:29", "00:50:56",
		"08:00:27",
		"00:1c:42",
		"00:16:3e",
	}
	ifs, err := net.Interfaces()
	if err == nil {
		for _, i := range ifs {
			mac := i.HardwareAddr.String()
			for _, p := range vmPrefixes {
				if strings.HasPrefix(mac, p) {
					return 1.0, nil
				}
			}
		}
	}
	// DMI (Linux only)
	if runtime.GOOS == "linux" {
		dmiPath := "/sys/class/dmi/id/"
		dmiFiles := []string{"product_name", "sys_vendor", "board_vendor", "board_name"}
		vmStrings := []string{"VMware", "VirtualBox", "QEMU", "KVM", "Xen"}
		for _, f := range dmiFiles {
			content, _ := os.ReadFile(filepath.Join(dmiPath, f))
			for _, s := range vmStrings {
				if strings.Contains(string(content), s) {
					return 1.0, nil
				}
			}
		}
	}
	return 0.0, nil
}

// ensure _ = errors to keep parity with imports if needed later
var _ = errors.New
