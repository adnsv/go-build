package proj

import (
	"fmt"
	"io"
)

type CMakeOptions struct {
	CMakeMinimumRequired     string
	CMakeCXXStandard         string
	CMakeCXXStandardRequired bool
}

func GenerateCMake(w io.Writer, p Project, opts CMakeOptions) {
	ln := func(s ...interface{}) {
		fmt.Fprintln(w, s...)
	}
	lf := func(s string, v ...interface{}) {
		fmt.Fprintf(w, s, v...)
	}

	lf("cmake_minimum_required(VERSION %s)", opts.CMakeMinimumRequired)
	if opts.CMakeCXXStandard != "" {
		ln()
		lf("set (CMAKE_CXX_STANDARD %s)", opts.CMakeCXXStandard)
		lf("set(CMAKE_CXX_STANDARD_REQUIRED %s", boolStr(opts.CMakeCXXStandardRequired))
	}

	ln()
	lf("project(%s)", p.Name)
}

func boolStr(v bool) string {
	if v {
		return "ON"
	}
	return "OFF"
}
