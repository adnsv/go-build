package gcc

import (
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"runtime"
	"strings"

	"github.com/adnsv/go-build/compiler/toolchain"
	"github.com/blang/semver"
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
		const configuredWithPrefix = "Configured with: "
		if strings.HasPrefix(line, configuredWithPrefix) {
			line = strings.TrimPrefix(line, configuredWithPrefix)
			configs, err := parseConfig(line)
			if err == nil {
				ret.Languages = strings.Split(configs["enable-languages"], ",")
				ret.WithArch = configs["with-arch"]
			}
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

func Compare(c1, c2 *toolchain.Chain) int {
	v1, e1 := semver.ParseTolerant(c1.Version)
	v2, e2 := semver.ParseTolerant(c2.Version)
	if e1 == nil && e2 == nil {
		v1.Compare(v2)
	} else if e1 == nil {
		return -1
	} else if e2 == nil {
		return +1
	}
	if i := strings.Compare(c1.Target, c2.Target); i != 0 {
		return i
	}
	if i := strings.Compare(c1.FullVersion, c2.FullVersion); i != 0 {
		return i
	}
	return strings.Compare(c1.Tools[toolchain.CXXCompiler], c2.Tools[toolchain.CXXCompiler])
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

func parseConfig(s string) (map[string]string, error) {
	i, n := 0, len(s)
	for i < n && s[i] < ' ' {
		i++
	}

	key_char := func(c byte) bool {
		return c >= 'a' && c <= 'z' || c == '-' || c >= '0' && c <= '9'
	}
	m := map[string]string{}

	for i+2 < n {
		if s[i] != '-' || s[i+1] != '-' {
			i++
			continue
		}
		i += 2
		o := i
		for i < n && key_char(s[i]) {
			i++
		}
		key := s[o:i]
		val := ""
		if i < n && s[i] == '=' {
			i++
			o = i
			if i < n && s[i] == '\'' {
				i++
				o = i
				for i < n && s[i] != '\'' {
					i++
				}
				if i == n || s[i] != '\'' {
					return nil, fmt.Errorf("unterminated string literal")
				}
				val = s[o:i]
				i++
			} else {
				for i < n && s[i] != ' ' {
					i++
				}
				val = s[o:i]
			}
		}
		m[key] = val
	}
	return m, nil
}
