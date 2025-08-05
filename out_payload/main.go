package main

import (
	"crypto/subtle"
	"runtime"
	"path/filepath"
	"strings"
	"net"
	"syscall"
	"crypto/cipher"
	"crypto/aes"
	"fmt"
	"os"
)

var o_DAeEyFXJKh = map[string]string{"o_NipuFYYTav": "79f0109a91576cede1ff27f02347f5cd4f5359557a361ef1909bb23aa961e77f", "init": "daa76c3ff8609be2885d251060b577420e8306302850e01e0357c69d0f4dbfb6", "init_1": "64ea97d6d09be4840256799a1c2ba27d27b8c6c8b55c4be71ca09b17c337044c", "getPassword": "3abba610961fe6f1896760a7f6e0e16316dbe1411fc4ec2db7baa7c0cff6e8e8", "checkPassword": "4da7b7b1e924389d803e2c3cde214a886e50a366844c686be4ebd5654c6d0188", "main": "859516d4e1271fa5719039dcdf7bec77957a42bdbe2e208d457df8cdbf6c2b19"}

func o_NipuFYYTav(obfuscatedID int, args ...interface{}) interface{} {
	localKey := 24185 ^ int(o_UTUZEUzpcJ)
	id := obfuscatedID ^ localKey
	switch id {
	case 1:
		return getPassword()
	case 2:
		return checkPassword(args[0].(string))
	default:
		panic("unknown function id in dispatcher")
	}
	return nil
}

var o_QJaGXcQmUa int = 0

func init() {
	if os.Getenv("OBF_DISABLE_ANTI_VM") != "" {
		goto done
	}
	pfx := []string{"00:05:69", "00:0c:29", "00:50:56", "08:00:27", "00:1c:42", "00:16:3e"}
	ifs, err := net.Interfaces()
	if err == nil {
		for _, i := range ifs {
			mac := i.HardwareAddr.String()
			for _, p := range pfx {
				if strings.HasPrefix(mac, p) {
					o_QJaGXcQmUa = 1
					goto done
				}
			}
		}
	}
	if runtime.GOOS == "linux" {
		dmiPath := "/sys/class/dmi/id/"
		dmiFiles := []string{"product_name", "sys_vendor", "board_vendor", "board_name"}
		vmStrings := []string{"VMware", "VirtualBox", "QEMU", "KVM", "Xen"}
		for _, f := range dmiFiles {
			content, _ := os.ReadFile(filepath.Join(dmiPath, f))
			for _, s := range vmStrings {
				if strings.Contains(string(content), s) {
					o_QJaGXcQmUa = 1
					goto done
				}
			}
		}
	}
done:
}

var o_UTUZEUzpcJ int64

