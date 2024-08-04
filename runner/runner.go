package runner

import (
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"strings"
)

type runner struct {
	Executable  string
	WorkDir     string
	Stdout      io.Writer
	Stderr      io.Writer
	Environment []string
}

func (r *runner) run(args ...string) error {
	c := exec.Command(r.Executable, args...)
	c.Dir = r.WorkDir
	c.Env = r.Environment
	c.Stdout = r.Stderr
	c.Stderr = r.Stderr
	if c.Stdout != nil {
		fmt.Fprintln(c.Stdout, c)
	}
	err := c.Run()
	return err
}

type CC struct {
	runner
	Flags    []string
	CXXFlags []string
}

func (cc *CC) Make(src string, obj string) error {
	ff := append(cc.Flags, []string{}...)
	ext := strings.ToLower(filepath.Ext(src))
	if ext != ".c" {
		ff = append(ff, cc.CXXFlags...)
	}
	ff = append(ff, "/Fo"+obj, src)
	return cc.run(ff...)
}

type AR struct {
	runner
	Flags []string
}

func (ar *AR) Make(aout string, objs []string) error {
	ff := append(ar.Flags, objs...)
	ff = append(ff, "/OUT:"+aout)
	return ar.run(ff...)
}

type LD struct {
	runner
	Flags []string
}

func (ld *LD) Make(exe string, subsystem string, libs []string) error {
	ff := append(ld.Flags, []string{}...)
	if subsystem != "" {
		ff = append(ff, "/SUBSYSTEM:"+subsystem)
	}
	//ff = append(ff, "/IMPLIB:"+mainlib)
	ff = append(ff, libs...)
	ff = append(ff, "/OUT:"+exe)
	return ld.run(ff...)
}

type RC struct {
	runner
	Flags []string
}

func (rc *RC) Make(resout string, rcin string) error {
	ff := append(rc.Flags, "/Fo", resout, rcin)
	return rc.run(ff...)
}

type MT struct {
	runner
	Flags []string
}

func (mt *MT) Make(targetexe string, manifest string) error {
	ff := append(mt.Flags, "-manifest", manifest, "-outputresource:"+targetexe+";1")
	return mt.run(ff...)
}
