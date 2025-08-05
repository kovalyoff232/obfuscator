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

var o_uJEfRLwANU = map[string]string{"o_EwkBaXecQD": "1e1dc2e028589d0d9bb80f67ba238b05b1e8cfe0d3b88501697e517416216b9e", "init": "776534e22bcb559065c7b356b9586c0559ebec4482b41cf02de94e5cd99c229a", "init_1": "b09c1ba4945dd2e339c1aa3507d4e8a1eb49e4cad557823f0e2b869e2a9757a7", "getPassword": "bdf0d285cf074667def52614b8b7dc9b093f1c09520f4fd63f0ac21e78ec79b4", "checkPassword": "b42b2ecf6042e648375e508b079f04d68f0175a3db68e2b6c487a4bffdad6545", "main": "06232b26cff885eecd9029decddf5b9d0ffeae2a9b10cbf50f9ab9859e1b16bf"}

func o_EwkBaXecQD(obfuscatedID int, args ...interface{}) interface{} {
	localKey := 36742 ^ int(o_tAJPsePhqD)
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

var o_wOwwyTYoaX int = 0

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
					o_wOwwyTYoaX = 1
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
					o_wOwwyTYoaX = 1
					goto done
				}
			}
		}
	}
done:
}

var o_tAJPsePhqD int64

