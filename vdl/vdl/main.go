// The following enables go generate to generate the doc.go file.
//go:generate go run $VANADIUM_ROOT/release/go/src/v.io/lib/cmdline/testdata/gendoc.go .

package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"v.io/lib/cmdline"
	"v.io/lib/textutil"
	"v.io/v23/vdl/build"
	"v.io/v23/vdl/codegen/golang"
	"v.io/v23/vdl/codegen/java"
	"v.io/v23/vdl/codegen/javascript"
	"v.io/v23/vdl/compile"
	"v.io/v23/vdl/vdlutil"
	"v.io/v23/vdlroot/vdltool"
)

func init() {
	log.SetFlags(log.Lshortfile | log.Ltime | log.Lmicroseconds)
}

func main() {
	os.Exit(cmdVDL.Main())
}

func checkErrors(errs *vdlutil.Errors) error {
	if errs.IsEmpty() {
		return nil
	}
	return fmt.Errorf(`
%s   (run with "vdl -v" for verbose logging or "vdl help" for help)`, errs)
}

// runHelper returns a function that generates a sorted list of transitive
// targets, and calls the supplied run function.
func runHelper(run func(targets []*build.Package, env *compile.Env)) func(cmd *cmdline.Command, args []string) error {
	return func(cmd *cmdline.Command, args []string) error {
		if flagVerbose {
			vdlutil.SetVerbose()
		}
		if len(args) == 0 {
			// If the user doesn't specify any targets, the cwd is implied.
			args = append(args, ".")
		}
		env := compile.NewEnv(flagMaxErrors)
		env.DisallowPathQualifiers()
		mode := build.UnknownPathIsError
		if flagIgnoreUnknown {
			mode = build.UnknownPathIsIgnored
		}
		var opts build.Opts
		opts.Extensions = strings.Split(flagExts, ",")
		opts.VDLConfigName = flagVDLConfig
		targets := build.TransitivePackages(args, mode, opts, env.Errors)
		if err := checkErrors(env.Errors); err != nil {
			return err
		}
		run(targets, env)
		return checkErrors(env.Errors)
	}
}

var topicPackages = cmdline.Topic{
	Name:  "packages",
	Short: "Description of package lists",
	Long: `
Most vdl commands apply to a list of packages:

   vdl command <packages>

<packages> are a list of packages to process, similar to the standard go tool.
In its simplest form each package is an import path; e.g.
   "v.io/core/veyron/lib/vdl"

A package that is an absolute path or that begins with a . or .. element is
interpreted as a file system path, and denotes the package in that directory.

A package is a pattern if it includes one or more "..." wildcards, each of which
can match any string, including the empty string and strings containing
slashes.  Such a pattern expands to all packages found in VDLPATH with names
matching the pattern.  As a special-case, x/... matches x as well as x's
subdirectories.

The special-case "all" is a synonym for "...", and denotes all packages found
in VDLPATH.

Import path elements and file names are not allowed to begin with "." or "_";
such paths are ignored in wildcard matches, and return errors if specified
explicitly.

 Run "vdl help vdlpath" to see docs on VDLPATH.
 Run "go help packages" to see the standard go package docs.
`,
}

var topicVdlPath = cmdline.Topic{
	Name:  "vdlpath",
	Short: "Description of VDLPATH environment variable",
	Long: `
The VDLPATH environment variable is used to resolve import statements.
It must be set to compile and generate vdl packages.

The format is a colon-separated list of directories, where each directory must
have a "src/" directory that holds vdl source code.  The path below 'src'
determines the import path.  If VDLPATH specifies multiple directories, imports
are resolved by picking the first directory with a matching import name.

An example:

   VDPATH=/home/user/vdlA:/home/user/vdlB

   /home/user/vdlA/
      src/
         foo/                 (import "foo" refers here)
            foo1.vdl
   /home/user/vdlB/
      src/
         foo/                 (this package is ignored)
            foo2.vdl
         bar/
            baz/              (import "bar/baz" refers here)
               baz.vdl
`,
}

var topicVdlRoot = cmdline.Topic{
	Name:  "vdlroot",
	Short: "Description of VDLROOT environment variable",
	Long: `
The VDLROOT environment variable is similar to VDLPATH, but instead of pointing
to multiple user source directories, it points at a single source directory
containing the standard vdl packages.

Setting VDLROOT is optional.

If VDLROOT is empty, we try to construct it out of the VANADIUM_ROOT environment
variable.  It is an error if both VDLROOT and VANADIUM_ROOT are empty.
`,
}

