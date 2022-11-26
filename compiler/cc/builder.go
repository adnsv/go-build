package cc

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/adnsv/go-build/compiler/toolchain"
	"github.com/adnsv/go-utils/filesystem"
)

type Builder struct {
	Tools          toolchain.Toolset `json:"tools" yaml:"tools"`
	Compiler       string            `json:"compiler" yaml:"compiler"` // MSVC/GCC/CLANG
	CCIncludeDirs  []string          `json:"cc-include-dirs,omitempty" yaml:"cc-include-dirs,omitempty"`
	CXXIncludeDirs []string          `json:"cxx-include-dirs,omitempty" yaml:"cxx-include-dirs,omitempty"`
	LibraryDirs    []string          `json:"library-dirs,omitempty" yaml:"library-dirs,omitempty"`
	Environment    []string          `json:"environment" yaml:"environment"`
}

func FromChain(tc toolchain.Chain) *Builder {
	return &Builder{
		Tools:          tc.Tools,
		Compiler:       tc.Compiler,
		CCIncludeDirs:  tc.CCIncludeDirs,
		CXXIncludeDirs: tc.CXXIncludeDirs,
		LibraryDirs:    tc.LibraryDirs,
		Environment:    tc.Environment,
	}
}

func FromEnv() (*Builder, error) {
	b := &Builder{}
	cc := os.Getenv("cc")
	cxx := os.Getenv("cxx")
	if cc == "" && cxx == "" {
		return nil, errors.New("missing CC/CXX environment vars")
	}

	if cc != "" {
		cc, err := filepath.Abs(cc)
		if err == nil {
			err = filesystem.ValidateFileExists(cc)
		}
		if err != nil {
			return nil, fmt.Errorf("invalid CC path: %w", err)
		}
		b.Tools[toolchain.CCompiler] = filepath.ToSlash(cc)
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

	}

	return nil, nil
}
