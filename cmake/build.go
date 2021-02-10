package cmake

import (
	"io"
	"os"
	"os/exec"
	"strings"
)

type BuildType int

const (
	Release = BuildType(iota)
	Debug
	MinSizeRel
	RelWithDebInfo
)

func (b BuildType) String() string {
	switch b {
	case Release:
		return "Release"
	case Debug:
		return "Debug"
	case MinSizeRel:
		return "MinSizeRel"
	case RelWithDebInfo:
		return "RelWithDebInfo"
	default:
		return "<INVALID-BuildType>"
	}
}

type Builder struct {
	CMakeCmd string
	Stdout   io.Writer
	Stderr   io.Writer

	SourceDir string
	BuildDir  string
	BuildType BuildType // injected as -DCMAKE_BUILD_TYPE:STRING=${BuildType}

	GenerateFlags []string // -DVAR=VALUE pairs
	BuildFlags    []string

	Generator *Generator
}

type Generator = struct {
	Name    string
	Arch    string
	Toolset string
}

func NinjaGenerator() *Generator {
	return &Generator{Name: "Ninja"}
}

func MSVCGenerator(name string, arch string, toolset string) *Generator {
	if arch != "" {
		ss := strings.Split(arch, "_")
		if len(ss) == 2 {
			arch = ss[1]
		}
		m, ok := map[string]string{
			"64":    "x64",
			"x64":   "x64",
			"amd64": "x64",
			"86":    "Win32",
			"x86":   "Win32",
			"386":   "Win32",
			"486":   "Win32",
			"586":   "Win32",
			"686":   "Win32",
			"arm":   "ARM",
			"arm32": "ARM",
			"arm64": "ARM64",
		}[strings.ToLower(arch)]
		if ok {
			arch = m
		}
	}
	if name == "" {
		name = "Visual Studio 16 2019"
	}
	return &Generator{
		Name:    name,
		Arch:    arch,
		Toolset: toolset,
	}
}

func (b *Builder) EffectiveGenerateArgs() []string {
	args := []string{}
	args = append(args, "-S", b.SourceDir)
	args = append(args, "-B", b.BuildDir)
	args = append(args, "-DCMAKE_BUILD_TYPE:STRING="+b.BuildType.String())
	args = append(args, b.GenerateFlags...)
	if b.Generator != nil && b.Generator.Name != "" {
		args = append(args, "-G", b.Generator.Name)
		if b.Generator.Arch != "" {
			args = append(args, "-A", b.Generator.Arch)
		}
		if b.Generator.Toolset != "" {
			args = append(args, "-T", b.Generator.Toolset)
		}
	}
	return args
}

func (b *Builder) EffectiveBuildArgs() []string {
	args := []string{}
	args = append(args, "--build")
	args = append(args, b.BuildDir)
	args = append(args, "--config", b.BuildType.String())
	args = append(args, "--parallel")
	return args
}

func (b *Builder) Generate() error {
	c := b.CMakeCmd
	if c == "" {
		c = "cmake"
	}
	cmd := exec.Command(c, b.EffectiveGenerateArgs()...)
	cmd.Stdout = b.Stdout
	cmd.Stderr = b.Stderr
	return cmd.Run()
}

func (b *Builder) Build() error {
	c := b.CMakeCmd
	if c == "" {
		c = "cmake"
	}
	cmd := exec.Command("cmake", b.EffectiveBuildArgs()...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
