package gcc

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
			Compiler:       "GCC",
			Version:        inst.Version,
			FullVersion:    inst.FullVersion,
			Target:         inst.Target,
			ThreadModel:    inst.ThreadModel,
			InstalledDir:   filepath.ToSlash(filepath.Dir(inst.CCompiler.PrimaryPath)),
			CCIncludeDirs:  inst.CCIncludeDirs,
			CXXIncludeDirs: inst.CXXIncludeDirs,
			Tools:          map[toolchain.Tool]string{},
		}

		if feedback != nil {
			feedback(fmt.Sprintf("scanning gcc %s targeting %s at %s",
				tc.Version, tc.Target.Original, inst.CCompiler.PrimaryPath))
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
		checkTool(toolchain.Ranlib, "ranlib", "gcc-ranlib")
		checkTool(toolchain.ResourceCompiler, "windres", "gcc-windres")
		checkTool(toolchain.Strip, "strip", "gcc-strip")

		if wantCxx && !haveCxx {
			continue
		}

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
