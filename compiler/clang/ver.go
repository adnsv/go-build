package clang

import (
	"errors"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/adnsv/go-build/compiler/gcc"
)

var reVersion = regexp.MustCompile(`^(?:Apple LLVM|.*clang) version ([\S]*).*`)
var reTarget = regexp.MustCompile(`Target:\s+(.*)`)
var reThreadModel = regexp.MustCompile(`Thread model:\s+(.*)`)
var reInstalledDir = regexp.MustCompile(`InstalledDir:\s+(.*)`)

func QueryVersion(exe string) (*Ver, error) {
	cmd := exec.Command(exe, "-v")
	buf, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}
	output := string(buf)
	lines := strings.Split(output, "\n")
	if len(lines) == 0 {
		return nil, errors.New("invalid version output")
	}

	match := reVersion.FindStringSubmatch(strings.TrimSpace(lines[0]))
	if len(match) != 2 {
		return nil, errors.New("invalid version output")
	}

	v := match[1]
	if i := strings.IndexByte(v, '-'); i >= 0 {
		v = v[:i]
	}

	ret := &Ver{
		FullVersion: strings.TrimSpace(match[0]),
		Version:     v,
	}
	for _, line := range lines {
		n := len(line)
		if n == 0 {
			continue
		}
		if line[n-1] == '\r' {
			line = line[:n-1]
		}
		match = reTarget.FindStringSubmatch(output)
		if len(match) == 2 {
			ret.Target = strings.TrimSpace(match[1])
		}
		match = reThreadModel.FindStringSubmatch(output)
		if len(match) == 2 {
			ret.ThreadModel = strings.TrimSpace(match[1])
		}
		match = reInstalledDir.FindStringSubmatch(output)
		if len(match) == 2 {
			ret.InstalledDir = filepath.ToSlash(strings.TrimSpace(match[1]))
		}
	}
	ret.IncludeDirs = append(ret.IncludeDirs, gcc.ExtractIncludePaths(exe, "c")...)
	ret.IncludeDirs = append(ret.IncludeDirs, gcc.ExtractIncludePaths(exe, "c++")...)
	return ret, nil
}
