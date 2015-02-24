package compile_test

import (
	"testing"

	"v.io/v23/vdl/build"
	"v.io/v23/vdl/compile"
	"v.io/v23/vdl/vdltest"
)

func TestIdentConflict(t *testing.T) {
	tests := []struct {
		Name string
		Data string
	}{
		// Test conflicting identifiers.
		{"Type", `type foo int64; type foo int64`},
		{"TypeMixed", `type FoO int64; type foo int64`},

		{"Const", `const foo = true; const foo = true`},
		{"ConstMixed", `const FoO = true; const foo = true`},

		{"Interface", `type foo interface{}; type foo interface{}`},
		{"InterfaceMixed", `type FoO interface{}; type foo interface{}`},

		{"Error", `error foo() {"en":"a"}; error foo() {"en":"a"}`},
		{"ErrorMixed", `error FoO() {"en":"a"}; error foo() {"en":"a"}`},

		{"TypeAndConst", `type foo int64; const foo = true`},
		{"TypeAndConstMixed", `type FoO int64; const foo = true`},
		{"TypeAndInterface", `type foo int64; type foo interface{}`},
		{"TypeAndInterfaceMixed", `type FoO int64; type foo interface{}`},
		{"TypeAndError", `type foo int64; error foo() {"en":"a"}`},
		{"TypeAndErrorMixed", `type foo int64; error FoO() {"en":"a"}`},

		{"ConstAndInterface", `const foo = true; type foo interface{}`},
		{"ConstAndInterfaceMixed", `const FoO = true; type foo interface{}`},
		{"ConstAndError", `const foo = true; error foo() {"en":"a"}`},
		{"ConstAndErrorMixed", `const foo = true; error FoO() {"en":"a"}`},

		{"InterfaceAndError", `type foo interface{}; error foo() {"en":"a"}`},
		{"InterfaceAndErrorMixed", `type foo interface{}; error FoO() {"en":"a"}`},
	}
	for _, test := range tests {
		env := compile.NewEnv(-1)
		files := map[string]string{
			test.Name + ".vdl": "package a\n" + test.Data,
		}
		buildPkg := vdltest.FakeBuildPackage(test.Name, test.Name, files)
		if pkg := build.BuildPackage(buildPkg, env); pkg != nil {
			t.Errorf("%s got package, want nil", test.Name)
		}
		vdltest.ExpectResult(t, env.Errors, test.Name, "name conflict")
	}
}
