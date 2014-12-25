// This file was auto-generated by the veyron vdl tool.
// Source: config.vdl

// Package vdltool describes types used by the vdl tool itself.
package vdltool

import (
	// The non-user imports are prefixed with "__" to prevent collisions.
	__fmt "fmt"
	__vdl "v.io/veyron/veyron2/vdl"
)

// Config specifies the configuration for the vdl tool.  This is typically
// represented in optional "vdl.config" files in each vdl source package.  Each
// vdl.config file implicitly imports this package.  E.g. you may refer to
// vdltool.Config in the "vdl.config" file without explicitly importing vdltool.
type Config struct {
	// GenLanguages restricts the set of code generation languages.  If the set is
	// empty, all supported languages are allowed to be generated.
	GenLanguages map[GenLanguage]struct{}
	// Language-specific configurations.
	Go         GoConfig
	Java       JavaConfig
	Javascript JavascriptConfig
}

func (Config) __VDLReflect(struct {
	Name string "vdltool.Config"
}) {
}

// GenLanguage enumerates the known code generation languages.
type GenLanguage int

const (
	GenLanguageGo GenLanguage = iota
	GenLanguageJava
	GenLanguageJavascript
)

// GenLanguageAll holds all labels for GenLanguage.
var GenLanguageAll = []GenLanguage{GenLanguageGo, GenLanguageJava, GenLanguageJavascript}

// GenLanguageFromString creates a GenLanguage from a string label.
func GenLanguageFromString(label string) (x GenLanguage, err error) {
	err = x.Set(label)
	return
}

// Set assigns label to x.
func (x *GenLanguage) Set(label string) error {
	switch label {
	case "Go", "go":
		*x = GenLanguageGo
		return nil
	case "Java", "java":
		*x = GenLanguageJava
		return nil
	case "Javascript", "javascript":
		*x = GenLanguageJavascript
		return nil
	}
	*x = -1
	return __fmt.Errorf("unknown label %q in vdltool.GenLanguage", label)
}

// String returns the string label of x.
func (x GenLanguage) String() string {
	switch x {
	case GenLanguageGo:
		return "Go"
	case GenLanguageJava:
		return "Java"
	case GenLanguageJavascript:
		return "Javascript"
	}
	return ""
}

func (GenLanguage) __VDLReflect(struct {
	Name string "vdltool.GenLanguage"
	Enum struct{ Go, Java, Javascript string }
}) {
}

// GoConfig specifies go specific configuration.
type GoConfig struct {
	NativeTypes []GoNativeType
}

func (GoConfig) __VDLReflect(struct {
	Name string "vdltool.GoConfig"
}) {
}

// GoNativeType describes the mapping from a VDL wire type to a Go native type.
// This is typically used when the Go native type is an idiomatic type
// convenient for the user, but you still need a standard VDL representation for
// wire compatibility.  E.g. the VDL time package is necessary for wire
// compatibility across languages, while generated Go code uses the standard Go
// time package.
//
// The code generator assumes the existence of a pair of conversion functions
// converting between the wire and native types:
//   type WireType ...
//   func (x WireType) VDLToNative(n *Native) error
//   func (x *WireType) VDLFromNative(n Native) error
//
// TODO(toddw): VDL compiler support for native types isn't implemented yet.
type GoNativeType struct {
	// WireType is the name of the vdl type that describes the format on the wire.
	// The wire type must be defined in the vdl package associated with the
	// vdl.config file; i.e. in the same directory.
	WireType string
	// NativeType is the name of the Go native type.  Include the package
	// qualifier if the Go native type isn't defined in the same package as the
	// wire type, and add import(s) for the necessary Go packages.
	NativeType string
	// Imports lists the Go imports required by NativeType.
	Imports []string
}

func (GoNativeType) __VDLReflect(struct {
	Name string "vdltool.GoNativeType"
}) {
}

// JavaConfig specifies java specific configuration.
type JavaConfig struct {
}

func (JavaConfig) __VDLReflect(struct {
	Name string "vdltool.JavaConfig"
}) {
}

// JavascriptConfig specifies javascript specific configuration.
type JavascriptConfig struct {
}

func (JavascriptConfig) __VDLReflect(struct {
	Name string "vdltool.JavascriptConfig"
}) {
}

func init() {
	__vdl.Register(Config{})
	__vdl.Register(GenLanguageGo)
	__vdl.Register(GoConfig{})
	__vdl.Register(GoNativeType{})
	__vdl.Register(JavaConfig{})
	__vdl.Register(JavascriptConfig{})
}
