package clang

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"

	"github.com/adnsv/go-build/compiler/toolchain"
	"github.com/adnsv/go-utils/filesystem"
	"github.com/blang/semver/v4"
)

// Implementation-specific filename patterns
var (
	reClangFilename = regexp.MustCompile(`^clang(?:-\d+(?:\.\d+)*)?(?:\.exe)?$`)
	reEmccFilename  = regexp.MustCompile(`^em(?:cc|c\+\+)(?:\.exe)?$`)
	reZigFilename   = regexp.MustCompile(`^zig(?:\.exe)?$`)
)

func DiscoverInstallations(feedback func(string)) []*Installation {
	if feedback != nil {
		feedback("discovering LLVM-based compiler installations")
	}

	search_paths := filepath.SplitList(os.Getenv("PATH"))

	// Add implementation-specific paths
	if runtime.GOOS == "windows" {
		// LLVM paths
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

	// Collect all potential compiler executables
	files := filesystem.SearchFilesAndSymlinks(search_paths,
		func(fi os.FileInfo) bool {
			fn := fi.Name()
			return reClangFilename.MatchString(fn) ||
				reEmccFilename.MatchString(fn) ||
				reZigFilename.MatchString(fn)
		})
	if len(files) == 0 {
		return nil
	}

	// Group by implementation and version
	vcs := map[string]*vcollect{}
	for fn, symlinks := range files {
		// Special handling for Zig
		var ver *Ver
		var err error
		if reZigFilename.MatchString(filepath.Base(fn)) {
			// Try as Zig's C compiler
			tool := toolchain.NewToolPath(fn, "cc")
			ver, err = QueryVersionWithRegex(tool, ZigClang, reZigVersion)
		} else {
			tool := toolchain.ToolPath(fn)
			ver, err = QueryVersion(tool)
		}
		if err != nil {
			continue
		}

		sigstr := string(ver.Implementation) + ver.FullVersion + ver.Version + ver.Target.Original + ver.ThreadModel +
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
		vc.files[filepath.ToSlash(fn)] = struct{}{}
		for _, sl := range symlinks {
			vc.symlinks[filepath.ToSlash(sl)] = struct{}{}
		}
		vcs[sigstr] = vc
	}

	// Collect results
	ret := []*Installation{}
	for _, vc := range vcs {
		if vc.ver == nil {
			continue
		}
		inst := &Installation{Ver: *vc.ver}

		// Add paths with implementation-specific handling
		for v := range vc.files {
			path := fixWSLPath(filepath.ToSlash(v))
			inst.CCompiler.OtherPaths = append(inst.CCompiler.OtherPaths, path)
		}
		for v := range vc.symlinks {
			path := fixWSLPath(filepath.ToSlash(v))
			inst.CCompiler.SymLinks = append(inst.CCompiler.SymLinks, path)
		}

		// Choose appropriate compiler name based on implementation
		compilerName := "clang"
		switch inst.Ver.Implementation {
		case EmScripten:
			compilerName = "emcc"
		case ZigClang:
			compilerName = "zig"
		}
		inst.CCompiler.ChoosePrimaryCCompilerPath(inst.Target.Original, compilerName, inst.Version, ToolNames)
		ret = append(ret, inst)
	}

	// Sort by version, latest first
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
		feedback(fmt.Sprintf("found %d LLVM-based compiler installation(s)", len(ret)))
	}
	return ret
}

type vcollect struct {
	files    map[string]struct{}
	symlinks map[string]struct{}
	ver      *Ver
}

// fixWSLPath converts WSL paths (/mnt/c/...) to Windows paths (C:/...)
func fixWSLPath(p string) string {
	if strings.HasPrefix(p, "/mnt/") {
		drive := string(p[5])
		rest := p[6:]
		return drive + ":" + rest
	}
	return p
}