var topicVdlConfig = cmdline.Topic{
	Name:  "vdl.config",
	Short: "Description of vdl.config files",
	Long: `
Each vdl source package P may contain an optional file "vdl.config" within the P
directory.  This file specifies additional configuration for the vdl tool.

The format of this file is described by the vdltool.Config type in the "vdltool"
standard package, located at VDLROOT/vdltool/config.vdl.

If the file does not exist, we use the zero value of vdl.Config.
`,
}

const pkgArgName = "<packages>"
const pkgArgLong = `
<packages> are a list of packages to process, similar to the standard go tool.
For more information, run "vdl help packages".
`

var cmdCompile = &cmdline.Command{
	Run:   runHelper(runCompile),
	Name:  "compile",
	Short: "Compile packages and dependencies, but don't generate code",
	Long: `
Compile compiles packages and their transitive dependencies, but does not
generate code.  This is useful to sanity-check that your VDL files are valid.
`,
	ArgsName: pkgArgName,
	ArgsLong: pkgArgLong,
}

var cmdGenerate = &cmdline.Command{
	Run:   runHelper(runGenerate),
	Name:  "generate",
	Short: "Compile packages and dependencies, and generate code",
	Long: `
Generate compiles packages and their transitive dependencies, and generates code
in the specified languages.
`,
	ArgsName: pkgArgName,
	ArgsLong: pkgArgLong,
}

var cmdAudit = &cmdline.Command{
	Run:   runHelper(runAudit),
	Name:  "audit",
	Short: "Check if any packages are stale and need generation",
	Long: `
Audit runs the same logic as generate, but doesn't write out generated files.
Returns a 0 exit code if all packages are up-to-date, otherwise returns a
non-0 exit code indicating some packages need generation.
`,
	ArgsName: pkgArgName,
	ArgsLong: pkgArgLong,
}

var cmdList = &cmdline.Command{
	Run:   runHelper(runList),
	Name:  "list",
	Short: "List package and dependency info in transitive order",
	Long: `
List returns information about packages and their transitive dependencies, in
transitive order.  This is the same order the generate and compile commands use
for processing.  If "vdl list A" is run and A depends on B, which depends on C,
the returned order will be C, B, A.  If multiple packages are specified the
ordering is over all combined dependencies.

Reminder: cyclic dependencies between packages are not allowed.  Cyclic
dependencies between VDL files within the same package are also not allowed.
This is more strict than regular Go; it makes it easier to generate code for
other languages like C++.
`,
	ArgsName: pkgArgName,
	ArgsLong: pkgArgLong,
}

var genLangAll = genLangs(vdltool.GenLanguageAll)

type genLangs []vdltool.GenLanguage

func (gls genLangs) String() string {
	var ret string
	for i, gl := range gls {
		if i > 0 {
			ret += ","
		}
		ret += gl.String()
	}
	return ret
}

func (gls *genLangs) Set(value string) error {
	// If the flag is repeated on the cmdline it is overridden.  Duplicates within
	// the comma separated list are ignored, and retain their original ordering.
	*gls = genLangs{}
	seen := make(map[vdltool.GenLanguage]bool)
	for _, str := range strings.Split(value, ",") {
		gl, err := vdltool.GenLanguageFromString(str)
		if err != nil {
			return err
		}
		if !seen[gl] {
			seen[gl] = true
			*gls = append(*gls, gl)
		}
	}
	return nil
}

// genOutDir has three modes:
//   1) If dir is non-empty, we use it as the out dir.
//   2) If rules is non-empty, we translate using the xlate rules.
//   3) If everything is empty, we generate in-place.
type genOutDir struct {
	dir   string
	rules xlateRules
}

// xlateSrcDst specifies a translation rule, where src must match the suffix of
// the path just before the package path, and dst is the replacement for src.
// If dst is the special string "SKIP" we'll skip generation of packages
// matching the src.
type xlateSrcDst struct {
	src, dst string
}

// xlateRules specifies a collection of translation rules.
type xlateRules []xlateSrcDst

func (x *xlateRules) String() (ret string) {
	for _, srcdst := range *x {
		if len(ret) > 0 {
			ret += ","
		}
		ret += srcdst.src + "->" + srcdst.dst
	}
	return
}

func (x *xlateRules) Set(value string) error {
	for _, rule := range strings.Split(value, ",") {
		srcdst := strings.Split(rule, "->")
		if len(srcdst) != 2 {
			return fmt.Errorf("invalid out dir xlate rule %q (not src->dst format)", rule)
		}
		*x = append(*x, xlateSrcDst{srcdst[0], srcdst[1]})
	}
	return nil
}

