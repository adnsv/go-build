package gcc

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/adnsv/go-build/compiler/toolchain"
	"github.com/adnsv/go-utils/filesystem"
)

func DiscoverToolchains(feedback func(string)) []*toolchain.Chain {
	installations := DiscoverInstallations(feedback)
	toolchains := []*toolchain.Chain{}

	for _, inst := range installations {
		tc := &toolchain.Chain{
			Compiler:       "gcc",
			Implementation: "gcc",
			Version:        inst.Version,
			FullVersion:    inst.FullVersion,
			Target:         inst.Target,
			ThreadModel:    inst.ThreadModel,
			InstalledDir:   filepath.ToSlash(filepath.Dir(inst.CCompiler.PrimaryPath)),
			CCIncludeDirs:  inst.CCIncludeDirs,
			CXXIncludeDirs: inst.CXXIncludeDirs,
			Tools:          toolchain.Toolset{},
		}

		if feedback != nil {
			feedback(fmt.Sprintf("scanning gcc %s targeting %s at %s",
				tc.FullVersion, tc.Target.Original, inst.CCompiler.PrimaryPath))
		}
		tc.Tools[toolchain.CCompiler] = toolchain.ToolPath(inst.CCompiler.PrimaryPath)

		{
			n := inst.CCompiler.PrimaryPath
			infix := "gcc"
			i := strings.LastIndex(n, infix)
			if i >= 0 {
				prefix := n[:i]
				postfix := n[i+len(infix):]
				tt := toolchain.FindTools(prefix, postfix, ToolNames)
				for tool, path := range tt {
					tc.Tools[tool] = path
				}
			}
		}

		if !tc.Tools.Contains(toolchain.CXXCompiler) {
			n := inst.CCompiler.PrimaryPath
			infix := "gcc"
			i := strings.LastIndex(n, infix)
			if i >= 0 {
				prefix := n[:i]
				postfix := n[i+len(infix):]
				tt := toolchain.FindTools(prefix+"g++", postfix, ToolNames)
				for tool, path := range tt {
					if _, exists := tc.Tools[tool]; !exists {
						tc.Tools[tool] = path
					}
				}
			}
		}

		em := map[string]string{}
		if v := tc.Tools[toolchain.CCompiler]; v != "" {
			em["CC"] = v.Path()
		}
		if v := tc.Tools[toolchain.CXXCompiler]; v != "" {
			em["CXX"] = v.Path()
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

func findBinUtils(prefix, version string) map[toolchain.Tool]string {
	utils := map[toolchain.Tool]string{}

	// Try version-specific names first
	versionedPrefix := fmt.Sprintf("%sgcc-%s-", prefix, version)
	for name, tool := range map[string]toolchain.Tool{
		"ar":      toolchain.Archiver,
		"ranlib":  toolchain.Ranlib,
		"objcopy": toolchain.OBJCopy,
		"objdump": toolchain.OBJDump,
		"strip":   toolchain.Strip,
	} {
		if path, err := exec.LookPath(versionedPrefix + name); err == nil {
			utils[tool] = path
			continue
		}
		// Fallback to non-versioned
		if path, err := exec.LookPath(prefix + name); err == nil {
			utils[tool] = path
		}
	}
	return utils
}

var ToolNames = map[string]toolchain.Tool{
	"gcc":         toolchain.CCompiler,
	"g++":         toolchain.CXXCompiler,
	"c++":         toolchain.CXXCompiler,
	"cpp":         toolchain.CXXCompiler,
	"ar":          toolchain.Archiver,
	"as":          toolchain.ASMCompiler,
	"ld":          toolchain.Linker,
	"objcopy":     toolchain.OBJCopy,
	"objdump":     toolchain.OBJDump,
	"ranlib":      toolchain.Ranlib,
	"windres":     toolchain.ResourceCompiler,
	"strip":       toolchain.Strip,
	"gcc-ar":      toolchain.Archiver,
	"gcc-as":      toolchain.ASMCompiler,
	"gcc-ld":      toolchain.Linker,
	"gcc-objcopy": toolchain.OBJCopy,
	"gcc-objdump": toolchain.OBJDump,
	"gcc-ranlib":  toolchain.Ranlib,
	"gcc-windres": toolchain.ResourceCompiler,
	"gcc-strip":   toolchain.Strip,
}

var ToolEnvs = map[string]toolchain.Tool{
	"CC":  toolchain.CCompiler,
	"CXX": toolchain.CXXCompiler,
}
