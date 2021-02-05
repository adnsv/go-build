package msvc

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/adnsv/go-build/compiler/toolchain"
	"github.com/adnsv/go-utils/fs"
)

const vswhereSubpath = "Microsoft Visual Studio/Installer/vswhere.exe"

func DiscoverInstallations(feedback func(string)) ([]*Installation, error) {
	if feedback != nil {
		feedback("discovering msvc installations")
	}
	paths := []string{}
	addPath := func(s string) {
		for _, p := range paths {
			if p == s {
				return
			}
		}
		paths = append(paths, s)
	}

	if pf := os.Getenv("ProgramFiles(x86)"); fs.DirExists(pf) {
		addPath(filepath.Join(pf, vswhereSubpath))
	}
	if pf := os.Getenv("ProgramFiles"); fs.DirExists(pf) {
		addPath(filepath.Join(pf, vswhereSubpath))
	}
	if fs.DirExists("C:\\Program Files (x86)") {
		addPath(filepath.Join("C:\\Program Files (x86)", vswhereSubpath))
	}
	addPath("vswhere.exe")

	vswherePath := ""
	for _, path := range paths {
		if fs.FileExists(path) {
			vswherePath = path
			break
		}
	}
	if vswherePath == "" {
		return nil, errors.New("failed to find vswhere.exe")
	}

	if feedback != nil {
		feedback(fmt.Sprintf("using vswhere utility: %s", vswherePath))
	}

	cmd := exec.Command(vswherePath, "-all", "-format", "json", "-products", "*", "-legacy", "-prerelease")
	buf, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	fmt.Printf("WSHERE: %s\n%s\n", cmd.String(), string(buf))

	installations := []*Installation{}
	err = json.Unmarshal(buf, &installations)
	if err != nil {
		return nil, fmt.Errorf("failed to parse vswhere output: %s", err)
	}

	for _, i := range installations {
		i.InstallationPath = filepath.ToSlash(i.InstallationPath)
	}

	// sort, latest versions first
	sort.SliceStable(installations, func(i, j int) bool {

		v1, e1 := ParseVersionQuad(installations[i].InstallationVersion)
		v2, e2 := ParseVersionQuad(installations[j].InstallationVersion)
		if e1 == nil && e2 == nil {
			return v1.Compare(v2) > 0
		} else if e1 != nil {
			return true
		} else if e2 != nil {
			return false
		}
		return installations[i].InstallationVersion > installations[j].InstallationVersion
	})

	if feedback != nil {
		feedback(fmt.Sprintf("found %d msvc installation(s)", len(installations)))
	}

	return installations, nil
}

type VersionQuad struct {
	Major int
	Minor int
	Patch int
	Build int
}

var ErrInvalidVersionQuad = errors.New("invalid version quad")

func ParseVersionQuad(s string) (v VersionQuad, err error) {
	n, err := fmt.Sscanf(s, "%d.%d.%d.%d", &v.Major, &v.Major, &v.Patch, &v.Build)
	if n != 4 {
		err = ErrInvalidVersionQuad
	}
	return
}

func (v VersionQuad) Compare(o VersionQuad) int {
	if v.Major != o.Major {
		if v.Major > o.Major {
			return 1
		}
		return -1
	}
	if v.Minor != o.Minor {
		if v.Minor > o.Minor {
			return 1
		}
		return -1
	}
	if v.Patch != o.Patch {
		if v.Patch > o.Patch {
			return 1
		}
		return -1
	}
	if v.Build != o.Build {
		if v.Build > o.Build {
			return 1
		}
		return -1
	}
	return 0
}

