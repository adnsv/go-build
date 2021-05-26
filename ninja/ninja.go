package ninja

import (
	"bytes"
	"fmt"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/adnsv/go-build/compiler/toolchain"
	"github.com/adnsv/go-build/proj"
	"github.com/tessonics/codegen"
)

func Prepare(p *proj.Project, tc *toolchain.Chain, flagset proj.Flagset) []byte {
	w := &bytes.Buffer{}

	ln := func(s string) {
		fmt.Fprintln(w, s)
	}
	lf := func(s string, v ...interface{}) {
		fmt.Fprintf(w, s, v...)
	}
	h1 := func(s string) {
		fmt.Fprintf(w, "\n# %s\n", s)
	}

	ln("# auto-generated\n")
	ln("ninja_required_version = 1.8")

	h1("paths")

	lf("builddir = %s", filepath.ToSlash(p.BuildDir))

	h1("tools")

	buildables := p.Buildables()

	reqTools := proj.RequiredTools(buildables)
	for _, t := range tools {
		if reqTools[t] {
			lf("%s\t= %s", toolName[t], tc.Tools[t])
		}
	}

	defPrefix := "-D"
	incPrefix := "-I"
	arExt := ".a"
	objExt := ".o"
	if tc.Compiler == "MSVC" {
		defPrefix = "/D"
		incPrefix = "/I"
		arExt = ".lib"
		objExt = ".obj"
	}

	/*
		// collect include paths from dependencies
		depIncludeVars := []string{}
		dependencies := map[string]string{}
		if len(p.Dependencies) > 0 {
			h1("dependencies")

			for _, dn := range p.Dependencies {
				libinfo := p.Libraries[dn]
				if libinfo == nil || libinfo.Kind != p.LibraryPackageTarget {
					continue
				}

					dc := libinfo.Context
					builddirVar := dc.Name + "_builddir"
					rootVar := dc.Name + "_root"
					lf("%s = %s", builddirVar, filepath.ToSlash(libinfo.Context.BuildDir))
					lf("%s = %s", rootVar, filepath.ToSlash(filepath.Join(libinfo.Context.Package.ActualAbsDir, libinfo.Context.Target.Root)))

					inc := []string{}
					if len(dc.Transforms) > 0 {
						if len(dc.Files.Includes) > 0 {
							for _, dir := range dc.Files.Includes {
								inc = append(inc, incPrefix+"$"+builddirVar+"/transforms/"+filepath.ToSlash(dir))
							}
						} else {
							inc = append(inc, incPrefix+"$"+builddirVar+"/transforms")
						}
					}
					if len(dc.Files.Includes) > 0 {
						for _, dir := range dc.Files.Includes {
							inc = append(inc, incPrefix+"$"+rootVar+"/"+filepath.ToSlash(dir))
						}
					} else {
						inc = append(inc, incPrefix+"$"+rootVar)
					}

					lf("%s_includes = %s", dc.Name, strings.Join(inc, " "))
					depIncludeVars = append(depIncludeVars, fmt.Sprintf("$%s_includes", dc.Name))
					dependencies[dn] = fmt.Sprintf("$%s_builddir/%s%s", dc.Name, dc.Name, arExt)

			}
		}
	*/

	h1("flags")

	writeFlags := func(name string, values []string) {
		lf("%s\t= %s", name, codegen.WrapSegments(values, " \r $\n\t  ", 8, 100))
	}

	segments := []string{}
	definitionFlags := wrapEach(flags.Defines, defPrefix, "")
	if len(definitionFlags) > 0 {
		writeFlags(p.TargetName+"_defs", definitionFlags)
		segments = append(segments, "$"+p.TargetName+"_defs")
	}

	inc := []string{}
	/*
		if len(context.Transforms) > 0 {
			// make sure -I... flags for transforms appear first
			if len(context.Files.Includes) > 0 {
				for _, dir := range context.Files.Includes {
					inc = append(inc, incPrefix+"$builddir/transforms/"+filepath.ToSlash(dir))
				}
			} else {
				inc = append(inc, incPrefix+"$builddir/transforms")
			}
		}*/
	for _, dir := range p.Includes {
		dir = "/" + filepath.ToSlash(dir)
		if dir == "/." {
			dir = ""
		}
		inc = append(inc, incPrefix+"$root"+dir)
	}

	inc = append(inc, depIncludeVars...)

	if len(inc) > 0 {
		writeFlags(p.TargetName+"_includes", inc)
		segments = append(segments, "$"+p.TargetName+"_includes")
	}

	cxxFlags := flags.ToolFlags[toolchain.CXXCompiler]
	ccFlags := flags.ToolFlags[toolchain.CCompiler]

	if p.RequiredTools[toolchain.CXXCompiler] && p.RequiredTools[toolchain.CCompiler] {
		common, cxx, cc := splitCommonFlags(cxxFlags, ccFlags)
		if len(common) > 1 {
			cxxFlags = cxx
			ccFlags = cc
			writeFlags(p.TargetName+"_common", common)
			segments = append(segments, "$"+p.TargetName+"_common")
		}
	}

	if p.RequiredTools[toolchain.CXXCompiler] {
		writeFlags(p.TargetName+"_cxxflags", append(segments, cxxFlags...))
	}
	if p.RequiredTools[toolchain.CCompiler] {
		writeFlags(p.TargetName+"_ccflags", append(segments, ccFlags...))
	}
	if p.RequiredTools[toolchain.ASMCompiler] {
		writeFlags(p.TargetName+"_asmflags", flags.ToolFlags[toolchain.ASMCompiler])
	}
	if p.RequiredTools[toolchain.ResourceCompiler] {
		writeFlags(p.TargetName+"_rcflags", flags.ToolFlags[toolchain.ResourceCompiler])
	}
	if p.RequiredTools[toolchain.Archiver] {
		writeFlags(p.TargetName+"_arflags", flags.ToolFlags[toolchain.Archiver])
	}

	dllFlags := flags.ToolFlags[toolchain.DLLLinker]
	exeFlags := flags.ToolFlags[toolchain.EXELinker]
	segments = []string{}
	if p.RequiredTools[toolchain.DLLLinker] && p.RequiredTools[toolchain.EXELinker] {
		common, dll, exe := splitCommonFlags(dllFlags, exeFlags)
		if len(common) > 1 {
			dllFlags = dll
			exeFlags = exe
			writeFlags(p.TargetName+"_ldflags", common)
			segments = append(segments, "$"+p.TargetName+"_ldflags")
		}
	}

	if p.RequiredTools[toolchain.DLLLinker] {
		writeFlags(p.TargetName+"_dllflags", append(segments, dllFlags...))
	}
	if p.RequiredTools[toolchain.EXELinker] {
		writeFlags(p.TargetName+"_exeflags", append(segments, exeFlags...))
	}

	pdbName := ""

	concatPrefix := ""
	if runtime.GOOS == "windows" {
		concatPrefix = "cmd /c "
	}

	delFile := func(fn string) string {
		if strings.IndexByte(fn, ' ') >= 0 {
			fn = `"` + fn + `"`
		}
		if runtime.GOOS == "windows" {
			return fmt.Sprintf("del /q %s", fn)
		}
		return fmt.Sprintf("rm -f %s", fn)
	}

	h1("rules")

	if tc.Compiler == "MSVC" {

		pdbName = context.Target.Name + ".pdb"
		lf("\npdb = %s", pdbName)
		lf("msvc_deps_prefix = Note: including file:")

		if p.RequiredTools[toolchain.CXXCompiler] {
			s := "\nrule " + ninjaRule[toolchain.CXXCompiler]
			s += "\n    command = $" + ninjaTool[toolchain.CXXCompiler] + " /nologo $" + context.Name + "_cxxflags -c $in /Fo$out /Fd$builddir/$pdb /FS /showIncludes"
			s += "\n    description = C++ $out"
			s += "\n    deps = msvc"
			ln(s)
		}

		if context.requiredTools[toolchain.CCompiler] {
			s := "\nrule " + ninjaRule[toolchain.CCompiler]
			s += "\n    command = $" + ninjaTool[toolchain.CCompiler] + " /nologo $" + context.Name + "_ccflags -c $in /Fo$out /Fd$builddir/$pdb /FS /showIncludes"
			s += "\n    description = C $out"
			s += "\n    deps = msvc"
			ln(s)
		}

		/*	if wantASM {
			s := "\nrule asm"
			s += "\n    command = $asm /nologo $asmflags -c $in /Fo$out /Fd$builddir/$pdb /FS /showIncludes"
			s += "\n    description = C $out"
			s += "\n    deps = msvc"
			ln(s)
		}*/

		if context.requiredTools[toolchain.Archiver] {
			s := "\nrule " + ninjaRule[toolchain.Archiver]
			s += "\n    command = $" + ninjaTool[toolchain.Archiver] + " /nologo $" + context.Name + "_arflags /out:$out $in"
			s += "\n    description = AR $out`"
			ln(s)
		}

		if context.requiredTools[toolchain.DLLLinker] {
			s := "\nrule " + ninjaRule[toolchain.DLLLinker]
			s += "\n    command = $" + ninjaTool[toolchain.DLLLinker] + " /nologo $" + context.Name + "_dllflags /out:$out $in"
			ln(s)
		}

		if context.requiredTools[toolchain.EXELinker] {
			s := "\nrule " + ninjaRule[toolchain.EXELinker]
			s += "\n    command = $" + ninjaTool[toolchain.EXELinker] + " /nologo $" + context.Name + "_exeflags /implib:" + context.Target.Name + ".lib /out:$out $in"
			s += "\n    description = LINK EXECUTABLE $out"
			ln(s)
		}
	} else {
		if context.requiredTools[toolchain.CXXCompiler] {
			s := "\nrule " + ninjaRule[toolchain.CXXCompiler]
			s += "\n    command = $" + ninjaTool[toolchain.CXXCompiler] + " -MMD -MT $out -MF $out.d $" + context.Name + "_cxxflags -c $in -o $out"
			s += "\n    description = C++ $out"
			s += "\n    depfile = $out.d"
			s += "\n    deps = gcc"
			ln(s)
		}

		if context.requiredTools[toolchain.CCompiler] {
			s := "\nrule " + ninjaRule[toolchain.CCompiler]
			s += "\n    command = $" + ninjaTool[toolchain.CCompiler] + " -MMD -MT $out -MF $out.d $" + context.Name + "_ccflags -c $in -o $out"
			s += "\n    description = C $out"
			s += "\n    depfile = $out.d"
			s += "\n    deps = gcc"
			ln(s)
		}

		/*	if wantASM {
			s := "\nrule asm"
			s += "\n    command = $asm /nologo $asmflags -c $in /Fo$out /Fd$builddir/$pdb /FS /showIncludes"
			s += "\n    description = C $out"
			s += "\n    deps = msvc"
			ln(s)
		}*/

		if context.requiredTools[toolchain.Archiver] {
			s := "\nrule " + ninjaRule[toolchain.Archiver]
			s += "\n    command = " + concatPrefix + delFile("$out") + " && $" + ninjaTool[toolchain.Archiver] + " $" + context.Name + "_arflags crs $out $in"
			s += "\n    description = AR $out"
			ln(s)
		}
	}

	linkFiles := []string{}
	if context.Target.Type == "executable" {

		for _, d := range dependencies {
			linkFiles = append(linkFiles, d)
		}
	}

	h1("compile")

	for _, src := range p.Sources {
		src = filepath.Clean(src)

		rule := ""
		switch KnownExtensions[filepath.Ext(src)] {
		case ExtensionCompilationUnitCXX:
			rule = ninjaRule[toolchain.CXXCompiler]
		case ExtensionCompilationUnitCC:
			rule = ninjaRule[toolchain.CCompiler]
		case ExtensionCompilationUnitASM:
			rule = ninjaRule[toolchain.ASMCompiler]
		default:
			continue
		}

		obj := "$builddir/" + filepath.Base(src) + objExt

		w.Lf("build %s:\t%s $root/%s", obj, rule, src)
		linkFiles = append(linkFiles, obj)
	}

	if context.Target.Type == "library" {
		arName := context.Target.Name + arExt
		if len(linkFiles) > 0 {
			h1("archive")
			lf("build %s: "+ninjaRule[toolchain.Archiver]+" $\n\t%s", arName, strings.Join(linkFiles, " $\n\t"))
		}
	} else if context.Target.Type == "executable" {
		h1("link executable")
		exeName := context.Target.Name + ".exe"
		lf("build %s: "+ninjaRule[toolchain.EXELinker]+" $\n\t%s", exeName, strings.Join(linkFiles, " $\n\t"))
	}

	return w.Bytes()
}

var tools = []toolchain.Tool{
	toolchain.CXXCompiler,
	toolchain.CCompiler,
	toolchain.ASMCompiler,
	toolchain.ResourceCompiler,
	toolchain.Archiver,
	toolchain.DLLLinker,
	toolchain.EXELinker,
}

var toolName = map[toolchain.Tool]string{
	toolchain.CXXCompiler:      "cxx_compiler",
	toolchain.CCompiler:        "c_compiler",
	toolchain.ASMCompiler:      "asm_compiler",
	toolchain.ResourceCompiler: "res_compiler",
	toolchain.Archiver:         "archiver",
	toolchain.DLLLinker:        "dll_linker",
	toolchain.EXELinker:        "exe_linker",
}

var ruleSuffix = map[toolchain.Tool]string{
	toolchain.CXXCompiler:      "_compile_cxx",
	toolchain.CCompiler:        "_compile_c",
	toolchain.ASMCompiler:      "_compile_asm",
	toolchain.ResourceCompiler: "_compile_res",
	toolchain.Archiver:         "_archive",
	toolchain.DLLLinker:        "_link_dll",
	toolchain.EXELinker:        "_link_exe",
}
