package clang

import (
	"fmt"
	"path/filepath"
	"sort"

	"github.com/adnsv/go-build/compiler/toolchain"
	"github.com/adnsv/go-utils/filesystem"
)

func DiscoverToolchains(wantCxx bool, feedback func(string)) []*toolchain.Chain {

	installations := DiscoverInstallations(feedback)
	toolchains := []*toolchain.Chain{}

	for _, inst := range installations {
		tc := &toolchain.Chain{
			Compiler:       "CLANG",
			Version:        inst.Version,
			FullVersion:    inst.FullVersion,
			Target:         inst.Target,
			ThreadModel:    inst.ThreadModel,
			InstalledDir:   filepath.ToSlash(inst.InstalledDir),
			CCIncludeDirs:  inst.CCIncludeDirs,
			CXXIncludeDirs: inst.CXXIncludeDirs,
			Tools:          map[toolchain.Tool]string{},
		}
		if feedback != nil {
			feedback(fmt.Sprintf("scanning clang %s targeting %s at %s",
				tc.FullVersion, tc.Target.Original, inst.CCompiler.PrimaryPath))
		}
		tc.Tools[toolchain.CCompiler] = inst.CCompiler.PrimaryPath
		tc.Tools[toolchain.CXXCompiler] = inst.CCompiler.PrimaryPath
		checkTool := func(tool toolchain.Tool, names ...string) {
			fn := inst.CCompiler.FindTool("gcc", names...)
			if fn != "" {
				tc.Tools[tool] = filepath.ToSlash(fn)
			}
		}

		checkTool(toolchain.CXXCompiler, "clang++")
		checkTool(toolchain.Archiver, "llvm-ar", "ar")
		checkTool(toolchain.ASMCompiler, "llvm-as", "as")
		checkTool(toolchain.DLLLinker, "lld")
		checkTool(toolchain.EXELinker, "lld")
		checkTool(toolchain.OBJCopy, "objcopy", "llvm-objcopy")
		checkTool(toolchain.OBJDump, "objdump", "llvm-objdump")
		checkTool(toolchain.Runlib, "runlib", "llvm-runlib")

		em := map[string]string{}
		if v := tc.Tools[toolchain.CCompiler]; v != "" {
			em["CC"] = v
		}
		if v := tc.Tools[toolchain.CXXCompiler]; v != "" {
			em["CXX"] = v
		}
		em["C_INCLUDE_PATH"] = filesystem.JoinPathList(tc.CCIncludeDirs...)
		em["CPLUS_INCLUDE_PATH"] = filesystem.JoinPathList(tc.CXXIncludeDirs...)
		for k, v := range em {
			tc.Environment = append(tc.Environment, fmt.Sprintf("%s=%s", k, v))
		}
		sort.Strings(tc.Environment)

		toolchains = append(toolchains, tc)
	}

	return toolchains
}
