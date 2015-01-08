// This file was auto-generated by the veyron vdl tool.
// Source: stats.vdl

// Package stats defines an interface to access statistical information for
// troubleshooting and monitoring purposes.
package stats

import (
	"v.io/core/veyron2/services/security/access"

	"v.io/core/veyron2/services/watch"

	// The non-user imports are prefixed with "__" to prevent collisions.
	__veyron2 "v.io/core/veyron2"
	__context "v.io/core/veyron2/context"
	__ipc "v.io/core/veyron2/ipc"
	__vdlutil "v.io/core/veyron2/vdl/vdlutil"
	__wiretype "v.io/core/veyron2/wiretype"
)

// TODO(toddw): Remove this line once the new signature support is done.
// It corrects a bug where __wiretype is unused in VDL pacakges where only
// bootstrap types are used on interfaces.
const _ = __wiretype.TypeIDInvalid

// StatsClientMethods is the client interface
// containing Stats methods.
//
// The Stats interface is used to access stats for troubleshooting and
// monitoring purposes. The stats objects are discoverable via the Globbable
// interface and watchable via the GlobWatcher interface.
//
// The types of the object values are implementation specific, but should be
// primarily numeric in nature, e.g. counters, memory usage, latency metrics,
// etc.
type StatsClientMethods interface {
	// GlobWatcher allows a client to receive updates for changes to objects
	// that match a pattern.  See the package comments for details.
	watch.GlobWatcherClientMethods
	// Value returns the current value of an object, or an error. The type
	// of the value is implementation specific.
	// Some objects may not have a value, in which case, Value() returns
	// a NoValue error.
	Value(*__context.T, ...__ipc.CallOpt) (__vdlutil.Any, error)
}

// StatsClientStub adds universal methods to StatsClientMethods.
type StatsClientStub interface {
	StatsClientMethods
	__ipc.UniversalServiceMethods
}

// StatsClient returns a client stub for Stats.
func StatsClient(name string, opts ...__ipc.BindOpt) StatsClientStub {
	var client __ipc.Client
	for _, opt := range opts {
		if clientOpt, ok := opt.(__ipc.Client); ok {
			client = clientOpt
		}
	}
	return implStatsClientStub{name, client, watch.GlobWatcherClient(name, client)}
}

type implStatsClientStub struct {
	name   string
	client __ipc.Client

	watch.GlobWatcherClientStub
}

func (c implStatsClientStub) c(ctx *__context.T) __ipc.Client {
	if c.client != nil {
		return c.client
	}
	return __veyron2.GetClient(ctx)
}

func (c implStatsClientStub) Value(ctx *__context.T, opts ...__ipc.CallOpt) (o0 __vdlutil.Any, err error) {
	var call __ipc.Call
	if call, err = c.c(ctx).StartCall(ctx, c.name, "Value", nil, opts...); err != nil {
		return
	}
	if ierr := call.Finish(&o0, &err); ierr != nil {
		err = ierr
	}
	return
}

func (c implStatsClientStub) Signature(ctx *__context.T, opts ...__ipc.CallOpt) (o0 __ipc.ServiceSignature, err error) {
	var call __ipc.Call
	if call, err = c.c(ctx).StartCall(ctx, c.name, "Signature", nil, opts...); err != nil {
		return
	}
	if ierr := call.Finish(&o0, &err); ierr != nil {
		err = ierr
	}
	return
}

// StatsServerMethods is the interface a server writer
// implements for Stats.
//
// The Stats interface is used to access stats for troubleshooting and
// monitoring purposes. The stats objects are discoverable via the Globbable
// interface and watchable via the GlobWatcher interface.
//
// The types of the object values are implementation specific, but should be
// primarily numeric in nature, e.g. counters, memory usage, latency metrics,
// etc.
type StatsServerMethods interface {
	// GlobWatcher allows a client to receive updates for changes to objects
	// that match a pattern.  See the package comments for details.
	watch.GlobWatcherServerMethods
	// Value returns the current value of an object, or an error. The type
	// of the value is implementation specific.
	// Some objects may not have a value, in which case, Value() returns
	// a NoValue error.
	Value(__ipc.ServerContext) (__vdlutil.Any, error)
}

// StatsServerStubMethods is the server interface containing
// Stats methods, as expected by ipc.Server.
// The only difference between this interface and StatsServerMethods
// is the streaming methods.
type StatsServerStubMethods interface {
	// GlobWatcher allows a client to receive updates for changes to objects
	// that match a pattern.  See the package comments for details.
	watch.GlobWatcherServerStubMethods
	// Value returns the current value of an object, or an error. The type
	// of the value is implementation specific.
	// Some objects may not have a value, in which case, Value() returns
	// a NoValue error.
	Value(__ipc.ServerContext) (__vdlutil.Any, error)
}

// StatsServerStub adds universal methods to StatsServerStubMethods.
type StatsServerStub interface {
	StatsServerStubMethods
	// Describe the Stats interfaces.
	Describe__() []__ipc.InterfaceDesc
	// Signature will be replaced with Describe__.
	Signature(ctx __ipc.ServerContext) (__ipc.ServiceSignature, error)
}

