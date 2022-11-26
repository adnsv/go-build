package cc

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/adnsv/go-build/compiler/toolchain"
)

type Core struct {
	Compiler       string            `json:"compiler" yaml:"compiler"` // MSVC/GCC/CLANG
	Version        string            `json:"version" yaml:"version"`
	FullVersion    string            `json:"full-version,omitempty" yaml:"full-version,omitempty"`
	Tools          toolchain.Toolset `json:"tools" yaml:"tools"`
	CCIncludeDirs  []string          `json:"cc-include-dirs,omitempty" yaml:"cc-include-dirs,omitempty"`
	CXXIncludeDirs []string          `json:"cxx-include-dirs,omitempty" yaml:"cxx-include-dirs,omitempty"`
	LibraryDirs    []string          `json:"library-dirs,omitempty" yaml:"library-dirs,omitempty"`
	Environment    []string          `json:"environment,omitempty" yaml:"environment,omitempty"`
}

func BindObjMSVC(obj_fn, src string) []string {
	return []string{"/Fo" + obj_fn, src}
}

func BindObjGCC(obj_fn, src string) []string {
	return []string{"-o", obj_fn, "-c", src}
}

func BindArMSVC(ar_fn string, obj_fns ...string) []string {
	ret := []string{"/OUT:" + ar_fn}
	ret = append(ret, obj_fns...)
	return ret
}

func BindArGCC(ar_fn string, obj_fns ...string) []string {
	ret := []string{ar_fn}
	ret = append(ret, obj_fns...)
	return ret
}

func BindIncludeDirMSVC(dir string) string {
	return "/I" + dir
}

func BindIncludeDirGCC(dir string) string {
	return "-I" + dir
}

/*

func BindCompilePdbMSVC(pdb_fn string) []string {
	return []string{"/Fd" + pdb_fn, "/FS", "/Zf"}
}

func BindLinkPdbMSVC(pdb_fn string) []string {
	return []string{"/PDB:" + pdb_fn}
}

func BindCompilePdbDummy(string) []string {
	return nil
}

func BindLinkPdbDummy(string) []string {
	return nil
}
*/

type Builder struct {
	Core

	FlagsC   Flags // c-specific compiler flags to produce obj files
	FlagsCXX Flags // c++-specific compiler flags to produce obj files
	FlagsAR  Flags
	FlagsDLL Flags
	FlagsEXE Flags
	FlagsRC  Flags

	ExtOBJ string
	ExtAR  string
	ExtDLL string
	ExtEXE string

	BindOBJ        func(obj_fn, src string) []string
	BindAR         func(ar_fn string, obj_fns ...string) []string
	BindCompilePDB func(pdb_fn string) []string
	BindLinkPDB    func(pdb_fn string) []string
	BindIncludeDir func(dir string) string

	//	wantPDB bool

	WorkDir string
	Stdout  io.Writer
	Stderr  io.Writer
}

func FromChain(tc toolchain.Chain) *Builder {
	return &Builder{Core: Core{
		Tools:          tc.Tools,
		Compiler:       tc.Compiler,
		Version:        tc.Version,
		FullVersion:    tc.FullVersion,
		CCIncludeDirs:  tc.CCIncludeDirs,
		CXXIncludeDirs: tc.CXXIncludeDirs,
		LibraryDirs:    tc.LibraryDirs,
		Environment:    tc.Environment,
	}}
}

type BuilderOption int

const (
	BuildMsvcRuntimeDLL = BuilderOption(1 << iota)
	//	BuildWithPDB
	C11
	C17
	CXX14
	CXX17
	CXX20
	CXXLatest
)

