package gcc

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/adnsv/go-build/compiler/toolchain"
	"github.com/adnsv/go-build/compiler/triplet"
	"github.com/blang/semver/v4"
)

var reVersion = regexp.MustCompile("gcc version (.*?) .*")
var reTarget = regexp.MustCompile(`Target:\s+(.*)`)
var reThreadModel = regexp.MustCompile(`Thread model:\s+(.*)`)

func QueryVersion(exe string) (*Ver, error) {
	// First try to get version from predefined macros
	version, err := getVersionFromMacros(exe)
	if err != nil {
		// Fallback to parsing -v output
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
		version = strings.TrimSpace(match[1])
	}

	ret := &Ver{
		Version: version,
	}

	// Get additional information from -v output
	cmd := exec.Command(exe, "-v")
	buf, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}
	output := string(buf)
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		n := len(line)
		if n == 0 {
			continue
		}
		if line[n-1] == '\r' {
			line = line[:n-1]
		}
		match := reTarget.FindStringSubmatch(output)
		if len(match) == 2 {
			var err error
			ret.Target, err = triplet.ParseFull(strings.TrimSpace(match[1]))
			if err != nil {
				ret.Target = triplet.Full{Original: strings.TrimSpace(match[1])}
			}
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
			}
		}
	}

	// Get include paths with proper locale handling
	if ccIncludes, err := GetSystemIncludes(exe, "c"); err == nil {
		ret.CCIncludeDirs = ccIncludes
	}
	if cxxIncludes, err := GetSystemIncludes(exe, "c++"); err == nil {
		ret.CXXIncludeDirs = cxxIncludes
	}

	return ret, nil
}

func getVersionFromMacros(exe string) (string, error) {
	cmd := exec.Command(exe, "-dM", "-E", "-")
	cmd.Stdin = strings.NewReader("")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}

	var major, minor, patch string
	for _, line := range strings.Split(string(out), "\n") {
		switch {
		case strings.Contains(line, "__GNUC__"):
			major = strings.Fields(line)[2]
		case strings.Contains(line, "__GNUC_MINOR__"):
			minor = strings.Fields(line)[2]
		case strings.Contains(line, "__GNUC_PATCHLEVEL__"):
			patch = strings.Fields(line)[2]
		}
	}

	if major == "" {
		return "", errors.New("could not determine GCC version from macros")
	}

	if minor == "" {
		minor = "0"
	}
	if patch == "" {
		patch = "0"
	}

	return fmt.Sprintf("%s.%s.%s", major, minor, patch), nil
}

func GetSystemIncludes(exe, lang string) ([]string, error) {
	// Save current locale
	origLang := os.Getenv("LANG")
	origLCAll := os.Getenv("LC_ALL")
	os.Setenv("LANG", "C")
	os.Setenv("LC_ALL", "C")
	defer func() {
		os.Setenv("LANG", origLang)
		os.Setenv("LC_ALL", origLCAll)
	}()

	cmd := exec.Command(exe, "-x"+lang, "-E", "-v", "-")
	cmd.Stdin = strings.NewReader("")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}

	var includes []string
	collecting := false
	for _, line := range strings.Split(string(out), "\n") {
		if strings.Contains(line, "#include <...> search starts here:") {
			collecting = true
			continue
		}
		if strings.Contains(line, "End of search list.") {
			break
		}
		if collecting {
			path := strings.TrimSpace(line)
			if path != "" {
				path = fixWSLPath(path)
				includes = append(includes, path)
			}
		}
	}
	return includes, nil
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
	if i := strings.Compare(c1.Target.Original, c2.Target.Original); i != 0 {
		return i
	}
	if i := strings.Compare(c1.FullVersion, c2.FullVersion); i != 0 {
		return i
	}
	return strings.Compare(string(c1.Tools[toolchain.CXXCompiler]), string(c2.Tools[toolchain.CXXCompiler]))
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
