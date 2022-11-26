package compiler

import (
	"io"
	"runtime"
	"strings"

	"github.com/adnsv/go-build/compiler/clang"
	"github.com/adnsv/go-build/compiler/gcc"
	"github.com/adnsv/go-build/compiler/msvc"
	"github.com/adnsv/go-build/compiler/toolchain"
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

func normArch(arch string) string {
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
		"i686":    "x32",
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

func normOS(os string) string {
	os = strings.ToLower(os)
	return os
}

func FindOsArch(os, arch string, tt []*toolchain.Chain) []*toolchain.Chain {
	os = normOS(os)
	arch = normArch(arch)

	ret := []*toolchain.Chain{}
	for _, t := range tt {
		if t.Compiler == "MSVC" {
			if os == "windows" && arch == normArch(t.VisualStudioArch) {
				ret = append(ret, t)
			}
		} else {
			full := strings.ToLower(t.Target)
			parts := strings.Split(full, "-")
			if len(parts) > 0 && arch == normArch(parts[0]) {
				match := false
				switch os {
				case "linux":
					match = strings.Contains(full, "linux-gnu") || strings.Contains(full, "linux")
				case "windows":
					match = strings.Contains(full, "mingw32") || strings.Contains(full, "cygwin") || strings.Contains(full, "w64") || strings.Contains(full, "w32")
				default:
					match = strings.Contains(full, os)
				}

				if match {
					ret = append(ret, t)
				}

			}
		}
	}
	return ret
}

func FindNative(tt []*toolchain.Chain) []*toolchain.Chain {
	os := runtime.GOOS
	arch := runtime.GOARCH
	return FindOsArch(os, arch, tt)
}

func Choose(tt []*toolchain.Chain, prefer_compiler []string) *toolchain.Chain {
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
			slices.SortFunc(sel, func(c1, c2 *toolchain.Chain) bool {
				return msvc.Compare(c1, c2) < 0
			})
		} else {
			slices.SortFunc(sel, func(c1, c2 *toolchain.Chain) bool {
				return gcc.Compare(c1, c2) < 0
			})
		}
		return sel[len(sel)-1]
	}

	if prefer_compiler == nil {
		prefer_compiler = []string{"gcc", "clang", "msvc"}
	}
	for _, c := range prefer_compiler {
		r := handle_compiler(c)
		if r != nil {
			return r
		}
	}
	sel := slices.Clone(tt)
	slices.SortFunc(sel, func(c1, c2 *toolchain.Chain) bool {
		if i := strings.Compare(c1.Compiler, c2.Compiler); i != 0 {
			return i < 0
		}
		if i := strings.Compare(c1.Target, c2.Target); i != 0 {
			return i < 0
		}
		return strings.Compare(c1.FullVersion, c2.FullVersion) < 0
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
