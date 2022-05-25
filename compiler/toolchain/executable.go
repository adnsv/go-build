package toolchain

import (
	"path/filepath"
	"strings"

	"github.com/adnsv/go-utils/filesystem"
)

type Executable struct {
	PrimaryPath string   `json:"primary-path"`
	OtherPaths  []string `json:"alternative-paths,omitempty"`
	SymLinks    []string `json:"symlinks,omitempty"`
}

func (x *Executable) ChoosePrimaryCCompilerPath(target, cc, version string) {
	if len(x.OtherPaths) == 0 {
		return
	}
	if len(x.OtherPaths) == 1 {
		x.PrimaryPath = filepath.ToSlash(x.OtherPaths[0])
		x.OtherPaths = x.OtherPaths[:0]
	}

	i := FindBestString(x.OtherPaths, func(fn string) int {
		return CCompilerScore(target, cc, version, fn)
	})

	if i >= 0 {
		x.PrimaryPath = filepath.ToSlash(x.OtherPaths[i])
		x.OtherPaths[i] = "" // will get removed in the NormalizePathsToSlash call below
	}
	x.OtherPaths = filesystem.NormalizePathsToSlash(x.OtherPaths)
}

func FindBestString(ss []string, scorer func(string) int) int {
	bestScore := 0
	bestIndex := -1
	for i, s := range ss {
		score := scorer(s)
		if bestIndex < 0 || score > bestScore {
			bestScore = score
			bestIndex = i
		}
	}
	return bestIndex
}

func CCompilerScore(target, cc, version, fn string) int {
	// fn -> dir, base
	d, b := filepath.Dir(fn), filepath.Base(fn)

	// normalized name part
	n := strings.ToLower(b)
	if filepath.Ext(n) == ".exe" {
		n = n[:len(n)-4]
	}

	score := len(fn) // prefer longer paths

	if cc == "gcc" {
		// gcc/g++ gcc/c++ pairs
		if strings.Contains(b, cc) {
			if gxx := filepath.Join(d, strings.Replace(b, "gcc", "g++", 1)); filesystem.FileExists(gxx) {
				score += 64000
			} else if cxx := filepath.Join(d, strings.Replace(b, "gcc", "c++", 1)); filesystem.FileExists(cxx) {
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
	if strings.Contains(d, target) {
		score += 2000
	}
	if strings.Contains(d, version) {
		score += 1000
	}

	return score
}

func (x *Executable) FindTool(cc string, names ...string) string {
	for _, tn := range names {
		fn := x.PrimaryPath
		if i := strings.LastIndex(fn, cc); i >= 0 {
			if t := fn[:i] + tn + fn[i+3:]; filesystem.FileExists(t) {
				return t
			}
		}
		for _, fn = range x.OtherPaths {
			if i := strings.LastIndex(fn, cc); i >= 0 {
				if t := fn[:i] + tn + fn[i+3:]; filesystem.FileExists(t) {
					return t
				}
			}
		}
		for _, fn = range x.SymLinks {
			if i := strings.LastIndex(fn, cc); i >= 0 {
				if t := fn[:i] + tn + fn[i+3:]; filesystem.FileExists(t) {
					return t
				}
			}
		}
	}
	return ""
}
