package compiler

import (
	"io"
	"strings"

	"github.com/adnsv/go-build/compiler/clang"
	"github.com/adnsv/go-build/compiler/gcc"
	"github.com/adnsv/go-build/compiler/msvc"
	"github.com/adnsv/go-build/compiler/toolchain"
)

type Installation interface {
	PrintSummary(w io.Writer)
}

func DiscoverInstallations(types []string, feedback func(string)) []Installation {
	ret := []Installation{}
	if fltShow("msvc", types) {
		ii, _ := msvc.DiscoverInstallations(feedback)
		for _, i := range ii {
			ret = append(ret, i)
		}
	}
	if fltShow("gcc", types) || fltShow("gnu", types) {
		ii := gcc.DiscoverInstallations(feedback)
		for _, i := range ii {
			ret = append(ret, i)
		}
	}
	if fltShow("clang", types) || fltShow("llvm", types) {
		ii := clang.DiscoverInstallations(feedback)
		for _, i := range ii {
			ret = append(ret, i)
		}
	}
	return ret
}

func DiscoverToolchains(wantCxx bool, types []string, feedback func(string)) []*toolchain.Chain {
	ret := []*toolchain.Chain{}
	if fltShow("msvc", types) {
		ret = append(ret, msvc.DiscoverToolchains(feedback)...)
	}
	if fltShow("gcc", types) || fltShow("gnu", types) {
		ret = append(ret, gcc.DiscoverToolchains(wantCxx, feedback)...)
	}
	if fltShow("clang", types) || fltShow("llvm", types) {
		ret = append(ret, clang.DiscoverToolchains(wantCxx, feedback)...)
	}
	return ret
}

func archNorm(arch string) string {
	var archNorm = map[string]string{
		"x64":     "x64",
		"amd64":   "x64",
		"x86_64":  "x64",
		"x32":     "x32",
		"86":      "x32",
		"x86":     "x32",
		"386":     "x32",
		"486":     "x32",
		"586":     "x32",
		"686":     "x32",
		"arm":     "arm32",
		"arm32":   "arm32",
		"arm64":   "arm64",
		"aarch64": "arm64",
	}
	arch = strings.ToLower(arch)
	if norm, ok := archNorm[arch]; ok {
		arch = norm
	}
	return arch
}

func FindArch(arch string, tt []*toolchain.Chain) []*toolchain.Chain {
	arch = archNorm(arch)

	ret := []*toolchain.Chain{}
	for _, t := range tt {
		if t.Compiler == "MSVC" {
			if arch == archNorm(t.VisualStudioArch) {
				ret = append(ret, t)
			}
		} else {
			pp := strings.Split(t.Target, "-")
			if len(pp) > 0 && arch == archNorm(pp[0]) {
				ret = append(ret, t)
			}
		}
	}
	return ret
}

func fltShow(t string, tt []string) bool {
	if len(tt) == 0 || (len(tt) == 1 && tt[0] == "") {
		return true
	}
	for _, it := range tt {
		if it == t {
			return true
		}
	}
	return false
}
