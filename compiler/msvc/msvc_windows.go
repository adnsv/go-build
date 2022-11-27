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
	"github.com/adnsv/go-build/compiler/triplet"
	"github.com/adnsv/go-utils/filesystem"
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

	if pf := os.Getenv("ProgramFiles(x86)"); filesystem.DirExists(pf) {
		addPath(filepath.Join(pf, vswhereSubpath))
	}
	if pf := os.Getenv("ProgramFiles"); filesystem.DirExists(pf) {
		addPath(filepath.Join(pf, vswhereSubpath))
	}
	if filesystem.DirExists("C:\\Program Files (x86)") {
		addPath(filepath.Join("C:\\Program Files (x86)", vswhereSubpath))
	}
	addPath("vswhere.exe")

	vswherePath := ""
	for _, path := range paths {
		if filesystem.FileExists(path) {
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
	if feedback != nil {
		feedback(fmt.Sprintf("vswhere output: %s", string(buf)))
	}

	type vsWhereInstallation struct {
		InstanceID          string `json:"instanceId"`
		DisplayName         string `json:"displayName"`
		InstallationPath    string `json:"installationPath"`
		InstallationVersion string `json:"installationVersion"`
		Description         string `json:"description"`
		IsPrerelease        bool   `json:"isPrerelease"`
	}

	tmp := []*vsWhereInstallation{}
	err = json.Unmarshal(buf, &tmp)
	if err != nil {
		return nil, fmt.Errorf("failed to parse vswhere output: %s", err)
	}

	installations := make([]*Installation, 0, len(tmp))
	for _, i := range tmp {
		inst := &Installation{
			InstanceID:          i.InstanceID,
			DisplayName:         i.DisplayName,
			InstallationPath:    filepath.ToSlash(i.InstallationPath),
			InstallationVersion: i.InstallationVersion,
			Description:         i.Description,
			IsPrerelease:        i.IsPrerelease,
		}

		toolsetver_fn := filepath.Join(i.InstallationPath, "VC", "Auxiliary", "Build", "Microsoft.VCToolsVersion.default.txt")
		if filesystem.FileExists(toolsetver_fn) {
			if buf, err := os.ReadFile(toolsetver_fn); err == nil {
				inst.ToolsetVersion = strings.TrimSpace(string(buf))
			}
		}
		installations = append(installations, inst)
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

	if !filesystem.FileExists(devbat) {
		devbat = filepath.Join(inst.InstallationPath, "VC", "vcvarsall.bat")
		if !filesystem.FileExists(devbat) {
			devbat = ""
			if feedback != nil {
				feedback("ERROR: failed to locate vcvarsall.bat file")
			}
			return nil
		}
	}
	if feedback != nil {
		feedback(fmt.Sprintf("using devbat: %s", devbat))
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
			v, err := CollectBatVars(devbat, arch, strconv.Itoa(majorVer), commonDir, feedback)
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

		tt := triplet.Full{
			Target: triplet.Target{
				Arch: triplet.NormalizeArch(arch),
				OS:   "windows",
				ABI:  "pe",
				LibC: "msvcrt",
			}}

		tc := &toolchain.Chain{
			Compiler:            "MSVC",
			FullVersion:         fmt.Sprintf("%s - %s - %s", inst.DisplayName, tt.Arch, inst.InstallationVersion),
			Version:             inst.InstallationVersion,
			Target:              tt,
			InstalledDir:        filepath.ToSlash(inst.InstallationPath),
			VisualStudioID:      inst.InstanceID,
			VisualStudioArch:    arch,
			VisualStudioVersion: vars["VISUALSTUDIOVERSION"],
			UCRTVersion:         vars["UCRTVERSION"],
			ToolsetVersion:      inst.ToolsetVersion,
			Tools:               map[toolchain.Tool]string{},
		}

		if s := vars["WINDOWSSDKVERSION"]; s != "" {
			tc.WindowsSDKVersion = strings.TrimRight(s, `\`)
		}

		if incs := filepath.SplitList(vars["INCLUDE"]); len(incs) > 0 {
			for _, v := range incs {
				if v != "" && filesystem.DirExists(v) {
					tc.CCIncludeDirs = append(tc.CCIncludeDirs, filepath.ToSlash(v))
				}
			}
		}
		tc.CXXIncludeDirs = tc.CCIncludeDirs

		if libs := strings.Split(vars["LIB"], ";"); len(libs) > 0 {
			for _, v := range libs {
				if v != "" {
					tc.LibraryDirs = append(tc.LibraryDirs, filepath.ToSlash(v))
				}
			}
		}

		paths := filepath.SplitList(vars["PATH"])

		if s := vars["CL"]; s != "" {
			tc.Tools[toolchain.CCompiler] = filepath.ToSlash(s)
			tc.Tools[toolchain.CXXCompiler] = filepath.ToSlash(s)
		} else {
			for _, path := range paths {
				fn := filepath.Join(path, "cl.exe")
				if filesystem.FileExists(fn) {
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
				if filesystem.FileExists(fn) {
					tc.Tools[toolchain.DLLLinker] = filepath.ToSlash(fn)
					tc.Tools[toolchain.EXELinker] = filepath.ToSlash(fn)
					break
				}
			}
		}

		ar := filepath.Join(filepath.Dir(tc.Tools[toolchain.DLLLinker]), "lib.exe")
		if !filesystem.FileExists(ar) {
			for _, path := range paths {
				fn := filepath.Join(path, "lib.exe")
				if filesystem.FileExists(fn) {
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
			if filesystem.FileExists(fn) {
				tc.Tools[toolchain.ResourceCompiler] = filepath.ToSlash(fn)
				break
			}
		}

		for _, path := range paths {
			fn := filepath.Join(path, "mt.exe")
			if filesystem.FileExists(fn) {
				tc.Tools[toolchain.ManifestTool] = filepath.ToSlash(fn)
				break
			}
		}

		for k, v := range vars {
			tc.Environment = append(tc.Environment, fmt.Sprintf("%s=%s", k, v))
		}
		sort.Strings(tc.Environment)

		toolchains = append(toolchains, tc)
	}

	return toolchains
}

func CollectBatVars(devbat string, arg string, majorVer string, commonDir string, feedback func(string)) (map[string]string, error) {
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

	if feedback != nil {
		feedback(fmt.Sprintf("- calling devbat %s", arg))
	}
	lines := strings.Split(string(buf), "\n")
	for _, line := range lines {
		if feedback != nil {
			feedback(fmt.Sprintf("> %s", line))
		}
		p := strings.Index(line, ":=")
		if p > 0 {
			n := strings.TrimSpace(line[:p])
			v := strings.TrimSpace(line[p+2:])
			ret[n] = v
			//if feedback != nil {
			//	feedback(fmt.Sprintf("    - %s=%s", n, v))
			//}
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