func (x *genOutDir) String() string {
	if x.dir != "" {
		return x.dir
	}
	return x.rules.String()
}

func (x *genOutDir) Set(value string) error {
	if strings.Contains(value, "->") {
		x.dir = ""
		return x.rules.Set(value)
	}
	x.dir = value
	return nil
}

var (
	// Common flags for the tool itself, applicable to all commands.
	flagVerbose       bool
	flagMaxErrors     int
	flagExts          string
	flagVDLConfig     string
	flagIgnoreUnknown bool

	// Options for each command.
	optCompileStatus bool
	optGenStatus     bool
	optGenGoOutDir   = genOutDir{}
	optGenJavaOutDir = genOutDir{
		rules: xlateRules{
			{"go/src", "java/src/vdl/java"},
		},
	}
	optGenJavascriptOutDir = genOutDir{
		rules: xlateRules{
			{"release/go/src", "release/javascript/core/src"},
			{"roadmap/go/src", "release/javascript/core/src"},
			{"third_party/go/src", "SKIP"},
			{"tools/go/src", "SKIP"},
			// TODO(toddw): Skip vdlroot javascript generation for now.
			{"release/go/src/v.io/v23/vdlroot", "SKIP"},
		},
	}
	optGenJavaOutPkg = xlateRules{
		{"v.io", "io/v"},
	}
	optPathToJSCore string
	// TODO(bjornick): Add javascript to the default gen langs.
	optGenLangs = genLangs{vdltool.GenLanguageGo, vdltool.GenLanguageJava}
)

// Root returns the root command for the VDL tool.
var cmdVDL = &cmdline.Command{
	Name:  "vdl",
	Short: "Manage veyron VDL source code",
	Long: `
The vdl tool manages veyron VDL source code.  It's similar to the go tool used
for managing Go source code.
`,
	Children: []*cmdline.Command{cmdGenerate, cmdCompile, cmdAudit, cmdList},
	Topics:   []cmdline.Topic{topicPackages, topicVdlPath, topicVdlRoot, topicVdlConfig},
}

func init() {
	// Common flags for the tool itself, applicable to all commands.
	cmdVDL.Flags.BoolVar(&flagVerbose, "v", false, "Turn on verbose logging.")
	cmdVDL.Flags.IntVar(&flagMaxErrors, "max_errors", -1, "Stop processing after this many errors, or -1 for unlimited.")
	cmdVDL.Flags.StringVar(&flagExts, "exts", ".vdl", "Comma-separated list of valid VDL file name extensions.")
	cmdVDL.Flags.StringVar(&flagVDLConfig, "vdl.config", "vdl.config", "Basename of the optional per-package config file.")
	cmdVDL.Flags.BoolVar(&flagIgnoreUnknown, "ignore_unknown", false, "Ignore unknown packages provided on the command line.")

	// Options for compile.
	cmdCompile.Flags.BoolVar(&optCompileStatus, "status", true, "Show package names while we compile")

	// Options for generate.
	cmdGenerate.Flags.Var(&optGenLangs, "lang", "Comma-separated list of languages to generate, currently supporting "+genLangAll.String())
	cmdGenerate.Flags.BoolVar(&optGenStatus, "status", true, "Show package names as they are updated")
	// TODO(toddw): Move out_dir configuration into vdl.config, and provide a
	// generic override mechanism for vdl.config.
	cmdGenerate.Flags.Var(&optGenGoOutDir, "go_out_dir", `
Go output directory.  There are three modes:
   ""                     : Generate output in-place in the source tree
   "dir"                  : Generate output rooted at dir
   "src->dst[,s2->d2...]" : Generate output using translation rules
Assume your source tree is organized as follows:
   VDLPATH=/home/vdl
      /home/vdl/src/veyron/test_base/base1.vdl
      /home/vdl/src/veyron/test_base/base2.vdl
Here's example output under the different modes:
   --go_out_dir=""
      /home/vdl/src/veyron/test_base/base1.vdl.go
      /home/vdl/src/veyron/test_base/base2.vdl.go
   --go_out_dir="/tmp/foo"
      /tmp/foo/veyron/test_base/base1.vdl.go
      /tmp/foo/veyron/test_base/base2.vdl.go
   --go_out_dir="vdl/src->foo/bar/src"
      /home/foo/bar/src/veyron/test_base/base1.vdl.go
      /home/foo/bar/src/veyron/test_base/base2.vdl.go
When the src->dst form is used, src must match the suffix of the path just
before the package path, and dst is the replacement for src.  Use commas to
separate multiple rules; the first rule matching src is used.  The special dst
SKIP indicates matching packages are skipped.`)
	cmdGenerate.Flags.Var(&optGenJavaOutDir, "java_out_dir",
		"Same semantics as --go_out_dir but applies to java code generation.")
	cmdGenerate.Flags.Var(&optGenJavaOutPkg, "java_out_pkg", `
Java output package translation rules.  Must be of the form:
   "src->dst[,s2->d2...]"
If a VDL package has a prefix src, the prefix will be replaced with dst.  Use
commas to separate multiple rules; the first rule matching src is used, and if
there are no matching rules, the package remains unchanged.  The special dst
SKIP indicates matching packages are skipped.`)
	cmdGenerate.Flags.Var(&optGenJavascriptOutDir, "js_out_dir",
		"Same semantics as --go_out_dir but applies to js code generation.")
	cmdGenerate.Flags.StringVar(&optPathToJSCore, "js_relative_path_to_core", "",
		"If set, this is the relative path from js_out_dir to the root of the JS core")

	// Options for audit are identical to generate.
	cmdAudit.Flags = cmdGenerate.Flags
}

