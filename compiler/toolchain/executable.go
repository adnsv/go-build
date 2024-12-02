package toolchain

import (
	"path/filepath"
	"sort"
	"strings"

	"github.com/adnsv/go-utils/filesystem"
	"golang.org/x/exp/maps"
)

type Executable struct {
	PrimaryPath string   `json:"primary-path"`
	Subcommands []string `json:"subcommands,omitempty"` // optional subcommands for the primary path (Zig cc)
	OtherPaths  []string `json:"alternative-paths,omitempty"`
	SymLinks    []string `json:"symlinks,omitempty"`
}

func (x *Executable) ChoosePrimaryCCompilerPath(target string, cc string, version string, toolnames map[string]Tool) {
	paths := make(map[string]struct{}, len(x.OtherPaths)+len(x.SymLinks))
	other_paths := make(map[string]struct{}, len(x.OtherPaths))
	sym_links := make(map[string]struct{}, len(x.SymLinks))

	for _, p := range x.OtherPaths {
		fn := filepath.ToSlash(p)
		paths[fn] = struct{}{}
		other_paths[fn] = struct{}{}
	}
	for _, p := range x.SymLinks {
		fn := filepath.ToSlash(p)
		paths[fn] = struct{}{}
		sym_links[fn] = struct{}{}
	}

	if len(paths) < 2 {
		for p := range paths {
			x.PrimaryPath = p
			x.OtherPaths = nil
			x.SymLinks = nil
			return
		}
	}

	s := FindBestString(paths, func(fn string) int {
		return CCompilerScore(target, cc, version, fn, toolnames)
	})

	x.PrimaryPath = s
	delete(other_paths, s)
	delete(sym_links, s)

	x.OtherPaths = maps.Keys(other_paths)
	x.SymLinks = maps.Keys(sym_links)
	sort.Strings(x.OtherPaths)
	sort.Strings(x.SymLinks)
}

func FindBestString(m map[string]struct{}, scorer func(string) int) string {
	bestScore := -1
	bestString := ""
	for s := range m {
		score := scorer(s)
		if score > bestScore {
			bestScore = score
			bestString = s
		}
	}
	return bestString
}

func CollectTools(fn, infix string, toolnames map[string]Tool) Toolset {
	dir, base := filepath.Split(fn)
	tools := Toolset{}

	if i := strings.LastIndex(base, infix); i >= 0 {
		prefix := dir + base[:i]
		postfix := base[i+len(infix):]
		for t, path := range FindTools(prefix, postfix, toolnames) {
			if !tools.Contains(t) {
				tools[t] = path
			}
		}
	}
	if infix == "clang" {
		infix = "llvm"
		if i := strings.LastIndex(base, infix); i >= 0 {
			prefix := dir + base[:i]
			postfix := base[i+len(infix):]
			for t, path := range FindTools(prefix, postfix, toolnames) {
				if !tools.Contains(t) {
					tools[t] = path
				}
			}
		}
	}
	return tools
}

func CCompilerScore(target, cc, version, fn string, toolnames map[string]Tool) int {
	// fn -> dir, base
	dir, base := filepath.Split(fn)
	score := len(fn) // prefer longer paths (slightly)

	{ // # of available tools is our main metric
		tools := CollectTools(fn, cc, toolnames)
		score += len(tools) * 128000
	}

	// normalized name part
	n := strings.ToLower(base)
	if filepath.Ext(n) == ".exe" {
		n = n[:len(n)-4]
	}

	if cc == "gcc" {
		// gcc/g++ gcc/c++ pairs
		if strings.Contains(base, cc) {
			if gxx := filepath.Join(dir, strings.Replace(base, "gcc", "g++", 1)); filesystem.FileExists(gxx) {
				score += 64000
			} else if cxx := filepath.Join(dir, strings.Replace(base, "gcc", "c++", 1)); filesystem.FileExists(cxx) {
				score += 64000
			}
		}
	}

	if strings.HasPrefix(n, target) {
		// target-cc
		score += 32000
		if strings.HasSuffix(n, version) {
			// target-cc-version a lesser preference than target-cc
			score -= 16000
		}
	} else if strings.HasSuffix(n, version) {
		// cc-version
		score += 8000
	} else if strings.HasSuffix(n, version) {
		// version is somewhere in the name
		score += 4000
	}

	// target/version in dir
	if strings.Contains(dir, target) {
		score += 2000
	}
	if strings.Contains(dir, version) {
		score += 1000
	}

	return score
}

func FindTools(path_prefix, path_postfix string, toolnames map[string]Tool) Toolset {
	ret := Toolset{}
	for name, tool := range toolnames {
		fn := path_prefix + name + path_postfix
		if filesystem.FileExists(fn) {
			ret[tool] = ToolPath(filepath.ToSlash(fn))
			continue
		}
	}
	return ret
}