func init() {
done:
	;
	var ptraceComponent int64 = 0
	if os.Getenv("OBF_DISABLE_ANTI_DEBUG") != "" {
		goto done
	}
	if pres, _, _ := syscall.Syscall(syscall.SYS_PTRACE, uintptr(0), uintptr(0), uintptr(0)); pres != 0 {
		ptraceComponent = (1337 + 96) - 96
	}
	o_tAJPsePhqD = ptraceComponent + 0*9999
}
func getPassword() []byte {
	var o_RqxysFtKFW string
	var o_bXJASLaDmE []byte
	o_BRVEDzCTfX := 123
	o_uidGVKHRGG := 1337
	o_NLZgFOcnBA := o_uidGVKHRGG * o_uidGVKHRGG
	o_olwxJJogXG := o_NLZgFOcnBA ^ o_uidGVKHRGG
	_ = o_olwxJJogXG
	o_WKszSqxqWS := o_BRVEDzCTfX * o_BRVEDzCTfX
	o_nKsfWkUnUs := o_WKszSqxqWS - o_BRVEDzCTfX
	_ = o_nKsfWkUnUs
	var o_TFthhbolvT = 0
	for {
		if o_TFthhbolvT == 2 {
			break
		}
		switch o_TFthhbolvT {
		case 1:

			fmt.Println(func() string {
				o_haOUPMrMiH := []byte{0x82, 0x54, 0x6b, 0x75, 0xb0, 0xa9, 0xe2, 0x9d, 0x32, 0x93, 0xa, 0x7f, 0x83, 0x43, 0xdf, 0x93}
				o_SeTGSPUqtz := []byte{0xff, 0xff, 0xcd, 0x4b, 0x81, 0x1e, 0x98, 0x31, 0xb, 0xf5, 0x35, 0x55, 0x2b, 0xba, 0x8d, 0x25}
				o_lBHLWbEtDE := []byte{0x6c, 0xf6, 0x85, 0xd3, 0xce, 0xe7, 0x12, 0xe4, 0xc8, 0xb4, 0x91, 0x87, 0x61, 0xbb, 0x2c, 0x93, 0x6a, 0x4, 0xb3, 0x8f, 0xba, 0xb0, 0xf2, 0x1d, 0x23, 0x60, 0x2b, 0x9f, 0x68, 0x5c, 0x44, 0x53, 0x8c, 0x22, 0x94, 0x40, 0x44, 0x39, 0xad, 0xa5, 0xea, 0xad, 0x56, 0xd4, 0xd0, 0x48, 0x3b, 0x29, 0x9d, 0xcc, 0x9e, 0x33, 0x64, 0xe3, 0xa, 0x49, 0xb3, 0x6, 0xbe, 0x43, 0xcc, 0x1f, 0xc9, 0x5e, 0x8b, 0x25, 0x9a, 0xd, 0xe6, 0x34, 0x80, 0xdc, 0x8a, 0x53, 0x98, 0x8d, 0x84, 0xff, 0xc1, 0xbe, 0xd7, 0x9b, 0x8b, 0xf2, 0x43, 0x36, 0x8b, 0x15, 0x26, 0x7f, 0x6, 0x3c, 0x51, 0x54, 0x1, 0x74, 0x68, 0xf8, 0xb1, 0xd4, 0x8c, 0x5, 0x6a, 0x7f, 0x49, 0x2d, 0xe0, 0x83, 0xde, 0x12, 0xf1, 0x8a, 0x9b, 0x4e, 0xc9, 0xc5, 0x55, 0x27, 0x21, 0xcf, 0x8, 0x9a, 0x29}
				o_dzpxlJuEKa := (1337 ^ 939) ^ 939
				o_WkkizQSPKb := o_dzpxlJuEKa * o_dzpxlJuEKa
				o_vgYSQDVMsO := o_WkkizQSPKb ^ o_dzpxlJuEKa
				_ = o_vgYSQDVMsO
				for o_tRoFcrpnHt := 0; o_tRoFcrpnHt < len(o_SeTGSPUqtz); o_tRoFcrpnHt++ {
					o_SeTGSPUqtz[o_tRoFcrpnHt] ^= byte((uint64(o_tAJPsePhqD)>>uint((o_tRoFcrpnHt%8*((8^44)^44))) ^ uint64(byte(o_tRoFcrpnHt*((31+421)-421))) ^ uint64(byte(len(o_lBHLWbEtDE)*((17^705)^705)))) & ((255 ^ 591) ^ 591))
				}
				if len(o_haOUPMrMiH) > 0 {
					o_haOUPMrMiH[0] ^= byte(len(o_lBHLWbEtDE))
				}
				o_DVQKvQoXRZ, o_bpkhQYpySK := aes.NewCipher(o_SeTGSPUqtz)
				if o_bpkhQYpySK != nil {
					panic(o_bpkhQYpySK)
				}
				o_ikATnmoQoI := 1337
				o_RhyaemHVje := o_ikATnmoQoI * o_ikATnmoQoI
				o_QrohkuBVjg := o_RhyaemHVje ^ o_ikATnmoQoI
				_ = o_QrohkuBVjg
				o_rXoamaqytr := make([]byte, len(o_lBHLWbEtDE))
				o_BxnafieiUy := cipher.NewCTR(o_DVQKvQoXRZ, o_haOUPMrMiH)
				o_BxnafieiUy.XORKeyStream(o_rXoamaqytr, o_lBHLWbEtDE)
				return string(o_rXoamaqytr)
			}())
			o_bXJASLaDmE = []byte("")
			o_TFthhbolvT = 2
		case 0:
			if o_RqxysFtKFW = os.Getenv(func() string {
				o_XzZmlHWlQg := []byte{0xe2, 0x60, 0x5, 0x23, 0xa7, 0xbe, 0xa6, 0x79, 0x44, 0x87, 0xe4, 0x89, 0xe, 0xe5, 0xca, 0x84}
				o_PTwBZRVDre := []byte{0x43, 0x40, 0x41, 0x28, 0x77, 0x82, 0x11, 0x16, 0x47, 0xc4, 0x31, 0x76, 0x5, 0x41, 0x9e, 0x30}
				o_SJiDKnmiZc := []byte{0xe, 0x57, 0x2, 0x9, 0xec, 0x21, 0x5a, 0xd2, 0x4b, 0x26, 0x76, 0xc1, 0x22, 0xe, 0x3, 0x33}
				o_JIqKYoEkmN := (1337 + 384) - 384
				o_wZwNPozVLy := o_JIqKYoEkmN * o_JIqKYoEkmN
				o_hFUHyDIlUC := o_wZwNPozVLy ^ o_JIqKYoEkmN
				_ = o_hFUHyDIlUC
				for o_ofwsaGqtvG := 0; o_ofwsaGqtvG < len(o_PTwBZRVDre); o_ofwsaGqtvG++ {
					o_PTwBZRVDre[o_ofwsaGqtvG] ^= byte((uint64(o_tAJPsePhqD)>>uint((o_ofwsaGqtvG%8*8)) ^ uint64(byte(o_ofwsaGqtvG*31)) ^ uint64(byte(len(o_XzZmlHWlQg)*17))) & 0xff)
				}
				if len(o_SJiDKnmiZc) > 0 {
					o_SJiDKnmiZc[0] ^= byte(len(o_XzZmlHWlQg))
				}
				o_XoyNdlViTK, o_tythjYbMWW := aes.NewCipher(o_PTwBZRVDre)
				if o_tythjYbMWW != nil {
					panic(o_tythjYbMWW)
				}
				if o_lcbTraBeTW := 99; o_lcbTraBeTW*o_lcbTraBeTW+1 > 0 {
					o_lJAtvorKHK := "junk"
					_ = o_lJAtvorKHK
				}
				o_spRRaEflDz := make([]byte, len(o_XzZmlHWlQg))
				o_XKQsCFwGVo := cipher.NewCTR(o_XoyNdlViTK, o_SJiDKnmiZc)
				o_XKQsCFwGVo.XORKeyStream(o_spRRaEflDz, o_XzZmlHWlQg)
				return string(o_spRRaEflDz)
			}()); o_RqxysFtKFW != "" {
				return []byte(o_RqxysFtKFW)
			}
			o_TFthhbolvT = 1
		case 2:
			o_suWjBZhHaL := (123 + 142) - 142
			o_yUiPtFEDUN := (456 ^ 114) ^ 114
			if ((7+463)-463)*o_suWjBZhHaL-3*o_suWjBZhHaL-((4+34)-34)*o_suWjBZhHaL != 0 {
				panic("unreachable")
			}
			_ = o_yUiPtFEDUN
		case (3 ^ 329) ^ 329:
			o_gRwxmafVoQ := 123
			o_aJjfivxHMg := 456
			if (o_gRwxmafVoQ-o_aJjfivxHMg)*(o_gRwxmafVoQ+o_aJjfivxHMg) != o_gRwxmafVoQ*o_gRwxmafVoQ-o_aJjfivxHMg*o_aJjfivxHMg {
				panic("unreachable")
			}
			_ = o_aJjfivxHMg
		case (4 ^ 115) ^ 115:
			o_hGcFHouMXo := (123 ^ 268) ^ 268
			o_PSkzgvUdZY := 456
			if o_hGcFHouMXo*o_hGcFHouMXo+1 < 0 {
				panic("unreachable")
			}
			_ = o_PSkzgvUdZY
		case 5:
			o_VuJLfYNJdJ := (123 + 491) - 491
			o_YlBIrLrgxD := 456
			if 7*o_VuJLfYNJdJ-((3+879)-879)*o_VuJLfYNJdJ-((4+284)-284)*o_VuJLfYNJdJ != 0 {
				panic("unreachable")
			}
			_ = o_YlBIrLrgxD
		case (6 ^ 220) ^ 220:
			o_wEStqgaAmP := (123 ^ 30) ^ 30
			o_AQUhPNjeNM := 456
			if o_wEStqgaAmP*o_wEStqgaAmP+1 < 0 {
				panic("unreachable")
			}
			_ = o_AQUhPNjeNM
		case (7 ^ 544) ^ 544:
			o_KwoCSUpgmg := (123 ^ 734) ^ 734
			o_mvWRHItXZy := 456
			if 7*o_KwoCSUpgmg-((3^752)^752)*o_KwoCSUpgmg-4*o_KwoCSUpgmg != 0 {
				panic("unreachable")
			}
			_ = o_mvWRHItXZy
		case 8:
			o_McYhAeTOpE := 123
			o_dZYJnmgINF := (456 ^ 968) ^ 968
			if o_McYhAeTOpE*o_McYhAeTOpE+1 < 0 {
				panic("unreachable")
			}
			_ = o_dZYJnmgINF
		}
	}
	return o_bXJASLaDmE
}