func runCompile(targets []*build.Package, env *compile.Env) {
	for _, target := range targets {
		pkg := build.BuildPackage(target, env)
		if pkg != nil && optCompileStatus {
			fmt.Println(pkg.Path)
		}
	}
}

func runGenerate(targets []*build.Package, env *compile.Env) {
	gen(false, targets, env)
}

func runAudit(targets []*build.Package, env *compile.Env) {
	if gen(true, targets, env) && env.Errors.IsEmpty() {
		// Some packages are stale, and there were no errors; return an arbitrary
		// non-0 exit code.  Errors are handled in runHelper, as usual.
		os.Exit(10)
	}
}

func shouldGenerate(config vdltool.Config, lang vdltool.GenLanguage) bool {
	// If config.GenLanguages is empty, all languages are allowed to be generated.
	_, ok := config.GenLanguages[lang]
	return len(config.GenLanguages) == 0 || ok
}

// gen generates the given targets with env.  If audit is true, only checks
// whether any packages are stale; otherwise files will actually be written out.
// Returns true if any packages are stale.
func gen(audit bool, targets []*build.Package, env *compile.Env) bool {
	anychanged := false
	for _, target := range targets {
		pkg := build.BuildPackage(target, env)
		if pkg == nil {
			// Stop at the first package that fails to compile.
			if env.Errors.IsEmpty() {
				env.Errors.Errorf("%s: internal error (compiled into nil package)", target.Path)
			}
			return true
		}
		// TODO(toddw): Skip code generation if the semantic contents of the
		// generated file haven't changed.
		pkgchanged := false
		for _, gl := range optGenLangs {
			switch gl {
			case vdltool.GenLanguageGo:
				if !shouldGenerate(pkg.Config, vdltool.GenLanguageGo) {
					continue
				}
				dir, err := xlateOutDir(target.Dir, target.GenPath, optGenGoOutDir, pkg.GenPath)
				if handleErrorOrSkip("--go_out_dir", err, env) {
					continue
				}
				for _, file := range pkg.Files {
					data := golang.Generate(file, env)
					if writeFile(audit, data, dir, file.BaseName+".go", env) {
						pkgchanged = true
					}
				}
			case vdltool.GenLanguageJava:
				if !shouldGenerate(pkg.Config, vdltool.GenLanguageJava) {
					continue
				}
				pkgPath, err := xlatePkgPath(pkg.GenPath, optGenJavaOutPkg)
				if handleErrorOrSkip("--java_out_pkg", err, env) {
					continue
				}
				dir, err := xlateOutDir(target.Dir, target.GenPath, optGenJavaOutDir, pkgPath)
				if handleErrorOrSkip("--java_out_dir", err, env) {
					continue
				}
				java.SetPkgPathXlator(func(pkgPath string) string {
					result, _ := xlatePkgPath(pkgPath, optGenJavaOutPkg)
					return result
				})
				for _, file := range java.Generate(pkg, env) {
					fileDir := filepath.Join(dir, file.Dir)
					if writeFile(audit, file.Data, fileDir, file.Name, env) {
						pkgchanged = true
					}
				}
			case vdltool.GenLanguageJavascript:
				if !shouldGenerate(pkg.Config, vdltool.GenLanguageJavascript) {
					continue
				}
				dir, err := xlateOutDir(target.Dir, target.GenPath, optGenJavascriptOutDir, pkg.GenPath)
				if handleErrorOrSkip("--js_out_dir", err, env) {
					continue
				}
				path := func(importPath string) string {
					prefix := filepath.Clean(target.Dir[0 : len(target.Dir)-len(target.GenPath)])
					pkgDir := filepath.Join(prefix, filepath.FromSlash(importPath))
					fullDir, err := xlateOutDir(pkgDir, importPath, optGenJavascriptOutDir, importPath)
					if err != nil {
						panic(err)
					}
					cleanPath, err := filepath.Rel(dir, fullDir)
					if err != nil {
						panic(err)
					}
					return cleanPath
				}
				data := javascript.Generate(pkg, env, path, optPathToJSCore)
				if writeFile(audit, data, dir, "index.js", env) {
					pkgchanged = true
				}
			default:
				env.Errors.Errorf("Generating code for language %v isn't supported", gl)
			}
		}
		if pkgchanged {
			anychanged = true
			if optGenStatus {
				fmt.Println(pkg.Path)
			}
		}
	}
	return anychanged
}