// StatsServer returns a server stub for Stats.
// It converts an implementation of StatsServerMethods into
// an object that may be used by ipc.Server.
func StatsServer(impl StatsServerMethods) StatsServerStub {
	stub := implStatsServerStub{
		impl: impl,
		GlobWatcherServerStub: watch.GlobWatcherServer(impl),
	}
	// Initialize GlobState; always check the stub itself first, to handle the
	// case where the user has the Glob method defined in their VDL source.
	if gs := __ipc.NewGlobState(stub); gs != nil {
		stub.gs = gs
	} else if gs := __ipc.NewGlobState(impl); gs != nil {
		stub.gs = gs
	}
	return stub
}

type implStatsServerStub struct {
	impl StatsServerMethods
	watch.GlobWatcherServerStub
	gs *__ipc.GlobState
}

func (s implStatsServerStub) Value(ctx __ipc.ServerContext) (__vdlutil.Any, error) {
	return s.impl.Value(ctx)
}

func (s implStatsServerStub) Globber() *__ipc.GlobState {
	return s.gs
}

func (s implStatsServerStub) Describe__() []__ipc.InterfaceDesc {
	return []__ipc.InterfaceDesc{StatsDesc, watch.GlobWatcherDesc}
}

// StatsDesc describes the Stats interface.
var StatsDesc __ipc.InterfaceDesc = descStats

// descStats hides the desc to keep godoc clean.
var descStats = __ipc.InterfaceDesc{
	Name:    "Stats",
	PkgPath: "v.io/core/veyron2/services/mgmt/stats",
	Doc:     "// The Stats interface is used to access stats for troubleshooting and\n// monitoring purposes. The stats objects are discoverable via the Globbable\n// interface and watchable via the GlobWatcher interface.\n//\n// The types of the object values are implementation specific, but should be\n// primarily numeric in nature, e.g. counters, memory usage, latency metrics,\n// etc.",
	Embeds: []__ipc.EmbedDesc{
		{"GlobWatcher", "v.io/core/veyron2/services/watch", "// GlobWatcher allows a client to receive updates for changes to objects\n// that match a pattern.  See the package comments for details."},
	},
	Methods: []__ipc.MethodDesc{
		{
			Name: "Value",
			Doc:  "// Value returns the current value of an object, or an error. The type\n// of the value is implementation specific.\n// Some objects may not have a value, in which case, Value() returns\n// a NoValue error.",
			OutArgs: []__ipc.ArgDesc{
				{"", ``}, // __vdlutil.Any
				{"", ``}, // error
			},
			Tags: []__vdlutil.Any{access.Tag("Debug")},
		},
	},
}

func (s implStatsServerStub) Signature(ctx __ipc.ServerContext) (__ipc.ServiceSignature, error) {
	// TODO(toddw): Replace with new Describe__ implementation.
	result := __ipc.ServiceSignature{Methods: make(map[string]__ipc.MethodSignature)}
	result.Methods["Value"] = __ipc.MethodSignature{
		InArgs: []__ipc.MethodArgument{},
		OutArgs: []__ipc.MethodArgument{
			{Name: "", Type: 65},
			{Name: "", Type: 66},
		},
	}

	result.TypeDefs = []__vdlutil.Any{
		__wiretype.NamedPrimitiveType{Type: 0x1, Name: "anydata", Tags: []string(nil)}, __wiretype.NamedPrimitiveType{Type: 0x1, Name: "error", Tags: []string(nil)}}
	var ss __ipc.ServiceSignature
	var firstAdded int
	ss, _ = s.GlobWatcherServerStub.Signature(ctx)
	firstAdded = len(result.TypeDefs)
	for k, v := range ss.Methods {
		for i, _ := range v.InArgs {
			if v.InArgs[i].Type >= __wiretype.TypeIDFirst {
				v.InArgs[i].Type += __wiretype.TypeID(firstAdded)
			}
		}
		for i, _ := range v.OutArgs {
			if v.OutArgs[i].Type >= __wiretype.TypeIDFirst {
				v.OutArgs[i].Type += __wiretype.TypeID(firstAdded)
			}
		}
		if v.InStream >= __wiretype.TypeIDFirst {
			v.InStream += __wiretype.TypeID(firstAdded)
		}
		if v.OutStream >= __wiretype.TypeIDFirst {
			v.OutStream += __wiretype.TypeID(firstAdded)
		}
		result.Methods[k] = v
	}
	//TODO(bprosnitz) combine type definitions from embeded interfaces in a way that doesn't cause duplication.
	for _, d := range ss.TypeDefs {
		switch wt := d.(type) {
		case __wiretype.SliceType:
			if wt.Elem >= __wiretype.TypeIDFirst {
				wt.Elem += __wiretype.TypeID(firstAdded)
			}
			d = wt
		case __wiretype.ArrayType:
			if wt.Elem >= __wiretype.TypeIDFirst {
				wt.Elem += __wiretype.TypeID(firstAdded)
			}
			d = wt
		case __wiretype.MapType:
			if wt.Key >= __wiretype.TypeIDFirst {
				wt.Key += __wiretype.TypeID(firstAdded)
			}
			if wt.Elem >= __wiretype.TypeIDFirst {
				wt.Elem += __wiretype.TypeID(firstAdded)
			}
			d = wt
		case __wiretype.StructType:
			for i, fld := range wt.Fields {
				if fld.Type >= __wiretype.TypeIDFirst {
					wt.Fields[i].Type += __wiretype.TypeID(firstAdded)
				}
			}
			d = wt
			// NOTE: other types are missing, but we are upgrading anyways.
		}
		result.TypeDefs = append(result.TypeDefs, d)
	}

	return result, nil
}
