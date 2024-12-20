package gcc

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/adnsv/go-utils/filesystem"
	"github.com/blang/semver/v4"
)

var reGCC = regexp.MustCompile(`^((?:\w+-)*)gcc(?:-\d+(?:\.\d+)*)?(?:\.exe)?$`)

func DiscoverInstallations(feedback func(string)) []*Installation {
	if feedback != nil {
		feedback("discovering gcc installations")
	}

	// First check environment variables
	if envCompiler := getCompilerFromEnv(); envCompiler != "" {
		if feedback != nil {
			feedback(fmt.Sprintf("checking compiler from environment: %s", envCompiler))
		}
		if ver, err := QueryVersion(envCompiler); err == nil {
			inst := &Installation{Ver: *ver}
			inst.CCompiler.OtherPaths = []string{envCompiler}
			inst.CCompiler.ChoosePrimaryCCompilerPath(inst.Target.Original, "gcc", inst.Version, ToolNames)
			return []*Installation{inst}
		}
	}

	// Then search in PATH
	files := filesystem.SearchFilesAndSymlinks(filepath.SplitList(os.Getenv("PATH")),
		func(fi os.FileInfo) bool {
			fn := fi.Name()
			if !strings.Contains(fn, "gcc") {
				return false
			}
			ss := reGCC.FindStringSubmatch(fn)
			if len(ss) != 2 {
				return false
			}
			for _, p := range strings.Split(ss[1], "-") {
				if p == "gfortran" {
					return false
				}
			}
			return true
		})
	if len(files) == 0 {
		return nil
	}

	// Group compilers by their signature
	vcs := map[string]*vcollect{}
	for fn, symlinks := range files {
		ver, err := QueryVersion(fn)
		if err != nil {
			continue
		}

		// Extract toolchain prefix
		prefix := detectToolchainPrefix(fn)
		if prefix != "" {
			ver.ToolchainPrefix = prefix
		}

		sigstr := ver.Version + ver.Target.Original + ver.ThreadModel +
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

	// Collect results
	ret := []*Installation{}
	for _, vc := range vcs {
		if vc.ver == nil {
			continue
		}
		inst := &Installation{Ver: *vc.ver}
		for v := range vc.files {
			inst.CCompiler.OtherPaths = append(inst.CCompiler.OtherPaths, fixWSLPath(v))
		}
		for v := range vc.symlinks {
			inst.CCompiler.SymLinks = append(inst.CCompiler.SymLinks, fixWSLPath(v))
		}
		inst.CCompiler.ChoosePrimaryCCompilerPath(inst.Target.Original, "gcc", inst.Version, ToolNames)
		ret = append(ret, inst)
	}

	// Sort, latest versions first
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
		feedback(fmt.Sprintf("found %d gcc installation(s)", len(ret)))
	}
	return ret
}

func getCompilerFromEnv() string {
	if cc := os.Getenv("CC"); cc != "" {
		if path, err := exec.LookPath(cc); err == nil {
			return path
		}
	}
	return ""
}

func detectToolchainPrefix(compilerPath string) string {
	base := filepath.Base(compilerPath)
	if match := regexp.MustCompile(`^(.*-)gcc`).FindStringSubmatch(base); len(match) > 1 {
		// Remove llvm- prefix if present
		prefix := match[1]
		prefix = strings.TrimPrefix(prefix, "llvm-")
		return prefix
	}
	return ""
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
