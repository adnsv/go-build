package clang

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

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

		{
			n := inst.CCompiler.PrimaryPath
			infix := "clang"
			i := strings.LastIndex(n, infix)
			if i >= 0 {
				prefix := n[:i]
				postfix := n[i+len(infix):]
				tt := toolchain.FindTools(prefix, postfix, ToolNames)
				for tool, path := range tt {
					if _, exists := tc.Tools[tool]; !exists {
						tc.Tools[tool] = path
					}
				}
				tt = toolchain.FindTools(prefix+"llvm", postfix, ToolNames)
				for tool, path := range tt {
					if _, exists := tc.Tools[tool]; !exists {
						tc.Tools[tool] = path
					}
				}
			}
		}

		if !tc.Tools.Contains(toolchain.CXXCompiler) {
			tc.Tools[toolchain.CXXCompiler] = tc.Tools[toolchain.CCompiler]
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

var ToolNames = map[string]toolchain.Tool{
	"clang":        toolchain.CCompiler,
	"clang++":      toolchain.CXXCompiler,
	"ar":           toolchain.Archiver,
	"as":           toolchain.ASMCompiler,
	"lld":          toolchain.Linker,
	"objcopy":      toolchain.OBJCopy,
	"objdump":      toolchain.OBJDump,
	"ranlib":       toolchain.Ranlib,
	"windres":      toolchain.ResourceCompiler,
	"strip":        toolchain.Strip,
	"llvm-ar":      toolchain.Archiver,
	"llvm-as":      toolchain.ASMCompiler,
	"llvm-objcopy": toolchain.OBJCopy,
	"llvm-objdump": toolchain.OBJDump,
	"llvm-ranlib":  toolchain.Ranlib,
	"llvm-windres": toolchain.ResourceCompiler,
	"llvm-strip":   toolchain.Strip,
}

var ToolEnvs = map[string]toolchain.Tool{
	"CC":  toolchain.CCompiler,
	"CXX": toolchain.CXXCompiler,
}
