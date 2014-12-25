// This file was auto-generated by the veyron vdl tool.
// Source: types.vdl

package main

import (
	// The non-user imports are prefixed with "__" to prevent collisions.
	__fmt "fmt"
	__vdl "v.io/veyron/veyron2/vdl"
)

type dataRep int

const (
	dataRepHex dataRep = iota
	dataRepBinary
)

// dataRepAll holds all labels for dataRep.
var dataRepAll = []dataRep{dataRepHex, dataRepBinary}

// dataRepFromString creates a dataRep from a string label.
func dataRepFromString(label string) (x dataRep, err error) {
	err = x.Set(label)
	return
}

// Set assigns label to x.
func (x *dataRep) Set(label string) error {
	switch label {
	case "Hex", "hex":
		*x = dataRepHex
		return nil
	case "Binary", "binary":
		*x = dataRepBinary
		return nil
	}
	*x = -1
	return __fmt.Errorf("unknown label %q in main.dataRep", label)
}

// String returns the string label of x.
func (x dataRep) String() string {
	switch x {
	case dataRepHex:
		return "Hex"
	case dataRepBinary:
		return "Binary"
	}
	return ""
}

func (dataRep) __VDLReflect(struct {
	Name string "v.io/veyron/veyron2/vom2/vom.dataRep"
	Enum struct{ Hex, Binary string }
}) {
}

func init() {
	__vdl.Register(dataRepHex)
}
