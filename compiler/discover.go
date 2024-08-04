package compiler

import (
	"io"
	"runtime"
	"strings"

	"github.com/adnsv/go-build/compiler/clang"
	"github.com/adnsv/go-build/compiler/gcc"
	"github.com/adnsv/go-build/compiler/msvc"
	"github.com/adnsv/go-build/compiler/toolchain"
	"github.com/adnsv/go-build/compiler/triplet"
	"golang.org/x/exp/slices"
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

func Find(target triplet.Target, tt []*toolchain.Chain) []*toolchain.Chain {
	ret := []*toolchain.Chain{}
	for _, t := range tt {
		if t.Target.Match(target) {
			ret = append(ret, t)
		}
	}
	return ret
}

func Natives(tt []*toolchain.Chain) []*toolchain.Chain {
	target := triplet.Target{
		OS:   triplet.NormalizeOS(runtime.GOOS),
		Arch: triplet.NormalizeArch(runtime.GOARCH),
	}
	return Find(target, tt)
}

func ChooseNative(tt []*toolchain.Chain, order_of_preference ...string) *toolchain.Chain {
	tt = Natives(tt)
	if len(tt) <= 0 {
		return nil
	} else if len(tt) == 0 {
		return tt[0]
	}

	handle_compiler := func(compiler string) *toolchain.Chain {
		sel := []*toolchain.Chain{}
		for _, t := range tt {
			if t.Compiler == compiler {
				sel = append(sel, t)
			}
		}
		if len(sel) == 0 {
			return nil
		} else if len(sel) == 1 {
			return sel[0]
		}

		if compiler == "msvc" {
			slices.SortFunc(sel, func(c1, c2 *toolchain.Chain) int {
				return msvc.Compare(c1, c2)
			})
		} else {
			slices.SortFunc(sel, func(c1, c2 *toolchain.Chain) int {
				return gcc.Compare(c1, c2)
			})
		}
		return sel[len(sel)-1]
	}

	if order_of_preference == nil {
		order_of_preference = []string{"gcc", "clang", "msvc"}
	}
	for _, c := range order_of_preference {
		r := handle_compiler(c)
		if r != nil {
			return r
		}
	}
	sel := slices.Clone(tt)
	slices.SortFunc(sel, func(c1, c2 *toolchain.Chain) int {
		if i := strings.Compare(c1.Compiler, c2.Compiler); i != 0 {
			return i
		}
		if i := strings.Compare(c1.Target.Original, c2.Target.Original); i != 0 {
			return i
		}
		return strings.Compare(c1.FullVersion, c2.FullVersion)
	})
	return sel[0]
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
