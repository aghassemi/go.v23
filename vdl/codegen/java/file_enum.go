package java

import (
	"bytes"
	"log"

	"v.io/veyron/veyron2/vdl/compile"
)

const enumTmpl = `
// This file was auto-generated by the veyron vdl tool.
// Source: {{.Source}}
package {{.PackagePath}};

/**
 * type {{.Name}} {{.VdlTypeString}} {{.Doc}}
 **/
@io.veyron.veyron.veyron2.vdl.GeneratedFromVdl(name = "{{.VdlTypeName}}")
{{ .AccessModifier }} final class {{.Name}} extends io.veyron.veyron.veyron2.vdl.VdlEnum {
    {{ range $index, $label := .EnumLabels }}
        @io.veyron.veyron.veyron2.vdl.GeneratedFromVdl(name = "{{$label}}", index = {{$index}})
        public static final {{$.Name}} {{$label}};
    {{ end }}

    public static final io.veyron.veyron.veyron2.vdl.VdlType VDL_TYPE =
            io.veyron.veyron.veyron2.vdl.Types.getVdlTypeFromReflect({{.Name}}.class);

    static {
        {{ range $label := .EnumLabels }}
            {{$label}} = new {{$.Name}}("{{$label}}");
        {{ end }}
    }

    private {{.Name}}(String name) {
        super(VDL_TYPE, name);
    }

    public static {{.Name}} valueOf(String name) {
        {{ range $label := .EnumLabels }}
            if ("{{$label}}".equals(name)) {
                return {{$label}};
            }
        {{ end }}
        throw new java.lang.IllegalArgumentException();
    }

    @Override
    public void writeToParcel(android.os.Parcel out, int flags) {
        out.writeString(name());
    }

    @SuppressWarnings("hiding")
    public static final android.os.Parcelable.Creator<{{.Name}}> CREATOR =
            new android.os.Parcelable.Creator<{{.Name}}>() {
        @SuppressWarnings("unchecked")
        @Override
        public {{.Name}} createFromParcel(android.os.Parcel in) {
            return {{.Name}}.valueOf(in.readString());
        }

        @Override
        public {{.Name}}[] newArray(int size) {
            return new {{.Name}}[size];
        }
    };
}
`

// genJavaEnumFile generates the Java class file for the provided user-defined enum type.
func genJavaEnumFile(tdef *compile.TypeDef, env *compile.Env) JavaFileInfo {
	labels := make([]string, tdef.Type.NumEnumLabel())
	for i := 0; i < tdef.Type.NumEnumLabel(); i++ {
		labels[i] = tdef.Type.EnumLabel(i)
	}
	javaTypeName := toUpperCamelCase(tdef.Name)
	data := struct {
		AccessModifier string
		EnumLabels     []string
		Doc            string
		Name           string
		PackagePath    string
		Source         string
		VdlTypeName    string
		VdlTypeString  string
	}{
		AccessModifier: accessModifierForName(tdef.Name),
		EnumLabels:     labels,
		Doc:            javaDocInComment(tdef.Doc),
		Name:           javaTypeName,
		PackagePath:    javaPath(javaGenPkgPath(tdef.File.Package.Path)),
		Source:         tdef.File.BaseName,
		VdlTypeName:    tdef.Type.Name(),
		VdlTypeString:  tdef.Type.String(),
	}
	var buf bytes.Buffer
	err := parseTmpl("enum", enumTmpl).Execute(&buf, data)
	if err != nil {
		log.Fatalf("vdl: couldn't execute enum template: %v", err)
	}
	return JavaFileInfo{
		Name: javaTypeName + ".java",
		Data: buf.Bytes(),
	}
}