var msvcEnvVars = []string{
	"CL",
	"_CL_",
	"INCLUDE",
	"LIBPATH",
	"LINK",
	"_LINK_",
	"LIB",
	"PATH",
	"TMP",
	"FRAMEWORKDIR",
	"FRAMEWORKDIR64",
	"FRAMEWORKVERSION",
	"FRAMEWORKVERSION64",
	"UCRTCONTEXTROOT",
	"UCRTVERSION",
	"UNIVERSALCRTSDKDIR",
	"VCINSTALLDIR",
	"VCTARGETSPATH",
	"WINDOWSLIBPATH",
	"WINDOWSSDKDIR",
	"WINDOWSSDKLIBVERSION",
	"WINDOWSSDKVERSION",
	"VISUALSTUDIOVERSION",
}

func TestArches(inst *Installation, feedback func(string)) []*toolchain.Chain {
	toolchains := []*toolchain.Chain{}

	commonDir := filepath.Join(inst.InstallationPath, "Common7", "tools")
	devbat := filepath.Join(inst.InstallationPath, "VC", "Auxiliary", "Build", "vcvarsall.bat")

	majorVer := func(s string) int {
		vv := strings.Split(s, ".")
		if len(vv) > 1 {
			v, err := strconv.Atoi(vv[0])
			if err == nil {
				return v
			}
		}
		return 0
	}(inst.InstallationVersion)

	if feedback != nil {
		feedback(fmt.Sprintf("testing installation: %s", inst.InstallationPath))
	}

	if !fs.FileExists(devbat) {
		devbat = filepath.Join(inst.InstallationPath, "VC", "vcvarsall.bat")
		if !fs.FileExists(devbat) {
			devbat = ""
			if feedback != nil {
				feedback("ERROR: failed to locate vcvarsall.bat file")
			}
			return nil
		}
	}

	targetArches := []string{"x86", "amd64", "arm", "arm64"}
	hostArch := runtime.GOARCH
	combArches := make([]string, len(targetArches))
	for i, targetArch := range targetArches {
		if targetArch == hostArch {
			combArches[i] = targetArch
		} else {
			combArches[i] = hostArch + "_" + targetArch
		}
	}

	vvs := make([]map[string]string, len(targetArches))

	var wg sync.WaitGroup
	for i, arch := range combArches {
		wg.Add(1)
		go func(i int, arch string) {
			defer wg.Done()
			v, err := CollectBatVars(devbat, arch, strconv.Itoa(majorVer), commonDir)
			if err == nil {
				vvs[i] = v
			}
		}(i, arch)
	}
	wg.Wait()

	for i, arch := range targetArches {
		vars := vvs[i]
		if len(vars) == 0 {
			continue
		}

		if feedback != nil {
			feedback(fmt.Sprintf("%s: architecture %s - supported", inst.DisplayName, arch))
		}

		archname := ""
		switch arch {
		case "amd64":
			archname = "x64"
		case "x86":
			archname = "x32"
		default:
			archname = arch
		}

		tc := &toolchain.Chain{
			Compiler:            "MSVC",
			FullVersion:         fmt.Sprintf("%s - %s - %s", inst.DisplayName, archname, inst.InstallationVersion),
			Version:             inst.InstallationVersion,
			Target:              "Windows",
			InstalledDir:        filepath.ToSlash(inst.InstallationPath),
			VisualStudioID:      inst.InstanceID,
			VisualStudioArch:    arch,
			VisualStudioVersion: vars["VISUALSTUDIOVERSION"],
			UCRTVersion:         vars["UCRTVERSION"],
			Tools:               map[toolchain.Tool]string{},
		}

		if s := vars["WINDOWSSDKVERSION"]; s != "" {
			tc.WindowsSDKVersion = strings.TrimRight(s, `\`)
		}

		if incs := strings.Split(vars["INCLUDE"], ";"); len(incs) > 0 {
			for _, v := range incs {
				if v != "" && fs.DirExists(v) {
					tc.IncludeDirs = append(tc.IncludeDirs, filepath.ToSlash(v))
				}
			}
		}

		if libs := strings.Split(vars["LIB"], ";"); len(libs) > 0 {
			for _, v := range libs {
				if v != "" {
					tc.LibraryDirs = append(tc.LibraryDirs, filepath.ToSlash(v))
				}
			}
		}

		paths := strings.Split(vars["PATH"], ";")

		if s := vars["CL"]; s != "" {
			tc.Tools[toolchain.CCompiler] = filepath.ToSlash(s)
			tc.Tools[toolchain.CXXCompiler] = filepath.ToSlash(s)
		} else {
			for _, path := range paths {
				fn := filepath.Join(path, "cl.exe")
				if fs.FileExists(fn) {
					tc.Tools[toolchain.CCompiler] = filepath.ToSlash(fn)
					tc.Tools[toolchain.CXXCompiler] = filepath.ToSlash(fn)
					break
				}
			}
		}

		if s := vars["LINK"]; s != "" {
			tc.Tools[toolchain.DLLLinker] = filepath.ToSlash(s)
			tc.Tools[toolchain.EXELinker] = filepath.ToSlash(s)
		} else {
			for _, path := range paths {
				fn := filepath.Join(path, "link.exe")
				if fs.FileExists(fn) {
					tc.Tools[toolchain.DLLLinker] = filepath.ToSlash(fn)
					tc.Tools[toolchain.EXELinker] = filepath.ToSlash(fn)
					break
				}
			}
		}

		ar := filepath.Join(filepath.Dir(tc.Tools[toolchain.DLLLinker]), "lib.exe")
		if !fs.FileExists(ar) {
			for _, path := range paths {
				fn := filepath.Join(path, "lib.exe")
				if fs.FileExists(fn) {
					ar = fn
					break
				}
			}
		}
		if ar != "" {
			tc.Tools[toolchain.Archiver] = filepath.ToSlash(ar)
		}

		for _, path := range paths {
			fn := filepath.Join(path, "rc.exe")
			if fs.FileExists(fn) {
				tc.Tools[toolchain.ResourceCompiler] = filepath.ToSlash(fn)
				break
			}
		}

		toolchains = append(toolchains, tc)
	}

	return toolchains
}

func CollectBatVars(devbat string, arg string, majorVer string, commonDir string) (map[string]string, error) {
	ret := map[string]string{}
	fn := "test.bat"
	batfname := "vs-cmt-" + fn
	envfname := batfname + ".env"
	bat := []string{
		`@echo off`,
		`cd /d "%~dp0"`,
		`set "VS` + majorVer + `0COMNTOOLS=` + commonDir + `"`,
		`call "` + devbat + `" ` + arg + ` || exit`,
	}
	for _, v := range msvcEnvVars {
		bat = append(bat, fmt.Sprintf("echo %s := %%%s%% >> %s\n", v, v, envfname))
	}
	tmpdir, err := ioutil.TempDir("", "bushido")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tmpdir)
	batpath := filepath.Join(tmpdir, batfname)
	envpath := filepath.Join(tmpdir, envfname)
	err = ioutil.WriteFile(batpath, []byte(strings.Join(bat, "\r\n")), 0777)
	if err != nil {
		return nil, err
	}

	err = exec.Command("cmd", "/C", batpath).Run()
	if err != nil {
		return nil, err
	}

	buf, err := ioutil.ReadFile(envpath)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(buf), "\n")
	for _, line := range lines {
		p := strings.Index(line, ":=")
		if p > 0 {
			ret[strings.TrimSpace(line[:p])] = strings.TrimSpace(line[p+2:])
		}
	}
	if ret["INCLUDE"] == "" {
		return nil, fmt.Errorf("invalid batch output, can't find INCLUDE entry")
	}

	return ret, nil
}

func DiscoverToolchains(feedback func(string)) []*toolchain.Chain {
	ret := []*toolchain.Chain{}
	msvcs, err := DiscoverInstallations(feedback)
	if err != nil {
		if feedback != nil {
			feedback(err.Error())
		}
		return nil
	}
	for _, msvc := range msvcs {
		ret = append(ret, TestArches(msvc, feedback)...)
	}
	return ret
}
