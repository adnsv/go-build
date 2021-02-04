package clang

import (
	"fmt"
	"path/filepath"

	"github.com/adnsv/go-build/compiler/toolchain"
	"github.com/adnsv/go-utils/fs"
)

func DiscoverToolchains(wantCxx bool, feedback func(string)) []*toolchain.Chain {

	installations := DiscoverInstallations(feedback)
	toolchains := []*toolchain.Chain{}

	for _, inst := range installations {
		tc := &toolchain.Chain{
			Compiler:     "CLANG",
			Version:      inst.Version,
			FullVersion:  inst.FullVersion,
			Target:       inst.Target,
			ThreadModel:  inst.ThreadModel,
			InstalledDir: filepath.ToSlash(inst.InstalledDir),
			IncludeDirs:  fs.NormalizePathsToSlash(inst.IncludeDirs),
			Tools:        map[toolchain.Tool]string{},
		}
		if feedback != nil {
			feedback(fmt.Sprintf("scanning clang %s targeting %s at %s",
				tc.FullVersion, tc.Target, inst.CCompiler.PrimaryPath))
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

		toolchains = append(toolchains, tc)
	}

	return toolchains
}