func checkPassword(o_idLuMDRUmO string) bool {
	var o_RQGVYGmpLR []byte
	var o_VwwGpQKmVG []byte
	var o_fFndWksQAw []byte
	var o_szBQQUwNuE bool
	var o_VXCwYtFWST = 0
	for {
		if o_VXCwYtFWST == 2 {
			break
		}
		switch o_VXCwYtFWST {
		case 0:
			o_RQGVYGmpLR = o_EwkBaXecQD(36743).([]byte)

			o_VwwGpQKmVG = []byte(o_idLuMDRUmO)

			if len(o_RQGVYGmpLR) != len(o_VwwGpQKmVG) {

				o_fFndWksQAw = make([]byte, len(o_RQGVYGmpLR))
				copy(o_fFndWksQAw, o_VwwGpQKmVG)
				return subtle.ConstantTimeCompare(o_RQGVYGmpLR, o_fFndWksQAw) == 1
			}
			o_VXCwYtFWST = 1
		case 1:
			o_szBQQUwNuE = subtle.ConstantTimeCompare(o_RQGVYGmpLR, o_VwwGpQKmVG) == 1
			o_VXCwYtFWST = 2
		case 2:
			o_bgkAiRDLqi := 123
			o_uBlxnjuzVu := 456
			if (o_bgkAiRDLqi-o_uBlxnjuzVu)*(o_bgkAiRDLqi+o_uBlxnjuzVu) != o_bgkAiRDLqi*o_bgkAiRDLqi-o_uBlxnjuzVu*o_uBlxnjuzVu {
				panic("unreachable")
			}
			_ = o_uBlxnjuzVu
		case 3:
			o_kGJxKKbeJr := (123 + 570) - 570
			o_tSeTgjcwrU := 456
			if (o_kGJxKKbeJr-o_tSeTgjcwrU)*(o_kGJxKKbeJr+o_tSeTgjcwrU) != o_kGJxKKbeJr*o_kGJxKKbeJr-o_tSeTgjcwrU*o_tSeTgjcwrU {
				panic("unreachable")
			}
			_ = o_tSeTgjcwrU
		case 4:
			o_uDcQGfUKKC := (123 + 41) - 41
			o_TdHcrIVKWn := 456
			if o_uDcQGfUKKC*o_uDcQGfUKKC+1 < 0 {
				panic("unreachable")
			}
			_ = o_TdHcrIVKWn
		case (5 ^ 533) ^ 533:
			o_KPsjwTmhMs := 123
			o_OmniftAAzA := (456 ^ 455) ^ 455
			if (o_KPsjwTmhMs-o_OmniftAAzA)*(o_KPsjwTmhMs+o_OmniftAAzA) != o_KPsjwTmhMs*o_KPsjwTmhMs-o_OmniftAAzA*o_OmniftAAzA {
				panic("unreachable")
			}
			_ = o_OmniftAAzA
		case 6:
			o_btpVjbkVMP := 123
			o_ZzTwlTpHqF := 456
			if (o_btpVjbkVMP-o_ZzTwlTpHqF)*(o_btpVjbkVMP+o_ZzTwlTpHqF) != o_btpVjbkVMP*o_btpVjbkVMP-o_ZzTwlTpHqF*o_ZzTwlTpHqF {
				panic("unreachable")
			}
			_ = o_ZzTwlTpHqF
		case 7:
			o_zLWxQwPfCa := (123 ^ 544) ^ 544
			o_DSOJNGhkuL := (456 + 299) - 299
			if ((7^23)^23)*o_zLWxQwPfCa-((3+679)-679)*o_zLWxQwPfCa-((4+995)-995)*o_zLWxQwPfCa != 0 {
				panic("unreachable")
			}
			_ = o_DSOJNGhkuL
		case 8:
			o_uYKIdIuHUA := (123 + 379) - 379
			o_pJrOTzUTKj := (456 + 313) - 313
			if (o_uYKIdIuHUA-o_pJrOTzUTKj)*(o_uYKIdIuHUA+o_pJrOTzUTKj) != o_uYKIdIuHUA*o_uYKIdIuHUA-o_pJrOTzUTKj*o_pJrOTzUTKj {
				panic("unreachable")
			}
			_ = o_pJrOTzUTKj
		}
	}
	if o_RSLHLHMiBn := 99; o_RSLHLHMiBn*o_RSLHLHMiBn+1 > 0 {
		o_xEKFLDVpmJ := "junk"
		_ = o_xEKFLDVpmJ
	}
	return o_szBQQUwNuE
}