func (b *Builder) ConfigureDefaults(options ...BuilderOption) {

	c11 := false
	c17 := false
	cxx14 := false
	cxx17 := false
	cxx20 := false
	cxxLatest := false

	with_mt := true
	for _, opt := range options {
		switch opt {
		case BuildMsvcRuntimeDLL:
			with_mt = false
			//		case BuildWithPDB:
			//			b.wantPDB = true
		case C11:
			c11 = true
		case C17:
			c17 = true
		case CXX14:
			cxx14 = true
		case CXX17:
			cxx17 = true
		case CXX20:
			cxx20 = true
		case CXXLatest:
			cxxLatest = true
		}
	}

	if b.Compiler == "MSVC" {
		b.ExtOBJ = ".obj"
		b.ExtAR = ".lib"
		b.ExtDLL = ".dll"
		b.ExtEXE = ".exe"

		b.BindOBJ = BindObjMSVC
		b.BindAR = BindArMSVC
		//		b.BindCompilePDB = BindCompilePdbMSVC
		//		b.BindLinkPDB = BindLinkPdbMSVC
		b.BindIncludeDir = BindIncludeDirMSVC

		b.FlagsC.Add(All, "/nologo", "/DWIN32", "/D_WINDOWS")
		b.FlagsC.Add(Debug, "/Zi", "/Ob0", "/Od", "/RTC1")
		b.FlagsC.Add(Release, "/O2", "/Ob2", "/DNDEBUG")
		b.FlagsC.Add(MinSizeRel, "/O1", "/Ob1", "/DNDEBUG")
		b.FlagsC.Add(RelWithDebInfo, "/Zi", "/O2", "/Ob1", "/DNDEBUG")

		b.FlagsCXX.Add(All, "/nologo", "/DWIN32", "/D_WINDOWS", "/EHsc")
		b.FlagsCXX.Add(Debug, "/Zi", "/Ob0", "/Od", "/RTC1")
		b.FlagsCXX.Add(Release, "/O2", "/Ob2", "/DNDEBUG")
		b.FlagsCXX.Add(MinSizeRel, "/O1", "/Ob1", "/DNDEBUG")
		b.FlagsCXX.Add(RelWithDebInfo, "/Zi", "/O2", "/Ob1", "/DNDEBUG")

		if with_mt {
			b.FlagsC.Add(Debug, "/MTd")
			b.FlagsC.Add(Release, "/MT")
			b.FlagsC.Add(MinSizeRel, "/MT")
			b.FlagsC.Add(RelWithDebInfo, "/MTd")
			b.FlagsCXX.Add(Debug, "/MTd")
			b.FlagsCXX.Add(Release, "/MT")
			b.FlagsCXX.Add(MinSizeRel, "/MT")
			b.FlagsCXX.Add(RelWithDebInfo, "/MTd")
		}
		if cxxLatest {
			b.FlagsCXX.Add(All, "/std:c++latest")
		} else if cxx20 {
			b.FlagsCXX.Add(All, "/std:c++20")
		} else if cxx17 {
			b.FlagsCXX.Add(All, "/std:c++17")
		} else if cxx14 {
			b.FlagsCXX.Add(All, "/std:c++14")
		}
		if c17 {
			b.FlagsC.Add(All, "/std:c17")
		} else if c11 {
			b.FlagsC.Add(All, "/std:c11")
		}

		b.FlagsAR.Add(All, "/machine:x64")

		b.FlagsDLL.Add(All, "/machine:x64")
		b.FlagsDLL.Add(Debug, "/debug", "/INCREMENTAL")
		b.FlagsDLL.Add(Release, "/INCREMENTAL:NO")
		b.FlagsDLL.Add(MinSizeRel, "/INCREMENTAL:NO")
		b.FlagsDLL.Add(RelWithDebInfo, "/debug", "/INCREMENTAL")

		b.FlagsEXE.Add(All, "/machine:x64")
		b.FlagsEXE.Add(Debug, "/debug", "/INCREMENTAL")
		b.FlagsEXE.Add(Release, "/INCREMENTAL:NO")
		b.FlagsEXE.Add(MinSizeRel, "/INCREMENTAL:NO")
		b.FlagsEXE.Add(RelWithDebInfo, "/debug", "/INCREMENTAL")

		b.FlagsRC.Add(All, "-DWIN32")
		b.FlagsRC.Add(Debug, "-D_DEBUG")

	} else {
		b.ExtOBJ = ".o"
		b.ExtAR = ".a"
		b.ExtDLL = ".so"
		b.ExtEXE = ""

		b.BindOBJ = BindObjGCC
		b.BindAR = BindArGCC
		//		b.BindCompilePDB = BindCompilePdbDummy
		//		b.BindLinkPDB = BindLinkPdbDummy
		b.BindIncludeDir = BindIncludeDirMSVC

		b.FlagsC.Add(All)
		b.FlagsC.Add(Debug, "-g")
		b.FlagsC.Add(Release, "-O3", "-DNDEBUG")
		b.FlagsC.Add(MinSizeRel, "-Os", "-DNDEBUG")
		b.FlagsC.Add(RelWithDebInfo, "-O2", "-g", "-DNDEBUG")

		b.FlagsCXX.Add(All)
		b.FlagsCXX.Add(Debug, "-g")
		b.FlagsCXX.Add(Release, "-O3", "-DNDEBUG")
		b.FlagsCXX.Add(MinSizeRel, "-Os", "-DNDEBUG")
		b.FlagsCXX.Add(RelWithDebInfo, "-O2", "-g", "-DNDEBUG")

		b.FlagsAR.Add(All)

		b.FlagsDLL.Add(All)

		b.FlagsEXE.Add(All)

		b.FlagsRC.Add(All)
	}
}

