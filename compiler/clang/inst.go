package clang

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"

	"github.com/adnsv/go-utils/filesystem"
	"github.com/blang/semver/v4"
)

var reCLANG = regexp.MustCompile(`^clang(-\d+(\.\d+(\.\d+)?)?)?(\.exe)?$`)

func DiscoverInstallations(feedback func(string)) []*Installation {
	if feedback != nil {
		feedback("discovering clang installations")
	}

	search_paths := filepath.SplitList(os.Getenv("PATH"))

	if runtime.GOOS == "windows" {
		if f := os.Getenv("LLVM_ROOT"); filesystem.DirExists(f) {
			search_paths = append(search_paths, filepath.Join(f, "LLVM", "bin"))
		}
		if f := os.Getenv("ProgramFiles(x86)"); filesystem.DirExists(f) {
			search_paths = append(search_paths, filepath.Join(f, "LLVM", "bin"))
		}
		if f := os.Getenv("ProgramFiles"); filesystem.DirExists(f) {
			search_paths = append(search_paths, filepath.Join(f, "LLVM", "bin"))
		}
	}

	files := filesystem.SearchFilesAndSymlinks(search_paths,
		func(fi os.FileInfo) bool {
			return reCLANG.MatchString(fi.Name())
		})
	if len(files) == 0 {
		return nil
	}

	// group things together if they report the same signature size + version
	vcs := map[string]*vcollect{}
	for fn, symlinks := range files {
		ver, err := QueryVersion(fn)
		if err != nil {
			continue
		}
		sigstr := ver.FullVersion + ver.Version + ver.Target.Original + ver.ThreadModel +
			strings.Join(ver.CCIncludeDirs, "|") + "#" +
			strings.Join(ver.CXXIncludeDirs, "|")

		vc := vcs[sigstr]
		if vc == nil {
			vc = &vcollect{
				files:    make(map[string]struct{}),
				symlinks: make(map[string]struct{}),
				ver:      ver,
			}
		}
		vc.files[fn] = struct{}{}
		for _, sl := range symlinks {
			vc.symlinks[sl] = struct{}{}
		}
		vcs[sigstr] = vc
	}

	// collect results
	ret := []*Installation{}
	for _, vc := range vcs {
		if vc.ver == nil {
			continue
		}
		inst := &Installation{Ver: *vc.ver}
		for v := range vc.files {
			inst.CCompiler.OtherPaths = append(inst.CCompiler.OtherPaths, v)
		}
		for v := range vc.symlinks {
			inst.CCompiler.SymLinks = append(inst.CCompiler.SymLinks, v)
		}

		inst.CCompiler.ChoosePrimaryCCompilerPath(inst.Target.Original, "clang", inst.Version)
		ret = append(ret, inst)
	}

	// sort, latest versions first
	sort.SliceStable(ret, func(i, j int) bool {
		v1, e1 := semver.ParseTolerant(ret[i].Ver.Version)
		v2, e2 := semver.ParseTolerant(ret[j].Ver.Version)
		if e1 == nil && e2 == nil {
			return v1.GT(v2)
		} else if e1 != nil {
			return true
		} else if e2 != nil {
			return false
		}
		return ret[i].Ver.Version > ret[j].Ver.Version
	})

	if feedback != nil {
		feedback(fmt.Sprintf("found %d clang installation(s)", len(ret)))
	}
	return ret
}

type vcollect struct {
	files    map[string]struct{}
	symlinks map[string]struct{}
	ver      *Ver
}