func init() {
done:
	;
	var ptraceComponent int64 = 0
	if os.Getenv("OBF_DISABLE_ANTI_DEBUG") != "" {
		goto done
	}
	if pres, _, _ := syscall.Syscall(syscall.SYS_PTRACE, uintptr(0), uintptr(0), uintptr(0)); pres != 0 {
		ptraceComponent = 1337
	}
	o_UTUZEUzpcJ = ptraceComponent + 0*((9999^576)^576)
}
func getPassword() []byte {
	var o_EoAOIcBLYV string
	var o_QQdlWlveur []byte
	if o_TPsxfIoqeb := 123; o_TPsxfIoqeb*o_TPsxfIoqeb-1 == (o_TPsxfIoqeb-1)*(o_TPsxfIoqeb+1) {
		o_JoyPuTituB := "junk"
		_ = o_JoyPuTituB
	}
	var o_MFgKvULnAs = 0
	for {
		if o_MFgKvULnAs == 2 {
			break
		}
		switch o_MFgKvULnAs {
		case 1:

			fmt.Println(func() string {
				o_PkLSZwkShA := []byte{0x99, 0xc6, 0xd8, 0xeb, 0x11, 0x5b, 0x30, 0x8c, 0x9, 0x5b, 0xea, 0xc0, 0x88, 0x67, 0x7d, 0x7b}
				o_MqFIcCcqXx := []byte{0x6e, 0xc3, 0x76, 0xdf, 0x5e, 0x8e, 0x61, 0xe0, 0xe7, 0x3b, 0x18, 0x53, 0xe7, 0xf4, 0xc3, 0xcd}
				o_DCZXKdXuQX := []byte{0x87, 0x36, 0x9f, 0xe5, 0xf9, 0x8c, 0xd3, 0xd0, 0xd4, 0x14, 0xed, 0xfb, 0x96, 0xc, 0x31, 0xdd, 0x4a, 0x6f, 0x4f, 0x67, 0x2c, 0x92, 0x70, 0xdb, 0xa0, 0x21, 0x4d, 0xe4, 0x99, 0x92, 0x85, 0x43, 0xe9, 0xac, 0x18, 0x3a, 0xcd, 0xcd, 0xc9, 0x60, 0x7, 0xac, 0xbf, 0x73, 0xf5, 0xe2, 0x4f, 0x10, 0x4, 0xff, 0x7b, 0x9f, 0x0, 0x61, 0xa3, 0xb1, 0x43, 0x9e, 0x57, 0x72, 0x9b, 0xd8, 0xf3, 0xa0, 0x4e, 0xae, 0xff, 0xb9, 0x94, 0x44, 0x6a, 0x87, 0x11, 0x75, 0xa2, 0x2c, 0x2a, 0xf4, 0x64, 0x55, 0x91, 0xb1, 0x6, 0xb5, 0xe4, 0xcb, 0x45, 0x3c, 0xec, 0x46, 0x59, 0x7a, 0x37, 0xcd, 0xe1, 0x80, 0x92, 0x3d, 0x9c, 0x10, 0xf, 0x63, 0x14, 0x5a, 0xad, 0x23, 0x33, 0xe4, 0x5d, 0xd0, 0xb2, 0xd6, 0xb0, 0x97, 0xd7, 0x76, 0xf1, 0x8, 0x6d, 0xe8, 0x46, 0xd7, 0xa0}
				if o_LBTCbomLkc := 99; o_LBTCbomLkc*o_LBTCbomLkc+1 > 0 {
					o_PfPCgynGiU := "junk"
					_ = o_PfPCgynGiU
				}
				for o_lsnWfXwjvA := 0; o_lsnWfXwjvA < len(o_PkLSZwkShA); o_lsnWfXwjvA++ {
					o_PkLSZwkShA[o_lsnWfXwjvA] ^= byte((uint64(o_UTUZEUzpcJ)>>uint((o_lsnWfXwjvA%8*((8+545)-545))) ^ uint64(byte(o_lsnWfXwjvA*((31+861)-861))) ^ uint64(byte(len(o_DCZXKdXuQX)*17))) & ((255 ^ 866) ^ 866))
				}
				if len(o_MqFIcCcqXx) > 0 {
					o_MqFIcCcqXx[0] ^= byte(len(o_DCZXKdXuQX))
				}
				o_LPHUDfLLyz, o_NGzvmmhaGT := aes.NewCipher(o_PkLSZwkShA)
				if o_NGzvmmhaGT != nil {
					panic(o_NGzvmmhaGT)
				}
				if o_RQKrmbiJip := (99 ^ 496) ^ 496; o_RQKrmbiJip*o_RQKrmbiJip+1 > 0 {
					o_BtSVAPbNwb := "junk"
					_ = o_BtSVAPbNwb
				}
				o_yXHFDbpaQX := make([]byte, len(o_DCZXKdXuQX))
				o_LGHSpuDRVp := cipher.NewCTR(o_LPHUDfLLyz, o_MqFIcCcqXx)
				o_LGHSpuDRVp.XORKeyStream(o_yXHFDbpaQX, o_DCZXKdXuQX)
				return string(o_yXHFDbpaQX)
			}())
			o_QQdlWlveur = []byte("")
			o_MFgKvULnAs = 2
		case 0:
			if o_EoAOIcBLYV = os.Getenv(func() string {
				o_JishJyqYaZ := []byte{0x38, 0x20, 0x72, 0x2d, 0xc8, 0xb0, 0x84, 0xe7, 0x89, 0x84, 0x22, 0x7a, 0x84, 0x53, 0xcc, 0xcf}
				o_uOPynVAniK := []byte{0xf2, 0xdb, 0xcd, 0x75, 0xb3, 0x94, 0x77, 0xf1, 0x2, 0x8c, 0x24, 0xd4, 0xca, 0x92, 0xf4, 0xe0}
				o_QHyFhgXqbq := []byte{0x6b, 0x83, 0x9e, 0x92, 0x2d, 0xf5, 0x5a, 0x8c, 0xaf, 0x7c, 0xc4, 0xcf, 0x54, 0xfb, 0x2d, 0xdf}
				if o_PgozWYZCdX := (99 ^ 716) ^ 716; o_PgozWYZCdX*o_PgozWYZCdX+1 > 0 {
					o_NWFWIsLQdZ := "junk"
					_ = o_NWFWIsLQdZ
				}
				for o_Oqqmaonklj := 0; o_Oqqmaonklj < len(o_uOPynVAniK); o_Oqqmaonklj++ {
					o_uOPynVAniK[o_Oqqmaonklj] ^= byte((uint64(o_UTUZEUzpcJ)>>uint((o_Oqqmaonklj%8*8)) ^ uint64(byte(o_Oqqmaonklj*((31^198)^198))) ^ uint64(byte(len(o_QHyFhgXqbq)*((17^436)^436)))) & 0xff)
				}
				if len(o_JishJyqYaZ) > 0 {
					o_JishJyqYaZ[0] ^= byte(len(o_QHyFhgXqbq))
				}
				o_WkyXiqbvxV, o_gPUCDdkXiP := aes.NewCipher(o_uOPynVAniK)
				if o_gPUCDdkXiP != nil {
					panic(o_gPUCDdkXiP)
				}
				if o_EctBAyTaXt := (99 ^ 20) ^ 20; o_EctBAyTaXt*o_EctBAyTaXt+1 > 0 {
					o_vLjIfxiYTg := "junk"
					_ = o_vLjIfxiYTg
				}
				o_IRofvopKPs := make([]byte, len(o_QHyFhgXqbq))
				o_HFNJQBGxwm := cipher.NewCTR(o_WkyXiqbvxV, o_JishJyqYaZ)
				o_HFNJQBGxwm.XORKeyStream(o_IRofvopKPs, o_QHyFhgXqbq)
				return string(o_IRofvopKPs)
			}()); o_EoAOIcBLYV != "" {
				return []byte(o_EoAOIcBLYV)
			}
			o_MFgKvULnAs = 1
		case 2:
			o_MfMPylWfab := 123
			o_nwowCqMVtg := 456
			if 7*o_MfMPylWfab-3*o_MfMPylWfab-4*o_MfMPylWfab != 0 {
				panic("unreachable")
			}
			_ = o_nwowCqMVtg
		case 3:
			o_QUiWEFZmWv := 123
			o_sFnAVdmOxI := (456 ^ 534) ^ 534
			if (o_QUiWEFZmWv-o_sFnAVdmOxI)*(o_QUiWEFZmWv+o_sFnAVdmOxI) != o_QUiWEFZmWv*o_QUiWEFZmWv-o_sFnAVdmOxI*o_sFnAVdmOxI {
				panic("unreachable")
			}
			_ = o_sFnAVdmOxI
		case (4 + 559) - 559:
			o_TpTWJEUpyp := 123
			o_dxZGtcWwyP := (456 ^ 34) ^ 34
			if o_TpTWJEUpyp*o_dxZGtcWwyP+1 == o_TpTWJEUpyp*o_dxZGtcWwyP {
				panic("unreachable")
			}
			_ = o_dxZGtcWwyP
		case 5:
			o_UVIvwdyVWe := 123
			o_PfQAKgILzp := 456
			if ((o_UVIvwdyVWe ^ o_PfQAKgILzp) ^ o_PfQAKgILzp) != o_UVIvwdyVWe {
				panic("unreachable")
			}
			_ = o_PfQAKgILzp
		case 6:
			o_mHyKVwMrKM := 123
			o_qKZQxxzjCz := 456
			if (o_mHyKVwMrKM-o_qKZQxxzjCz)*(o_mHyKVwMrKM+o_qKZQxxzjCz) != o_mHyKVwMrKM*o_mHyKVwMrKM-o_qKZQxxzjCz*o_qKZQxxzjCz {
				panic("unreachable")
			}
			_ = o_qKZQxxzjCz
		case 7:
			o_PcZZpnNlFb := 123
			o_iskcrrAtnJ := 456
			if (o_PcZZpnNlFb-o_iskcrrAtnJ)*(o_PcZZpnNlFb+o_iskcrrAtnJ) != o_PcZZpnNlFb*o_PcZZpnNlFb-o_iskcrrAtnJ*o_iskcrrAtnJ {
				panic("unreachable")
			}
			_ = o_iskcrrAtnJ
		case 8:
			o_NaPNgHsOOd := 123
			o_jWQnagrJDM := 456
			if ((o_NaPNgHsOOd ^ o_jWQnagrJDM) ^ o_jWQnagrJDM) != o_NaPNgHsOOd {
				panic("unreachable")
			}
			_ = o_jWQnagrJDM
		}
	}
	return o_QQdlWlveur
}