type BuildContext struct {
	BuildConfig BuildConfig
	FlagsC      []string
	FlagsCXX    []string
	FlagsAR     []string
	FlagsDLL    []string
	FlagsEXE    []string
	FlagsRC     []string

	SrcDir      string // all source files are specified relative to this dir
	ObjDir      string // all object files are placed here in subdirs
	LibDir      string
	LibraryDirs []string
	Environment []string

	PDBfn  string
	Stdout io.Writer
	Stderr io.Writer
}

func (b *Builder) NewBuildContext(cfg BuildConfig, srcdir, objdir, libdir string) (*BuildContext, error) {
	if cfg == All {
		return nil, fmt.Errorf("invalid build configuration")
	}

	cfg_flags := func(f Flags) []string {
		r := FlagSet{}
		r.Insert(f[All])
		r.Insert(f[cfg])
		flags := make([]string, 0, len(r))
		for flag := range r {
			flags = append(flags, flag)
		}
		sort.Strings(flags)
		return flags
	}

	ctx := &BuildContext{
		BuildConfig: cfg,
		SrcDir:      srcdir,
		ObjDir:      objdir,
		LibDir:      libdir,
	}
	ctx.FlagsC = append(ctx.FlagsC, b.CCIncludeDirs...)
	ctx.FlagsC = append(ctx.FlagsC, cfg_flags(b.FlagsC)...)
	ctx.FlagsCXX = append(ctx.FlagsC, b.CXXIncludeDirs...)
	ctx.FlagsCXX = append(ctx.FlagsC, cfg_flags(b.FlagsCXX)...)
	ctx.FlagsC = cfg_flags(b.FlagsC)
	ctx.FlagsCXX = cfg_flags(b.FlagsCXX)
	ctx.FlagsAR = cfg_flags(b.FlagsAR)
	ctx.FlagsDLL = cfg_flags(b.FlagsDLL)
	ctx.FlagsEXE = cfg_flags(b.FlagsEXE)
	ctx.FlagsRC = cfg_flags(b.FlagsRC)

	ctx.LibraryDirs = b.LibraryDirs
	ctx.Environment = b.Environment
	if len(ctx.Environment) == 0 {
		ctx.Environment = os.Environ()
	}

	return ctx, nil
}

//func (b *Builder) ConfigPDB(ctx *BuildContext, pdb_fn string) {
//	ctx.FlagsC = append(ctx.FlagsC, b.BindPDB(pdb_fn)...)
//	ctx.FlagsCXX = append(ctx.FlagsCXX, b.BindCompilePDB(pdb_fn)...)
//	ctx.FlagsDLL = append(ctx.FlagsEXE, b.BindLinkerPDB(pdb_fn)...)
//	ctx.FlagsEXE = append(ctx.FlagsEXE, b.BindLinkerPDB(pdb_fn)...)
//}

func (b *Builder) Compile(ctx *BuildContext, src_fn, obj_fn string) error {
	ext := strings.ToLower(filepath.Ext(src_fn))
	var flags []string
	exe := ""
	switch ext {
	case ".c":
		flags = ctx.FlagsC
		exe = b.Tools[toolchain.CCompiler]

	case ".cxx", ".cc", ".cpp":
		flags = ctx.FlagsCXX
		exe = b.Tools[toolchain.CCompiler]

	default:
		return fmt.Errorf("unsupported file extension in '%s'", src_fn)
	}

	args := flags

	cmd := exec.Command(exe, args...)
	cmd.Dir = ctx.SrcDir
	cmd.Env = ctx.Environment
	cmd.Stdout = ctx.Stdout
	cmd.Stderr = ctx.Stderr
	err := cmd.Run()
	return err
}

func (b *Builder) CompileLibrary(ctx *BuildContext, src_fns ...string) error {
	obj_fns := []string{}
	for _, src_fn := range src_fns {
		obj_fn := filepath.Join(ctx.SrcDir, src_fn+b.ExtOBJ)
		obj_fns = append(obj_fns, obj_fn)
		err := b.Compile(ctx, src_fn, obj_fn)
		if err != nil {
			return err
		}
	}
	for _, obj_fn := range obj_fns {
		fmt.Printf("- %s\n", obj_fn)
	}
	return nil
}
