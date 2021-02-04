package compiler

import (
	"io"

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
