package java

import (
	"bytes"
	"log"

	"v.io/veyron/veyron2/vdl/compile"
)

const setTmpl = `// This file was auto-generated by the veyron vdl tool.
// Source: {{.SourceFile}}

package {{.Package}};

/**
 * {{.Name}} {{.VdlTypeString}} {{.Doc}}
 **/
@io.veyron.veyron.veyron2.vdl.GeneratedFromVdl(name = "{{.VdlTypeName}}")
{{ .AccessModifier }} final class {{.Name}} extends io.veyron.veyron.veyron2.vdl.VdlSet<{{.KeyType}}> {
    public static final io.veyron.veyron.veyron2.vdl.VdlType VDL_TYPE =
            io.veyron.veyron.veyron2.vdl.Types.getVdlTypeFromReflect({{.Name}}.class);

    public {{.Name}}(java.util.Set<{{.KeyType}}> impl) {
        super(VDL_TYPE, impl);
    }

    @Override
    public void writeToParcel(android.os.Parcel out, int flags) {
        java.lang.reflect.Type keyType =
                new com.google.common.reflect.TypeToken<{{.KeyType}}>(){}.getType();
        io.veyron.veyron.veyron2.vdl.ParcelUtil.writeSet(out, this, keyType);
    }

    @SuppressWarnings("hiding")
    public static final android.os.Parcelable.Creator<{{.Name}}> CREATOR =
            new android.os.Parcelable.Creator<{{.Name}}>() {
        @SuppressWarnings("unchecked")
        @Override
        public {{.Name}} createFromParcel(android.os.Parcel in) {
            java.lang.reflect.Type keyType =
                    new com.google.common.reflect.TypeToken<{{.KeyType}}>(){}.getType();
            java.util.Set<?> set = io.veyron.veyron.veyron2.vdl.ParcelUtil.readSet(
                    in, {{.Name}}.class.getClassLoader(), keyType);
            return new {{.Name}}((java.util.Set<{{.KeyType}}>) set);
        }

        @Override
        public {{.Name}}[] newArray(int size) {
            return new {{.Name}}[size];
        }
    };
}
`

// genJavaSetFile generates the Java class file for the provided named set type.
func genJavaSetFile(tdef *compile.TypeDef, env *compile.Env) JavaFileInfo {
	javaTypeName := toUpperCamelCase(tdef.Name)
	data := struct {
		AccessModifier string
		Doc            string
		KeyType        string
		Name           string
		Package        string
		SourceFile     string
		VdlTypeName    string
		VdlTypeString  string
	}{
		AccessModifier: accessModifierForName(tdef.Name),
		Doc:            javaDocInComment(tdef.Doc),
		KeyType:        javaType(tdef.Type.Key(), true, env),
		Name:           javaTypeName,
		Package:        javaPath(javaGenPkgPath(tdef.File.Package.Path)),
		SourceFile:     tdef.File.BaseName,
		VdlTypeName:    tdef.Type.Name(),
		VdlTypeString:  tdef.Type.String(),
	}
	var buf bytes.Buffer
	err := parseTmpl("set", setTmpl).Execute(&buf, data)
	if err != nil {
		log.Fatalf("vdl: couldn't execute set template: %v", err)
	}
	return JavaFileInfo{
		Name: javaTypeName + ".java",
		Data: buf.Bytes(),
	}
}
