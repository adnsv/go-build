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

type Builder struct {
	Compiler       string            `json:"compiler" yaml:"compiler"` // MSVC/GCC/CLANG
	Version        string            `json:"version" yaml:"version"`
	FullVersion    string            `json:"full-version,omitempty" yaml:"full-version,omitempty"`
	Tools          toolchain.Toolset `json:"tools" yaml:"tools"`
	CCIncludeDirs  []string          `json:"cc-include-dirs,omitempty" yaml:"cc-include-dirs,omitempty"`
	CXXIncludeDirs []string          `json:"cxx-include-dirs,omitempty" yaml:"cxx-include-dirs,omitempty"`
	LibraryDirs    []string          `json:"library-dirs,omitempty" yaml:"library-dirs,omitempty"`
	Environment    []string          `json:"environment,omitempty" yaml:"environment,omitempty"`
}

func FromChain(tc toolchain.Chain) *Builder {
	return &Builder{
		Tools:          tc.Tools,
		Compiler:       tc.Compiler,
		Version:        tc.Version,
		FullVersion:    tc.FullVersion,
		CCIncludeDirs:  tc.CCIncludeDirs,
		CXXIncludeDirs: tc.CXXIncludeDirs,
		LibraryDirs:    tc.LibraryDirs,
		Environment:    tc.Environment,
	}
}

func FromEnv() (*Builder, error) {
	b := &Builder{
		Tools: toolchain.Toolset{},
	}
	cc := os.Getenv("cc")
	cxx := os.Getenv("cxx")
	if cc == "" && cxx == "" {
		return nil, errors.New("missing CC/CXX environment vars")
	}

	compilers := map[string]struct{}{} // multiple compiler detector

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
		case exe == "clang":
			b.Compiler = "CLANG"
		case strings.HasSuffix(exe, "gcc"):
			b.Compiler = "GCC"
		}
		if b.Compiler != "" {
			compilers[b.Compiler] = struct{}{}
		}
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
		exe := executableStem(cc)
		switch {
		case exe == "cl":
			b.Compiler = "MSVC"
		case exe == "clang":
			b.Compiler = "CLANG"
		case strings.HasSuffix(exe, "g++"):
			b.Compiler = "GCC"
		}
		if b.Compiler != "" {
			compilers[b.Compiler] = struct{}{}
		}
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

func (b *Builder) PrintSummary(w io.Writer) {
	fmt.Fprintf(w, "compiler: %s %s\n", b.Compiler, b.Version)
	if b.FullVersion != "" {
		fmt.Fprintf(w, "- full version: '%s'\n", b.FullVersion)
	}
}