// writeFile writes data into the standard location for file, using the given
// suffix.  Errors are reported via env.  Returns true iff the file doesn't
// already exist with the given data.
func writeFile(audit bool, data []byte, dirName, baseName string, env *compile.Env) bool {
	dstName := filepath.Join(dirName, baseName)
	// Don't change anything if old and new are the same.
	if oldData, err := ioutil.ReadFile(dstName); err == nil && bytes.Equal(oldData, data) {
		return false
	}
	if !audit {
		// Create containing directory, if it doesn't already exist.
		if err := os.MkdirAll(dirName, os.FileMode(0777)); err != nil {
			env.Errors.Errorf("Couldn't create directory %s: %v", dirName, err)
			return true
		}
		if err := ioutil.WriteFile(dstName, data, os.FileMode(0666)); err != nil {
			env.Errors.Errorf("Couldn't write file %s: %v", dstName, err)
			return true
		}
	}
	return true
}

func handleErrorOrSkip(prefix string, err error, env *compile.Env) bool {
	if err != nil {
		if err != errSkip {
			env.Errors.Errorf("%s error: %v", prefix, err)
		}
		return true
	}
	return false
}

var errSkip = fmt.Errorf("SKIP")

func xlateOutDir(dir, path string, outdir genOutDir, outPkgPath string) (string, error) {
	path = filepath.FromSlash(path)
	outPkgPath = filepath.FromSlash(outPkgPath)
	// Strip package path from the directory.
	if !strings.HasSuffix(dir, path) {
		return "", fmt.Errorf("package dir %q doesn't end with package path %q", dir, path)
	}
	dir = filepath.Clean(dir[:len(dir)-len(path)])

	switch {
	case outdir.dir != "":
		return filepath.Join(outdir.dir, outPkgPath), nil
	case len(outdir.rules) == 0:
		return filepath.Join(dir, outPkgPath), nil
	}
	// Try translation rules in order.
	for _, xlate := range outdir.rules {
		d := dir
		if !strings.HasSuffix(d, xlate.src) {
			continue
		}
		if xlate.dst == "SKIP" {
			return "", errSkip
		}
		d = filepath.Clean(d[:len(d)-len(xlate.src)])
		return filepath.Join(d, xlate.dst, outPkgPath), nil
	}
	return "", fmt.Errorf("package prefix %q doesn't match translation rules %q", dir, outdir)
}

func xlatePkgPath(pkgPath string, rules xlateRules) (string, error) {
	for _, xlate := range rules {
		if !strings.HasPrefix(pkgPath, xlate.src) {
			continue
		}
		if xlate.dst == "SKIP" {
			return pkgPath, errSkip
		}
		return xlate.dst + pkgPath[len(xlate.src):], nil
	}
	return pkgPath, nil
}

func runList(targets []*build.Package, env *compile.Env) {
	for tx, target := range targets {
		num := fmt.Sprintf("%d", tx)
		fmt.Printf("%s %s\n", num, strings.Repeat("=", termWidth()-len(num)-1))
		fmt.Printf("Name:    %v\n", target.Name)
		fmt.Printf("Config:  %+v\n", target.Config)
		fmt.Printf("Path:    %v\n", target.Path)
		fmt.Printf("GenPath: %v\n", target.GenPath)
		fmt.Printf("Dir:     %v\n", target.Dir)
		if len(target.BaseFileNames) > 0 {
			fmt.Print("Files:\n")
			for _, file := range target.BaseFileNames {
				fmt.Printf("   %v\n", file)
			}
		}
	}
}

func termWidth() int {
	if _, width, err := textutil.TerminalSize(); err == nil && width > 0 {
		return width
	}
	return 80 // have a reasonable default
}
