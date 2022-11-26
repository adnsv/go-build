package cc

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/adnsv/go-build/compiler/clang"
	"github.com/adnsv/go-build/compiler/gcc"
	"github.com/adnsv/go-build/compiler/msvc"
	"github.com/adnsv/go-build/compiler/toolchain"
	"github.com/adnsv/go-utils/filesystem"
)

func FromEnv() (*Builder, error) {
	b := &Builder{Core: Core{
		Tools: toolchain.Toolset{},
	}}
	cc := os.Getenv("CC")
	cxx := os.Getenv("CXX")
	if cc == "" && cxx == "" {
		return nil, errors.New("missing CC/CXX environment vars")
	}

	compilers := map[string]struct{}{} // multiple compiler detector
	dirs := map[string]struct{}{}

	cc_infix := ""
	cxx_infix := ""

	if cc != "" {
		cc, err := filepath.Abs(cc)
		if err == nil {
			err = filesystem.ValidateFileExists(cc)
		}
		if err != nil {
			return nil, fmt.Errorf("invalid CC path: %w", err)
		}
		b.Tools[toolchain.CCompiler] = filepath.ToSlash(cc)
		exe := executableStem(cc)
		switch {
		case exe == "cl":
			b.Compiler = "MSVC"
			cc_infix = "cl"
		case strings.Contains(exe, "clang"):
			b.Compiler = "CLANG"
			cc_infix = "clang"
		case strings.Contains(exe, "gcc"):
			b.Compiler = "GCC"
			cc_infix = "gcc"
		}
		if b.Compiler != "" {
			compilers[b.Compiler] = struct{}{}
		}
		dirs[filepath.Dir(cc)] = struct{}{}
	}

	if cxx != "" {
		cxx, err := filepath.Abs(cxx)
		if err == nil {
			err = filesystem.ValidateFileExists(cxx)
		}
		if err != nil {
			return nil, fmt.Errorf("invalid CXX path: %w", err)
		}
		b.Tools[toolchain.CXXCompiler] = filepath.ToSlash(cxx)
		exe := executableStem(cxx)
		switch {
		case exe == "cl":
			b.Compiler = "MSVC"
			cxx_infix = "cl"
		case strings.Contains(exe, "clang"):
			b.Compiler = "CLANG"
			cxx_infix = "clang"
		case strings.Contains(exe, "g++"):
			b.Compiler = "GCC"
			cxx_infix = "g++"
		}
		if b.Compiler != "" {
			compilers[b.Compiler] = struct{}{}
		}
		dirs[filepath.Dir(cc)] = struct{}{}
	}

	if len(compilers) > 1 {
		return nil, fmt.Errorf("inconsistent compiler types detected")
	}

	if b.Compiler == "" || b.Compiler == "GCC" {
		var v *gcc.Ver
		var err error
		if cxx != "" {
			v, err = gcc.QueryVersion(cxx)
		} else {
			v, err = gcc.QueryVersion(cc)
		}
		if err == nil {
			b.Compiler = "GCC"
			b.CCIncludeDirs = v.CCIncludeDirs
			b.CXXIncludeDirs = v.CXXIncludeDirs
			b.Version = v.Version
			b.FullVersion = v.FullVersion
		}
	}
	if b.Compiler == "" || b.Compiler == "CLANG" {
		var v *clang.Ver
		var err error
		if cxx != "" {
			v, err = clang.QueryVersion(cxx)
		} else {
			v, err = clang.QueryVersion(cc)
		}
		if err == nil {
			b.Compiler = "CLANG"
			b.CCIncludeDirs = v.CCIncludeDirs
			b.CXXIncludeDirs = v.CXXIncludeDirs
			b.Version = v.Version
			b.FullVersion = v.FullVersion
		}
	}
	if b.Compiler == "" || b.Compiler == "MSVC" {
		var ver string
		//var target string
		var err error
		if cxx != "" {
			ver, _, err = msvc.QueryVersion(cxx)
		} else {
			ver, _, err = msvc.QueryVersion(cc)
		}
		if err == nil {
			b.Compiler = "MSVC"
			b.Version = ver
		}
	}

	if b.Compiler == "" {
		return nil, fmt.Errorf("unsupported compiler type")
	}

	if b.Compiler == "MSVC" {

	} else {
		check := func(tool toolchain.Tool, envvar string, names ...string) {
			if _, have := b.Tools[tool]; have {
				return
			}
			if envvar != "" {
				if v := os.Getenv(envvar); v != "" {
					b.Tools[tool] = v
					return
				}
			}
			if cxx_infix != "" {
				if fn := FindTool(cxx, cxx_infix, tool, names...); fn != "" {
					b.Tools[tool] = fn
					return
				}
			} else {
				if fn := FindTool(cxx, cc_infix, tool, names...); fn != "" {
					b.Tools[tool] = fn
					return
				}
			}
		}
		check(toolchain.CXXCompiler, "CXX", "g++", "c++")
		check(toolchain.Archiver, "AR", "ar", "gcc-ar")
		check(toolchain.ASMCompiler, "AS", "as", "gcc-as")
		check(toolchain.DLLLinker, "", "ld", "gcc-ld")
		check(toolchain.EXELinker, "", "ld", "gcc-ld")
		check(toolchain.OBJCopy, "", "objcopy", "gcc-objcopy")
		check(toolchain.OBJDump, "", "objdump", "gcc-objdump")
		check(toolchain.Runlib, "", "runlib", "gcc-runlib")
		check(toolchain.ResourceCompiler, "", "windres", "gcc-windres")
		check(toolchain.Strip, "", "strip", "gcc-strip")
	}

	return b, nil
}

func executableStem(fn string) string {
	fn = strings.ToLower(filepath.Base(fn))
	ext := filepath.Ext(fn)
	if ext == ".exe" {
		fn = strings.TrimSuffix(fn, ext)
	}
	return fn
}

func (b *Core) PrintSummary(w io.Writer) {
	fmt.Fprintf(w, "compiler: %s %s\n", b.Compiler, b.Version)
	if b.FullVersion != "" {
		fmt.Fprintf(w, "- full version: '%s'\n", b.FullVersion)
	}
}

func FindTool(base, infix string, tool toolchain.Tool, names ...string) string {
	i, n := strings.LastIndex(base, infix), len(infix)
	if i < 0 {
		return ""
	}

	for _, tn := range names {
		if t := base[:i] + tn + base[i+n:]; filesystem.FileExists(t) {
			return t
		}

	}
	return ""
}
