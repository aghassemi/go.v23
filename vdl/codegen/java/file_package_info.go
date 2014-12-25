package java

import (
	"bytes"
	"log"

	"v.io/veyron/veyron2/vdl/compile"
)

const packageTmpl = `// This file was auto-generated by the veyron vdl tool.
// Source: {{ .Source }}

{{ .Doc }}
package {{ .PackagePath }};
`

// genPackageFileJava generates the Java package info file, iff any package
// comments were specified in the package's VDL files.
func genJavaPackageFile(pkg *compile.Package, env *compile.Env) *JavaFileInfo {
	generated := false
	for _, file := range pkg.Files {
		if file.PackageDef.Doc != "" {
			if generated {
				log.Printf("WARNING: Multiple vdl files with package documentation. One will be overwritten.")
				return nil
			}
			generated = true

			data := struct {
				Source      string
				PackagePath string
				Doc         string
			}{
				Source:      javaFileNames(pkg.Files),
				PackagePath: javaPath(javaGenPkgPath(pkg.Path)),
				Doc:         javaDoc(file.PackageDef.Doc),
			}
			var buf bytes.Buffer
			err := parseTmpl("package", packageTmpl).Execute(&buf, data)
			if err != nil {
				log.Fatalf("vdl: couldn't execute package template: %v", err)
			}
			return &JavaFileInfo{
				Name: "package-info.java",
				Data: buf.Bytes(),
			}
		}
	}
	return nil
}
