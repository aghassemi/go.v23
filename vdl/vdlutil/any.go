package vdlutil

import (
	"encoding/gob"

	"v.io/veyron/veyron2/vdl"
	"v.io/veyron/veyron2/vom"
)

// TODO(toddw): Move the contents of this file to the vdl package after the vom2
// transition.  We can't just move it now since vom has too many bad
// dependencies that we don't want to pull in to the vdl package.

// Any represents a value of the Any type in generated Go code.  We define a
// special type rather than just using interface{} in generated code, to make it
// easy to identify and add special-casing later.
//
// TODO(toddw): Rename to AnyRep
type Any interface{}

func init() {
	// TODO(toddw): Remove this call after the vom2 transition.
	vom.Register((*Any)(nil))
}

// Register is a convenience that registers the value with gob, vom and vdl.
// TODO(toddw): Remove after the vom2 transition, and change calls to
// vdl.Register.
func Register(value interface{}) {
	gob.Register(value)
	vom.Register(value)
	vdl.Register(value)
}
