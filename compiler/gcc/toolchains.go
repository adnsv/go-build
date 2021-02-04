package gcc

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
			Compiler:     "GCC",
			Version:      inst.Version,
			FullVersion:  inst.FullVersion,
			Target:       inst.Target,
			ThreadModel:  inst.ThreadModel,
			InstalledDir: filepath.ToSlash(filepath.Dir(inst.CCompiler.PrimaryPath)),
			IncludeDirs:  fs.NormalizePathsToSlash(inst.IncludeDirs),
			Tools:        map[toolchain.Tool]string{},
		}

		if feedback != nil {
			feedback(fmt.Sprintf("scanning gcc %s targeting %s at %s",
				tc.Version, tc.Target, inst.CCompiler.PrimaryPath))
		}
		tc.Tools[toolchain.CCompiler] = inst.CCompiler.PrimaryPath

		checkTool := func(tool toolchain.Tool, names ...string) bool {
			fn := inst.CCompiler.FindTool("gcc", names...)
			if fn != "" {
				tc.Tools[tool] = filepath.ToSlash(fn)
				return true
			}
			return false
		}

		haveCxx := checkTool(toolchain.CXXCompiler, "g++", "c++")
		checkTool(toolchain.Archiver, "ar", "gcc-ar")
		checkTool(toolchain.ASMCompiler, "as", "gcc-as")
		checkTool(toolchain.DLLLinker, "ld", "gcc-ld")
		checkTool(toolchain.EXELinker, "ld", "gcc-ld")
		checkTool(toolchain.OBJCopy, "objcopy", "gcc-objcopy")
		checkTool(toolchain.OBJDump, "objdump", "gcc-objdump")
		checkTool(toolchain.Runlib, "runlib", "gcc-runlib")
		checkTool(toolchain.ResourceCompiler, "windres", "gcc-windres")
		checkTool(toolchain.Strip, "strip", "gcc-strip")

		if wantCxx && !haveCxx {
			continue
		}

		toolchains = append(toolchains, tc)
	}

	return toolchains
}
