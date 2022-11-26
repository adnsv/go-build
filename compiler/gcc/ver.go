package gcc

import (
	"errors"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
)

var reVersion = regexp.MustCompile("gcc version (.*?) .*")
var reTarget = regexp.MustCompile(`Target:\s+(.*)`)
var reThreadModel = regexp.MustCompile(`Thread model:\s+(.*)`)

func QueryVersion(exe string) (*Ver, error) {
	cmd := exec.Command(exe, "-v")
	buf, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}
	output := string(buf)
	lines := strings.Split(output, "\n")
	for len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	if len(lines) == 0 {
		return nil, errors.New("invalid version output")
	}

	match := reVersion.FindStringSubmatch(strings.TrimSpace(lines[len(lines)-1]))
	if len(match) != 2 {
		return nil, errors.New("invalid version output")
	}

	ret := &Ver{
		FullVersion: strings.TrimSpace(match[0]),
		Version:     strings.TrimSpace(match[1]),
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
	}
	ret.CCIncludeDirs = append(ret.CCIncludeDirs, ExtractIncludePaths(exe, "c")...)
	ret.CXXIncludeDirs = append(ret.CXXIncludeDirs, ExtractIncludePaths(exe, "c++")...)
	return ret, nil
}

func ExtractIncludePaths(exe string, lang string) []string {
	cmd := exec.Command(exe, "-x"+lang, "-E", "-v", "-")
	buf, err := cmd.CombinedOutput()
	if err != nil {
		return nil
	}
	lines := strings.Split(string(buf), "\n")
	ret := []string{}
	includeLine := false

	fixpath := func(s string) string { return s }
	if runtime.GOOS != "windows" && strings.HasPrefix(exe, "/mnt/") {
		fixpath = fixWSLpath
	}

	for _, line := range lines {
		line = strings.TrimRight(line, "\r")
		if includeLine {
			if line == "End of search list." {
				includeLine = false
				continue
			}
			line = strings.TrimSpace(line)
			line = fixpath(line)
			ret = append(ret, line)
		} else if line == "#include <...> search starts here:" {
			includeLine = true
		}
	}
	return ret
}

func fixWSLpath(p string) string {
	if len(p) < 3 {
		return p
	}
	if p[1] == ':' && (p[2] == '\\' || p[2] == '/') {
		ret := "/mnt/" + strings.ToLower(p[:1]) + "/" + strings.ReplaceAll(p[3:], "\\", "/")
		return ret
	} else {
		return p
	}
}
