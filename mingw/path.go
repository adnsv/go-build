package mingw

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// ErrBadUsage is returned when a library is used on non-Windows hosts.
var ErrBadUsage = errors.New("msys/mingw utilities are to be used only on Windows hosts")

// ErrRootNotFound error is returned by FindRoot if it fails to
// locate the root directory.
var ErrRootNotFound = errors.New("failed to locate msys root directory")

// IsRoot returns true if dir is the root of msys/mingw tree
func IsRoot(dir string) bool {
	fn := filepath.Join(dir, "usr", "bin", "bash.exe")
	stat, err := os.Stat(fn)
	return err == nil && !stat.IsDir()
}

// FindRoot returns the root directory of the msys/mingw from a
// path somewhere inside it.
func FindRoot(fn string) (string, error) {
	if runtime.GOOS != "windows" {

	}
	p, err := filepath.Abs(fn)
	if err != nil {
		return "", err
	}

	if IsRoot(p) {
		return p, nil
	}

	for {
		d, b := filepath.Dir(p), filepath.Base(p)
		if b == "" || b == "\\" || b == "/" {
			break
		} else if IsRoot(d) {
			return d, nil
		} else {
			p = d
		}
	}
	return "", ErrRootNotFound
}

func PathRepresentation(path string) string {
	path = filepath.ToSlash(path)
	if filepath.IsAbs(path) && len(path) > 3 && path[1] == ':' && path[2] == '/' {
		drive := path[0]
		if drive <= 'a' {
			drive += 'a' - 'A'
			if drive >= 'a' && drive <= 'z' {
				path = "/" + string(drive) + path[2:]
			}
		}
	}
	path = strings.ReplaceAll(path, " ", "\\ ")
	return path
}
