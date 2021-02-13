package env

import (
	"runtime"
	"sort"
	"strings"
)

func Split(lines []string) map[string]string {
	ret := make(map[string]string, len(lines))
	for _, ln := range lines {
		p := strings.IndexByte(ln, '=')
		if p >= 0 {
			k := ln[:p]
			v := ln[p+1:]
			ret[strings.TrimSpace(k)] = strings.TrimSpace(v)
		}
	}
	return ret
}

func Join(m map[string]string) []string {
	ret := make([]string, 0, len(m))
	for k, v := range m {
		ret = append(ret, k+"="+v)
	}
	sort.Strings(ret)
	return ret

}

func Merge(a, b map[string]string) map[string]string {
	ret := make(map[string]string, len(a))

	if runtime.GOOS == "windows" {
		for k, v := range a {
			ret[strings.ToUpper(k)] = v
		}
		for k, v := range b {
			ret[strings.ToUpper(k)] = v
		}

	} else {
		for k, v := range a {
			ret[k] = v
		}
		for k, v := range b {
			ret[k] = v
		}
	}

	return ret
}
