package gopackage

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/adnsv/go-build/osarch"
)

// BuildCommand prepares build command for building golang package
func BuildCommand(exepath string, pkgpath string, oa *osarch.Pair, args ...string) *exec.Cmd {
	aa := []string{"build"}
	aa = append(aa, args...)
	aa = append(aa, "-o", exepath)

	// make sure go build recognizes relative paths outside of the gopath
	pkgpath = filepath.ToSlash(pkgpath)
	if strings.HasPrefix(pkgpath, ".") && !strings.HasPrefix(pkgpath, "..") && !strings.HasPrefix(pkgpath, "./") {
		pkgpath = "./" + pkgpath
	}
	aa = append(aa, pkgpath)
	cmd := exec.Command("go", aa...)
	cmd.Env = append(os.Environ(), "GOOS="+oa.OS, "GOARCH="+oa.Arch)
	return cmd
}

// LDFlagsFromVariables creates LD args for injecting custom variables
// into a package builder
func LDFlagsFromVariables(vars map[string]string) string {
	ret := make([]string, 0, len(vars))
	for n, v := range vars {
		ret = append(ret, fmt.Sprintf("-X '%s=%s'", n, v))
	}
	sort.Strings(ret)
	return "-ldflags=" + strings.Join(ret, " ")
}