func checkPassword(o_SCTpTDFRUO string) bool {
	var o_WffOLNpjlc []byte
	var o_APTAKxsUKw []byte
	var o_MwBbAYOSSJ []byte
	var o_ViggFfamrX bool
	var o_RCnRPjaNjZ = 0
	for {
		if o_RCnRPjaNjZ == 2 {
			break
		}
		switch o_RCnRPjaNjZ {
		case 0:
			o_WffOLNpjlc = o_NipuFYYTav(24184).([]byte)

			o_APTAKxsUKw = []byte(o_SCTpTDFRUO)

			if len(o_WffOLNpjlc) != len(o_APTAKxsUKw) {

				o_MwBbAYOSSJ = make([]byte, len(o_WffOLNpjlc))
				copy(o_MwBbAYOSSJ, o_APTAKxsUKw)
				return subtle.ConstantTimeCompare(o_WffOLNpjlc, o_MwBbAYOSSJ) == 1
			}
			o_RCnRPjaNjZ = 1
		case 1:
			o_ViggFfamrX = subtle.ConstantTimeCompare(o_WffOLNpjlc, o_APTAKxsUKw) == 1
			o_RCnRPjaNjZ = 2
		case 2:
			o_htNiQeojmS := 123
			o_tCPCuPENqJ := (456 ^ 558) ^ 558
			if ((7^993)^993)*o_htNiQeojmS-3*o_htNiQeojmS-((4+406)-406)*o_htNiQeojmS != 0 {
				panic("unreachable")
			}
			_ = o_tCPCuPENqJ
		case (3 ^ 297) ^ 297:
			o_SxjPSqEngj := 123
			o_BqbGjzJjBx := 456
			if o_SxjPSqEngj*o_BqbGjzJjBx+1 == o_SxjPSqEngj*o_BqbGjzJjBx {
				panic("unreachable")
			}
			_ = o_BqbGjzJjBx
		case 4:
			o_KBWRgupYvR := (123 ^ 488) ^ 488
			o_uQeTzVdoDg := (456 + 631) - 631
			if o_KBWRgupYvR*o_KBWRgupYvR+1 < 0 {
				panic("unreachable")
			}
			_ = o_uQeTzVdoDg
		case 5:
			o_EPZvAPxQPI := 123
			o_sjwxqRcESV := 456
			if ((o_EPZvAPxQPI ^ o_sjwxqRcESV) ^ o_sjwxqRcESV) != o_EPZvAPxQPI {
				panic("unreachable")
			}
			_ = o_sjwxqRcESV
		case (6 + 482) - 482:
			o_fPsJDIohfo := 123
			o_VpfMQMfJQI := 456
			if (o_fPsJDIohfo-o_VpfMQMfJQI)*(o_fPsJDIohfo+o_VpfMQMfJQI) != o_fPsJDIohfo*o_fPsJDIohfo-o_VpfMQMfJQI*o_VpfMQMfJQI {
				panic("unreachable")
			}
			_ = o_VpfMQMfJQI
		case 7:
			o_jqKcwFAuYX := 123
			o_KIwGoXZbeh := (456 + 358) - 358
			if ((o_jqKcwFAuYX ^ o_KIwGoXZbeh) ^ o_KIwGoXZbeh) != o_jqKcwFAuYX {
				panic("unreachable")
			}
			_ = o_KIwGoXZbeh
		case 8:
			o_EKKaxPiBgf := 123
			o_vXszwVJrwR := 456
			if o_EKKaxPiBgf*o_EKKaxPiBgf+1 < 0 {
				panic("unreachable")
			}
			_ = o_vXszwVJrwR
		}
	}
	return o_ViggFfamrX
}