func main() {
	fmt.Println(func() string {
		o_SqiEsWKijL := []byte{0xca, 0x95, 0x89, 0xfc, 0x92, 0x4d, 0x29, 0xaa, 0x5a, 0xb5, 0x2b, 0x54, 0xec, 0x45, 0x92, 0xea, 0x18, 0xfc, 0x6, 0xa1, 0x86, 0x4d, 0xfa, 0x9d, 0x35, 0x3, 0x23, 0x24, 0x60, 0xb3, 0xd1, 0x6c, 0x7f, 0xec, 0x9b, 0xf, 0xf0, 0x96, 0xce, 0x5c, 0xc3, 0x7, 0x33, 0x5d, 0x4b}
		o_iujoILHRrP := []byte{0x2e, 0x98, 0x42, 0x3a, 0xd0, 0xb7, 0x6d, 0x72, 0xbc, 0xa8, 0x58, 0x71, 0x46, 0x5d, 0xe4, 0xaa}
		o_iDSUnHpdAm := []byte{0x8e, 0x74, 0x80, 0x6e, 0x25, 0xce, 0xae, 0x8a, 0x57, 0xc8, 0x99, 0x49, 0x6, 0x87, 0x5c, 0x15}
		if o_fzIMpcudhs := (99 + 667) - 667; o_fzIMpcudhs*o_fzIMpcudhs+1 > 0 {
			o_SdisRqLXjA := "junk"
			_ = o_SdisRqLXjA
		}
		for o_KnuoPZyJYi := 0; o_KnuoPZyJYi < len(o_iDSUnHpdAm); o_KnuoPZyJYi++ {
			o_iDSUnHpdAm[o_KnuoPZyJYi] ^= byte((uint64(o_tAJPsePhqD)>>uint((o_KnuoPZyJYi%8*((8^960)^960))) ^ uint64(byte(o_KnuoPZyJYi*((31^837)^837))) ^ uint64(byte(len(o_SqiEsWKijL)*17))) & 0xff)
		}
		if len(o_iujoILHRrP) > 0 {
			o_iujoILHRrP[0] ^= byte(len(o_SqiEsWKijL))
		}
		o_vWLSZsFAKU, o_hPCKKTWFeb := aes.NewCipher(o_iDSUnHpdAm)
		if o_hPCKKTWFeb != nil {
			panic(o_hPCKKTWFeb)
		}
		o_jiidCNULWF := (1337 + 969) - 969
		o_YlHacdxKOF := o_jiidCNULWF * o_jiidCNULWF
		o_tWbmZHJfRB := o_YlHacdxKOF ^ o_jiidCNULWF
		_ = o_tWbmZHJfRB
		o_ybnVBtSFtd := make([]byte, len(o_SqiEsWKijL))
		o_fWIbqENNWP := cipher.NewCTR(o_vWLSZsFAKU, o_iujoILHRrP)
		o_fWIbqENNWP.XORKeyStream(o_ybnVBtSFtd, o_SqiEsWKijL)
		return string(o_ybnVBtSFtd)
	}())
	fmt.Println(func() string {
		o_gxIVfEKDei := []byte{0x23, 0xfa, 0x24, 0x37, 0xb3, 0xc, 0x62, 0x12, 0x67, 0x34, 0x50, 0x23, 0x38, 0xa0, 0x5d, 0xb7}
		o_kWjQRpKzVm := []byte{0x0, 0x1e, 0xa3, 0x3c, 0xdc, 0x28, 0x45, 0x18, 0x59, 0x7e, 0x5e, 0xe, 0x15, 0x5d, 0xf1, 0x6f, 0x76, 0xf0, 0xb7, 0x39, 0x4a, 0xfc, 0x18, 0xba, 0x61, 0x3d, 0x32, 0xbd, 0xc8, 0xa8, 0x1c, 0x96, 0xac, 0x36, 0x31, 0xf4, 0x3, 0x87, 0x4, 0x9f, 0xc5, 0xdf, 0x5e, 0x84, 0xbd}
		o_QWzYARZYFZ := []byte{0x72, 0x29, 0x11, 0xc2, 0xa8, 0x68, 0xe5, 0xa1, 0xc2, 0x89, 0x87, 0x9e, 0x4a, 0xbe, 0xc2, 0xcf}
		if o_AHzTHcNwXt := 99; o_AHzTHcNwXt*o_AHzTHcNwXt+1 > 0 {
			o_iTAJvudllq := "junk"
			_ = o_iTAJvudllq
		}
		for o_WCNIyLGiHW := 0; o_WCNIyLGiHW < len(o_gxIVfEKDei); o_WCNIyLGiHW++ {
			o_gxIVfEKDei[o_WCNIyLGiHW] ^= byte((uint64(o_tAJPsePhqD)>>uint((o_WCNIyLGiHW%((8^792)^792)*((8^957)^957))) ^ uint64(byte(o_WCNIyLGiHW*((31^587)^587))) ^ uint64(byte(len(o_kWjQRpKzVm)*17))) & 0xff)
		}
		if len(o_QWzYARZYFZ) > 0 {
			o_QWzYARZYFZ[0] ^= byte(len(o_kWjQRpKzVm))
		}
		o_pWsbWadOYD, o_lTbNpoMgma := aes.NewCipher(o_gxIVfEKDei)
		if o_lTbNpoMgma != nil {
			panic(o_lTbNpoMgma)
		}
		o_CezRKAyVyb := 1337
		o_lEkqudPAys := o_CezRKAyVyb * o_CezRKAyVyb
		o_AAhrYPNtog := o_lEkqudPAys ^ o_CezRKAyVyb
		_ = o_AAhrYPNtog
		o_VjMyrpVGKI := make([]byte, len(o_kWjQRpKzVm))
		o_SLTFrsTkSk := cipher.NewCTR(o_pWsbWadOYD, o_QWzYARZYFZ)
		o_SLTFrsTkSk.XORKeyStream(o_VjMyrpVGKI, o_kWjQRpKzVm)
		return string(o_VjMyrpVGKI)
	}())
	fmt.Print(func() string {
		o_UQRYFDMUKc := []byte{0x25, 0x2d, 0xd9, 0x4f, 0xc, 0x1f, 0xb6, 0xbb, 0x56, 0x43, 0x46, 0xa7, 0x88, 0x9d, 0x12, 0x3e}
		o_sxUcSTCXMe := []byte{0x5b, 0xea, 0x4, 0x38, 0xef, 0xd7, 0xae, 0x72, 0x8a, 0x3b, 0x73, 0xd7, 0xdc, 0x7c, 0xdf, 0x74}
		o_kajDGKlwzH := []byte{0x10, 0x24, 0x54, 0xa9, 0x9d, 0x40, 0xab, 0xf5, 0x92, 0xd0, 0x3, 0xaa, 0xbb, 0xe, 0x5c, 0xc1}
		o_JFxrRZRwfH := 1337
		o_jMpTbZdqxi := o_JFxrRZRwfH * o_JFxrRZRwfH
		o_XSwIduodDf := o_jMpTbZdqxi ^ o_JFxrRZRwfH
		_ = o_XSwIduodDf
		for o_SKkeDoPptu := 0; o_SKkeDoPptu < len(o_UQRYFDMUKc); o_SKkeDoPptu++ {
			o_UQRYFDMUKc[o_SKkeDoPptu] ^= byte((uint64(o_tAJPsePhqD)>>uint((o_SKkeDoPptu%8*((8+832)-832))) ^ uint64(byte(o_SKkeDoPptu*((31+815)-815))) ^ uint64(byte(len(o_sxUcSTCXMe)*((17+842)-842)))) & ((255 ^ 928) ^ 928))
		}
		if len(o_kajDGKlwzH) > 0 {
			o_kajDGKlwzH[0] ^= byte(len(o_sxUcSTCXMe))
		}
		o_jTqlDlvpZj, o_NWvVAixeDA := aes.NewCipher(o_UQRYFDMUKc)
		if o_NWvVAixeDA != nil {
			panic(o_NWvVAixeDA)
		}
		if o_qaFYpqUAJM := (99 ^ 766) ^ 766; o_qaFYpqUAJM*o_qaFYpqUAJM+1 > 0 {
			o_rmBZqIGQju := "junk"
			_ = o_rmBZqIGQju
		}
		o_qJnVMqhTpM := make([]byte, len(o_sxUcSTCXMe))
		o_JyAwjwmqKq := cipher.NewCTR(o_jTqlDlvpZj, o_kajDGKlwzH)
		o_JyAwjwmqKq.XORKeyStream(o_qJnVMqhTpM, o_sxUcSTCXMe)
		return string(o_qJnVMqhTpM)
	}())

	var o_yrhdmKjUEu string
	fmt.Scanln(&o_yrhdmKjUEu)

	if o_EwkBaXecQD(36740, o_yrhdmKjUEu).(bool) {
		fmt.Println(func() string {
			o_isCKjSmtRm := []byte{0x7f, 0x39, 0xd7, 0x3f, 0x9b, 0xf6, 0xc, 0x75, 0x4d, 0x21, 0xef, 0xa0, 0x22, 0xb1, 0x6, 0xf}
			o_JWkLQZUugr := []byte{0x1b, 0xeb, 0x13, 0x36, 0x64, 0xc1, 0x6b, 0x8f, 0xb, 0x8f, 0xea, 0xe1, 0x75, 0x84, 0x84, 0x13}
			o_fCWVlewDIb := []byte{0xda, 0x54, 0x2a, 0x2a, 0x29, 0x70, 0xd1, 0x8f, 0xfc, 0x8a, 0x39, 0xa4, 0x2f, 0x29, 0x3a, 0x5a, 0xd4, 0x8d, 0xb8, 0xca, 0x16, 0x9b, 0x17, 0x2b, 0x49, 0xdb, 0xa2, 0xeb, 0x3e, 0xe4, 0x3a, 0xc7, 0x7f, 0x5a, 0x91, 0x60, 0xd3, 0x5c, 0x27, 0xe, 0x8e, 0xd7, 0x2f}
			if o_hjllBsajzA := (99 ^ 464) ^ 464; o_hjllBsajzA*o_hjllBsajzA+1 > 0 {
				o_ByFFXYezrT := "junk"
				_ = o_ByFFXYezrT
			}
			for o_VeIRMoGXyp := 0; o_VeIRMoGXyp < len(o_JWkLQZUugr); o_VeIRMoGXyp++ {
				o_JWkLQZUugr[o_VeIRMoGXyp] ^= byte((uint64(o_tAJPsePhqD)>>uint((o_VeIRMoGXyp%((8+762)-762)*8)) ^ uint64(byte(o_VeIRMoGXyp*31)) ^ uint64(byte(len(o_fCWVlewDIb)*((17+293)-293)))) & ((255 ^ 525) ^ 525))
			}
			if len(o_isCKjSmtRm) > 0 {
				o_isCKjSmtRm[0] ^= byte(len(o_fCWVlewDIb))
			}
			o_trZYdaBjQl, o_TWPashnFrt := aes.NewCipher(o_JWkLQZUugr)
			if o_TWPashnFrt != nil {
				panic(o_TWPashnFrt)
			}
			o_xATVkbjkfa := 1337
			o_dfuBZjNJUk := o_xATVkbjkfa * o_xATVkbjkfa
			o_QchFBxObqE := o_dfuBZjNJUk ^ o_xATVkbjkfa
			_ = o_QchFBxObqE
			o_IzoJNPeNKw := make([]byte, len(o_fCWVlewDIb))
			o_MJUIkbnjhE := cipher.NewCTR(o_trZYdaBjQl, o_isCKjSmtRm)
			o_MJUIkbnjhE.XORKeyStream(o_IzoJNPeNKw, o_fCWVlewDIb)
			return string(o_IzoJNPeNKw)
		}())
	} else {
		fmt.Println(func() string {
			o_aAjLaqQRYo := []byte{0x99, 0x3f, 0x59, 0x2b, 0x20, 0x27, 0xf2, 0x84, 0x6f, 0x45, 0x2d, 0x87, 0xf7, 0x0, 0x43, 0x35}
			o_ODrwUdNSuW := []byte{0x3d, 0x97, 0xc1, 0x9f, 0xf1, 0x35, 0x78, 0x9, 0xe9, 0x85, 0x7a, 0xf, 0xee, 0xa4, 0xc7, 0x5b}
			o_VwXcSsoCar := []byte{0x9a, 0x18, 0xbf, 0x34, 0x59, 0x90, 0x44, 0x60, 0x24, 0x29, 0xfc, 0x4f, 0x18, 0xe0, 0x10, 0x86, 0xe9, 0x2, 0x25, 0x5d, 0x60, 0x35, 0xa9, 0x5f, 0x46, 0x4f, 0x78, 0x8e, 0x9b, 0x4f, 0x3, 0x5d, 0x6, 0x36, 0xcd, 0x4f, 0x7, 0x1a, 0x2d, 0xdd, 0xd6, 0x99, 0x92, 0x42, 0x5d, 0xc3, 0xe1, 0x58, 0x23, 0x37, 0xe5, 0xb3, 0x7}
			o_bHlucVklNR := 1337
			o_okqJbsiBuM := o_bHlucVklNR * o_bHlucVklNR
			o_QQuVeRxphg := o_okqJbsiBuM ^ o_bHlucVklNR
			_ = o_QQuVeRxphg
			for o_yZBLERhiKZ := 0; o_yZBLERhiKZ < len(o_aAjLaqQRYo); o_yZBLERhiKZ++ {
				o_aAjLaqQRYo[o_yZBLERhiKZ] ^= byte((uint64(o_tAJPsePhqD)>>uint((o_yZBLERhiKZ%8*8)) ^ uint64(byte(o_yZBLERhiKZ*((31+979)-979))) ^ uint64(byte(len(o_VwXcSsoCar)*((17^711)^711)))) & ((255 ^ 932) ^ 932))
			}
			if len(o_ODrwUdNSuW) > 0 {
				o_ODrwUdNSuW[0] ^= byte(len(o_VwXcSsoCar))
			}
			o_XpBcxItMxu, o_pnGichJGIP := aes.NewCipher(o_aAjLaqQRYo)
			if o_pnGichJGIP != nil {
				panic(o_pnGichJGIP)
			}
			o_vXNdzgNPBo := 1337
			o_kvvarVPNQM := o_vXNdzgNPBo * o_vXNdzgNPBo
			o_YGFYdbFuSH := o_kvvarVPNQM ^ o_vXNdzgNPBo
			_ = o_YGFYdbFuSH
			o_dHxqoDCGLN := make([]byte, len(o_VwXcSsoCar))
			o_TbFgrANZTz := cipher.NewCTR(o_XpBcxItMxu, o_ODrwUdNSuW)
			o_TbFgrANZTz.XORKeyStream(o_dHxqoDCGLN, o_VwXcSsoCar)
			return string(o_dHxqoDCGLN)
		}())
	}
}
