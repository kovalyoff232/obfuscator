//go:build windows
// +build windows

package vm

import (
	"context"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"unsafe"
)

// windowsVMDetector — эвристика VM по BIOS данным из реестра и fallback на wmic.
type windowsVMDetector struct{}

func NewDetector() VMDetector { return &windowsVMDetector{} }

func (d *windowsVMDetector) Name() string { return "windows-vm-detector" }

func (d *windowsVMDetector) Signals() []string {
	return []string{
		"reg:HKLM\\HARDWARE\\DESCRIPTION\\System\\BIOS:SystemManufacturer",
		"reg:HKLM\\HARDWARE\\DESCRIPTION\\System\\BIOS:SystemProductName",
		"wmic:csproduct vendor",
		"wmic:computersystem manufacturer",
	}
}

// Score возвращает 1.0 при наличии признаков виртуалки, иначе 0.0.
// Ошибки чтения не фатальны — деградация к 0.0.
func (d *windowsVMDetector) Score(ctx context.Context) (float64, error) {
	if os.Getenv("OBF_DISABLE_ANTI_VM") != "" {
		return 0.0, nil
	}

	vmVendors := []string{"VMWARE", "VIRTUALBOX", "KVM", "HYPER-V", "QEMU", "XEN", "MICROSOFT CORPORATION"} // Hyper-V часто как Microsoft Corporation
	check := func(s string) bool {
		u := strings.ToUpper(s)
		for _, v := range vmVendors {
			if strings.Contains(u, v) {
				return true
			}
		}
		return false
	}

	// 1) Попытка прочитать из реестра HKLM\HARDWARE\DESCRIPTION\System\BIOS
	if manu, ok := readRegistryString(syscall.HKEY_LOCAL_MACHINE, `HARDWARE\DESCRIPTION\System\BIOS`, "SystemManufacturer"); ok && check(manu) {
		return 1.0, nil
	}
	if prod, ok := readRegistryString(syscall.HKEY_LOCAL_MACHINE, `HARDWARE\DESCRIPTION\System\BIOS`, "SystemProductName"); ok && check(prod) {
		return 1.0, nil
	}

	// 2) Fallback: wmic без внешних зависимостей (через стандартный exec)
	//	- wmic csproduct get vendor
	//	- wmic computersystem get manufacturer
	if out, err := exec.Command("wmic", "csproduct", "get", "vendor").CombinedOutput(); err == nil {
		if check(string(out)) {
			return 1.0, nil
		}
	}
	if out, err := exec.Command("wmic", "computersystem", "get", "manufacturer").CombinedOutput(); err == nil {
		if check(string(out)) {
			return 1.0, nil
		}
	}

	return 0.0, nil
}

// readRegistryString читает строковое значение реестра через syscall (без внешних пакетов).
func readRegistryString(root syscall.Handle, path, name string) (string, bool) {
	var h syscall.Handle
	pathPtr, _ := syscall.UTF16PtrFromString(path)
	if err := syscall.RegOpenKeyEx(root, pathPtr, 0, syscall.KEY_READ, &h); err != nil {
		return "", false
	}
	defer syscall.RegCloseKey(h)

	namePtr, _ := syscall.UTF16PtrFromString(name)
	var valType uint32
	var bufLen uint32
	// Первый вызов — получить длину
	if r := regQueryValueEx(h, namePtr, &valType, nil, &bufLen); r != 0 {
		return "", false
	}
	if valType != syscall.REG_SZ && valType != syscall.REG_EXPAND_SZ {
		return "", false
	}
	if bufLen == 0 || bufLen > 1<<16 {
		return "", false
	}
	// bufLen — в байтах
	u16len := int(bufLen / 2)
	if u16len <= 0 {
		return "", false
	}
	buf := make([]uint16, u16len)
	if r := regQueryValueEx(h, namePtr, &valType, (*byte)(unsafe.Pointer(&buf[0])), &bufLen); r != 0 {
		return "", false
	}
	// Обрезаем по нулевому терминатору
	s := syscall.UTF16ToString(buf)
	return s, true
}

// обёртка RegQueryValueExW
func regQueryValueEx(h syscall.Handle, lpValueName *uint16, lpType *uint32, lpData *byte, lpcbData *uint32) syscall.Errno {
	r1, _, e1 := syscall.Syscall6(procRegQueryValueExW.Addr(), 6,
		uintptr(h),
		uintptr(unsafe.Pointer(lpValueName)),
		0,
		uintptr(unsafe.Pointer(lpType)),
		uintptr(unsafe.Pointer(lpData)),
		uintptr(unsafe.Pointer(lpcbData)),
	)
	if r1 != 0 {
		return syscall.Errno(r1)
	}
	if e1 != 0 {
		return e1
	}
	return 0
}

var (
	modAdvapi32          = syscall.NewLazyDLL("advapi32.dll")
	procRegQueryValueExW = modAdvapi32.NewProc("RegQueryValueExW")
)
