package java

import (
	"bytes"
	"log"

	"v.io/core/veyron2/vdl/compile"
	"v.io/core/veyron2/vdl/vdlutil"
)

const constTmpl = `// This file was auto-generated by the veyron vdl tool.
// Source(s): {{ .Source }}
package {{ .PackagePath }};


public final class {{ .ClassName }} {
    {{ range $file := .Files }}

    /* The following constants originate in file: {{ $file.Name }} */
    {{/*Constants*/}}
    {{ range $const := $file.Consts }}
    {{ $const.Doc }}
    {{ $const.AccessModifier }} static final {{ $const.Type }} {{ $const.Name }} = {{ $const.Value }};
    {{ end }} {{/* end range $file.Consts */}}
    {{/*Error Defs*/}}
    {{ range $error := $file.Errors }}
    {{ $error.Doc }}
    {{ $error.AccessModifier }} static final io.v.core.veyron2.verror2.VException.IDAction {{ $error.Name }} = io.v.core.veyron2.verror2.VException.register("{{ $error.ID }}", io.v.core.veyron2.verror2.VException.ActionCode.{{ $error.ActionName }}, "{{ $error.EnglishFmt }}");
    {{ end }} {{/* range $file.Errors */}}

    {{ end }} {{/* range .Files */}}

    static {
    	{{ range $file := .Files }}
    	/* The following errors originate in file: {{ $file.Name }} */
    	{{ range $error := $file.Errors }}
    	{{ range $format := $error.Formats}}
    	io.v.core.veyron2.i18n.Language.getDefaultCatalog().setWithBase("{{ $format.Lang }}", {{ $error.Name }}.getID(), "{{ $format.Fmt }}");
    	{{ end }} {{/* range $error.Formats */}}
    	{{ end }} {{/* range $file.Errors */}}
    	{{ end }} {{/* range .Files */}}
    }

    {{ range $file := .Files }}
    /* The following error creator methods originate in file: {{ $file.Name }} */
    {{ range $error := $file.Errors }}
    /**
     * Creates an error with {@code {{ $error.Name }}} identifier.
     */
    public static io.v.core.veyron2.verror2.VException {{ $error.MethodName }}(io.v.core.veyron2.context.VContext _ctx{{ $error.MethodArgs}}) {
    	final java.lang.Object[] _params = new java.lang.Object[] { {{ $error.Params }} };
    	final java.lang.reflect.Type[] _paramTypes = new java.lang.reflect.Type[]{ {{ $error.ParamTypes }} };
    	return io.v.core.veyron2.verror2.VException.make({{ $error.Name }}, _ctx, _paramTypes, _params);
    }
    {{ end }} {{/* range $file.Errors */}}
    {{ end }} {{/* range .Files */}}
}
`

type constConst struct {
	AccessModifier string
	Doc            string
	Type           string
	Name           string
	Value          string
}

type constError struct {
	AccessModifier string
	Doc            string
	Name           string
	ID             string
	ActionName     string
	EnglishFmt     string
	Formats        []constErrorFormat
	MethodName     string
	MethodArgs     string
	Params         string
	ParamTypes     string
}

type constErrorFormat struct {
	Lang string
	Fmt  string
}

type constFile struct {
	Name   string
	Consts []constConst
	Errors []constError
}

func shouldGenerateConstFile(pkg *compile.Package) bool {
	for _, file := range pkg.Files {
		if len(file.ConstDefs) > 0 || len(file.ErrorDefs) > 0 {
			return true
		}
	}
	return false
}

// genConstFileJava generates the (single) Java file that contains constant
// definitions from all the VDL files.
func genJavaConstFile(pkg *compile.Package, env *compile.Env) *JavaFileInfo {
	if !shouldGenerateConstFile(pkg) {
		return nil
	}

	className := "Constants"

	files := make([]constFile, len(pkg.Files))
	for i, file := range pkg.Files {
		consts := make([]constConst, len(file.ConstDefs))
		for j, cnst := range file.ConstDefs {
			consts[j].AccessModifier = accessModifierForName(cnst.Name)
			consts[j].Doc = javaDoc(cnst.Doc)
			consts[j].Type = javaType(cnst.Value.Type(), false, env)
			consts[j].Name = vdlutil.ToConstCase(cnst.Name)
			consts[j].Value = javaConstVal(cnst.Value, env)
		}
		errors := make([]constError, len(file.ErrorDefs))
		for j, err := range file.ErrorDefs {
			formats := make([]constErrorFormat, len(err.Formats))
			for k, format := range err.Formats {
				formats[k].Lang = string(format.Lang)
				formats[k].Fmt = format.Fmt
			}
			errors[j].AccessModifier = accessModifierForName(err.Name)
			errors[j].Doc = javaDoc(err.Doc)
			errors[j].Name = vdlutil.ToConstCase(err.Name)
			errors[j].ID = string(err.ID)
			errors[j].ActionName = vdlutil.ToConstCase(err.Action.String())
			errors[j].EnglishFmt = err.English
			errors[j].Formats = formats
			errors[j].MethodName = "make" + toUpperCamelCase(err.Name)
			errors[j].MethodArgs = javaDeclarationArgStr(err.Params, env, true)
			errors[j].Params = javaCallingArgStr(err.Params, false)
			errors[j].ParamTypes = javaCallingArgTypeStr(err.Params, env)
		}
		files[i].Name = file.BaseName
		files[i].Consts = consts
		files[i].Errors = errors
	}

	data := struct {
		ClassName   string
		Source      string
		PackagePath string
		Files       []constFile
	}{
		ClassName:   className,
		Source:      javaFileNames(pkg.Files),
		PackagePath: javaPath(javaGenPkgPath(pkg.Path)),
		Files:       files,
	}
	var buf bytes.Buffer
	err := parseTmpl("const", constTmpl).Execute(&buf, data)
	if err != nil {
		log.Fatalf("vdl: couldn't execute const template: %v", err)
	}
	return &JavaFileInfo{
		Name: className + ".java",
		Data: buf.Bytes(),
	}
}