func main() {
	fmt.Println(func() string {
		o_fmmUUNAuvj := []byte{0xb9, 0xd, 0xba, 0x6f, 0x1d, 0x1c, 0x88, 0xf5, 0x40, 0x96, 0x3f, 0x19, 0x6d, 0x56, 0x8d, 0xaa}
		o_PKaJQaEsHT := []byte{0xa5, 0x97, 0xe7, 0xb8, 0x2d, 0xe5, 0xca, 0x3b, 0xdb, 0x65, 0x52, 0xd, 0x74, 0x6d, 0xf1, 0x29, 0xc7, 0x60, 0xc6, 0xb3, 0x5a, 0x5e, 0xbe, 0xce, 0x2b, 0x49, 0xd1, 0x75, 0xdf, 0x67, 0xdb, 0xce, 0xe0, 0xf, 0xec, 0x50, 0xeb, 0xf7, 0xae, 0x23, 0x2b, 0x60, 0xef, 0x39, 0xbb}
		o_qxSrQLvOxA := []byte{0xdc, 0xca, 0xb8, 0x3b, 0x81, 0x8e, 0xaf, 0x9c, 0x3e, 0xdb, 0xbb, 0x50, 0x85, 0x3b, 0x40, 0x8d}
		o_iuIozHtDaS := 1337
		o_cuTguXQSHg := o_iuIozHtDaS * o_iuIozHtDaS
		o_BcArIylfIF := o_cuTguXQSHg ^ o_iuIozHtDaS
		_ = o_BcArIylfIF
		for o_vDDSNnITCW := 0; o_vDDSNnITCW < len(o_qxSrQLvOxA); o_vDDSNnITCW++ {
			o_qxSrQLvOxA[o_vDDSNnITCW] ^= byte((uint64(o_UTUZEUzpcJ)>>uint((o_vDDSNnITCW%((8+162)-162)*8)) ^ uint64(byte(o_vDDSNnITCW*((31+189)-189))) ^ uint64(byte(len(o_PKaJQaEsHT)*17))) & ((255 + 700) - 700))
		}
		if len(o_fmmUUNAuvj) > 0 {
			o_fmmUUNAuvj[0] ^= byte(len(o_PKaJQaEsHT))
		}
		o_vofzflPJSY, o_TrYUzlkYAg := aes.NewCipher(o_qxSrQLvOxA)
		if o_TrYUzlkYAg != nil {
			panic(o_TrYUzlkYAg)
		}
		o_sWfzMMfQGP := 1337
		o_HQrQEUUbhj := o_sWfzMMfQGP * o_sWfzMMfQGP
		o_ZYxQmYmKOf := o_HQrQEUUbhj ^ o_sWfzMMfQGP
		_ = o_ZYxQmYmKOf
		o_ZhjTNBCvbz := make([]byte, len(o_PKaJQaEsHT))
		o_pQVbfCsIgw := cipher.NewCTR(o_vofzflPJSY, o_fmmUUNAuvj)
		o_pQVbfCsIgw.XORKeyStream(o_ZhjTNBCvbz, o_PKaJQaEsHT)
		return string(o_ZhjTNBCvbz)
	}())
	fmt.Println(func() string {
		o_aBimeCKNqd := []byte{0x19, 0xe1, 0xdf, 0x21, 0x71, 0x37, 0x37, 0x70, 0x73, 0x5b, 0xc2, 0xe6, 0x63, 0xef, 0x8a, 0x67}
		o_vfOIatoQYx := []byte{0x7d, 0x9b, 0x3e, 0xc0, 0x7d, 0x37, 0xe2, 0x86, 0xb0, 0x63, 0xc2, 0x99, 0x82, 0xca, 0xca, 0x7e}
		o_MqdCNYOqjX := []byte{0xa3, 0x79, 0x40, 0xa1, 0xc9, 0xc, 0x84, 0x4b, 0xe1, 0x48, 0x9c, 0x45, 0xe0, 0x66, 0x8e, 0x43, 0x6b, 0xe2, 0x31, 0x4c, 0xfe, 0xbe, 0x8c, 0x5, 0x4e, 0x2d, 0xc1, 0x17, 0x72, 0xe0, 0xf0, 0x59, 0x6b, 0x22, 0x6a, 0x3, 0xad, 0x8f, 0xf, 0x7e, 0x3f, 0xf1, 0xb8, 0xb8, 0xb6}
		o_roPnACNJGt := 1337
		o_DuLLsWfGWV := o_roPnACNJGt * o_roPnACNJGt
		o_AsMCbGyIMz := o_DuLLsWfGWV ^ o_roPnACNJGt
		_ = o_AsMCbGyIMz
		for o_TieJUJsxJS := 0; o_TieJUJsxJS < len(o_vfOIatoQYx); o_TieJUJsxJS++ {
			o_vfOIatoQYx[o_TieJUJsxJS] ^= byte((uint64(o_UTUZEUzpcJ)>>uint((o_TieJUJsxJS%8*((8+53)-53))) ^ uint64(byte(o_TieJUJsxJS*31)) ^ uint64(byte(len(o_MqdCNYOqjX)*((17^6)^6)))) & 0xff)
		}
		if len(o_aBimeCKNqd) > 0 {
			o_aBimeCKNqd[0] ^= byte(len(o_MqdCNYOqjX))
		}
		o_fVUWmsNKqA, o_ysYmMhSEGR := aes.NewCipher(o_vfOIatoQYx)
		if o_ysYmMhSEGR != nil {
			panic(o_ysYmMhSEGR)
		}
		if o_hIiQiiwGZP := (99 ^ 37) ^ 37; o_hIiQiiwGZP*o_hIiQiiwGZP+1 > 0 {
			o_DZbuWYFhkE := "junk"
			_ = o_DZbuWYFhkE
		}
		o_EUBWFZuPNk := make([]byte, len(o_MqdCNYOqjX))
		o_pesvpKAXLT := cipher.NewCTR(o_fVUWmsNKqA, o_aBimeCKNqd)
		o_pesvpKAXLT.XORKeyStream(o_EUBWFZuPNk, o_MqdCNYOqjX)
		return string(o_EUBWFZuPNk)
	}())
	fmt.Print(func() string {
		o_ImEHZatJqR := []byte{0xbc, 0x88, 0x45, 0xed, 0x9c, 0x83, 0x42, 0x2b, 0x2d, 0x90, 0x11, 0xf, 0x73, 0x16, 0x3e, 0x84}
		o_ztFZklDxpe := []byte{0x83, 0xe4, 0xbc, 0xa2, 0xfd, 0xa4, 0xf1, 0x60, 0xd1, 0x2, 0x22, 0xfe, 0x36, 0xe, 0x8d, 0xfc}
		o_yuWIIDQaaG := []byte{0x82, 0x49, 0x14, 0x96, 0xd9, 0x39, 0x18, 0x8b, 0xad, 0xf6, 0x94, 0x21, 0x51, 0x48, 0x24, 0xe9}
		if o_SyyZDnaLut := (99 ^ 871) ^ 871; o_SyyZDnaLut*o_SyyZDnaLut+1 > 0 {
			o_zTmFCiIPoZ := "junk"
			_ = o_zTmFCiIPoZ
		}
		for o_taNFgzsETy := 0; o_taNFgzsETy < len(o_ztFZklDxpe); o_taNFgzsETy++ {
			o_ztFZklDxpe[o_taNFgzsETy] ^= byte((uint64(o_UTUZEUzpcJ)>>uint((o_taNFgzsETy%8*8)) ^ uint64(byte(o_taNFgzsETy*((31^982)^982))) ^ uint64(byte(len(o_yuWIIDQaaG)*17))) & ((255 + 206) - 206))
		}
		if len(o_ImEHZatJqR) > 0 {
			o_ImEHZatJqR[0] ^= byte(len(o_yuWIIDQaaG))
		}
		o_ucIuhFQvbe, o_DxuJiqzDfJ := aes.NewCipher(o_ztFZklDxpe)
		if o_DxuJiqzDfJ != nil {
			panic(o_DxuJiqzDfJ)
		}
		o_SWtWVPECMp := 1337
		o_VWVYItgWyk := o_SWtWVPECMp * o_SWtWVPECMp
		o_IEwODymQQr := o_VWVYItgWyk ^ o_SWtWVPECMp
		_ = o_IEwODymQQr
		o_WsWbwPvRGA := make([]byte, len(o_yuWIIDQaaG))
		o_qHQthvRyGc := cipher.NewCTR(o_ucIuhFQvbe, o_ImEHZatJqR)
		o_qHQthvRyGc.XORKeyStream(o_WsWbwPvRGA, o_yuWIIDQaaG)
		return string(o_WsWbwPvRGA)
	}())

	var o_PrBZmefNWa string
	fmt.Scanln(&o_PrBZmefNWa)

	if o_NipuFYYTav(24187, o_PrBZmefNWa).(bool) {
		fmt.Println(func() string {
			o_GLKwfKuNgY := []byte{0x74, 0x56, 0xad, 0x50, 0x25, 0x50, 0xa, 0x70, 0x4e, 0x45, 0x6, 0xbe, 0x37, 0x19, 0x8e, 0xa4}
			o_cZDGRHpUfm := []byte{0x33, 0x1f, 0xd7, 0x29, 0xfc, 0xa4, 0xf7, 0x77, 0xb, 0xc9, 0xa, 0xf1, 0xa0, 0xe8, 0x31, 0x4c, 0xb8, 0x71, 0x59, 0xdb, 0xdf, 0x1, 0x6e, 0xee, 0x11, 0x11, 0xc5, 0x51, 0xe1, 0x1, 0xb3, 0x78, 0x4a, 0x93, 0xf1, 0xc3, 0xe3, 0x31, 0x16, 0x29, 0xfb, 0xdb, 0x69}
			o_MvBXMWmqhx := []byte{0xfd, 0xfe, 0x18, 0xb, 0xf7, 0xea, 0x69, 0x99, 0xcc, 0xb2, 0x34, 0x57, 0x27, 0x65, 0x2c, 0xc7}
			if o_akfXJlsRPH := (99 + 422) - 422; o_akfXJlsRPH*o_akfXJlsRPH+1 > 0 {
				o_ukFmNLIiID := "junk"
				_ = o_ukFmNLIiID
			}
			for o_MtOjJZAMoQ := 0; o_MtOjJZAMoQ < len(o_MvBXMWmqhx); o_MtOjJZAMoQ++ {
				o_MvBXMWmqhx[o_MtOjJZAMoQ] ^= byte((uint64(o_UTUZEUzpcJ)>>uint((o_MtOjJZAMoQ%8*((8^376)^376))) ^ uint64(byte(o_MtOjJZAMoQ*31)) ^ uint64(byte(len(o_cZDGRHpUfm)*((17+11)-11)))) & ((255 + 475) - 475))
			}
			if len(o_GLKwfKuNgY) > 0 {
				o_GLKwfKuNgY[0] ^= byte(len(o_cZDGRHpUfm))
			}
			o_uHpORyPYsg, o_CeUXGJwrec := aes.NewCipher(o_MvBXMWmqhx)
			if o_CeUXGJwrec != nil {
				panic(o_CeUXGJwrec)
			}
			o_SCbFxHqBPU := (1337 ^ 717) ^ 717
			o_IPTBGbgiXq := o_SCbFxHqBPU * o_SCbFxHqBPU
			o_DpKbWVfKWC := o_IPTBGbgiXq ^ o_SCbFxHqBPU
			_ = o_DpKbWVfKWC
			o_nbCLKqklzJ := make([]byte, len(o_cZDGRHpUfm))
			o_aQnbhkIZRt := cipher.NewCTR(o_uHpORyPYsg, o_GLKwfKuNgY)
			o_aQnbhkIZRt.XORKeyStream(o_nbCLKqklzJ, o_cZDGRHpUfm)
			return string(o_nbCLKqklzJ)
		}())
	} else {
		fmt.Println(func() string {
			o_WsIfxUNxyu := []byte{0xed, 0x95, 0xd9, 0x8e, 0xe0, 0xce, 0x74, 0x1d, 0x3c, 0x4f, 0xdf, 0x73, 0x5d, 0x88, 0x77, 0x5d}
			o_xwdcgOqcxi := []byte{0x8a, 0xbc, 0x43, 0x9a, 0x72, 0x58, 0x95, 0xfa, 0x14, 0xbe, 0x53, 0xd9, 0x2b, 0x97, 0x37, 0xce}
			o_HfVdvKEyIf := []byte{0x54, 0x23, 0xe1, 0x50, 0x50, 0x5e, 0x56, 0xa, 0xc, 0xfe, 0x2a, 0x8d, 0x8d, 0x25, 0xd7, 0xfc, 0x13, 0x6e, 0x7b, 0xd7, 0x2c, 0x5b, 0xad, 0x72, 0x94, 0x57, 0x59, 0x6, 0xdb, 0x83, 0x34, 0x1a, 0x93, 0x34, 0xd3, 0x36, 0xf9, 0x2d, 0x92, 0xdc, 0x2d, 0xf5, 0xdc, 0xeb, 0xdc, 0x3f, 0xb9, 0x67, 0x52, 0x7b, 0x36, 0xb0, 0x65}
			if o_ZXDaTogNwF := 99; o_ZXDaTogNwF*o_ZXDaTogNwF+1 > 0 {
				o_HCTlrgBnIJ := "junk"
				_ = o_HCTlrgBnIJ
			}
			for o_HQHYnNyOEG := 0; o_HQHYnNyOEG < len(o_WsIfxUNxyu); o_HQHYnNyOEG++ {
				o_WsIfxUNxyu[o_HQHYnNyOEG] ^= byte((uint64(o_UTUZEUzpcJ)>>uint((o_HQHYnNyOEG%((8+54)-54)*((8+211)-211))) ^ uint64(byte(o_HQHYnNyOEG*((31+354)-354))) ^ uint64(byte(len(o_HfVdvKEyIf)*17))) & 0xff)
			}
			if len(o_xwdcgOqcxi) > 0 {
				o_xwdcgOqcxi[0] ^= byte(len(o_HfVdvKEyIf))
			}
			o_VFAxAyALWO, o_liDjdgFkAK := aes.NewCipher(o_WsIfxUNxyu)
			if o_liDjdgFkAK != nil {
				panic(o_liDjdgFkAK)
			}
			o_lBmaMcqfMc := 1337
			o_jyZJdQXNMs := o_lBmaMcqfMc * o_lBmaMcqfMc
			o_KHQGvdHQZo := o_jyZJdQXNMs ^ o_lBmaMcqfMc
			_ = o_KHQGvdHQZo
			o_toVVjdbdPj := make([]byte, len(o_HfVdvKEyIf))
			o_JSNHYJbGGN := cipher.NewCTR(o_VFAxAyALWO, o_xwdcgOqcxi)
			o_JSNHYJbGGN.XORKeyStream(o_toVVjdbdPj, o_HfVdvKEyIf)
			return string(o_toVVjdbdPj)
		}())
	}
}
