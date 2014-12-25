// This file was auto-generated by the veyron vdl tool.
// Source: types.vdl

// Package types contains the types used by the logreader interface.
package types

import (
	// The non-user imports are prefixed with "__" to prevent collisions.
	__vdl "v.io/veyron/veyron2/vdl"
	__verror "v.io/veyron/veyron2/verror"
)

// LogLine is a log entry from a log file.
type LogEntry struct {
	// The offset (in bytes) where this entry starts.
	Position int64
	// The content of the log entry.
	Line string
}

func (LogEntry) __VDLReflect(struct {
	Name string "v.io/veyron/veyron2/services/mgmt/logreader/types.LogEntry"
}) {
}

func init() {
	__vdl.Register(LogEntry{})
}

// A special NumEntries value that indicates that all entries should be
// returned by ReadLog.
const AllEntries = int32(-1)

// This error indicates that the end of the file was reached.
const EOF = __verror.ID("v.io/veyron/veyron2/services/mgmt/logreader/types.EOF")
