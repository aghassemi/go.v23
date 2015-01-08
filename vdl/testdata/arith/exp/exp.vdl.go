// This file was auto-generated by the veyron vdl tool.
// Source: exp.vdl

// Package exp is used to test that embedding interfaces works across packages.
// The arith.Calculator vdl interface embeds the Exp interface.
package exp

import (
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

// ExpClientMethods is the client interface
// containing Exp methods.
type ExpClientMethods interface {
	Exp(ctx *__context.T, x float64, opts ...__ipc.CallOpt) (float64, error)
}

// ExpClientStub adds universal methods to ExpClientMethods.
type ExpClientStub interface {
	ExpClientMethods
	__ipc.UniversalServiceMethods
}

// ExpClient returns a client stub for Exp.
func ExpClient(name string, opts ...__ipc.BindOpt) ExpClientStub {
	var client __ipc.Client
	for _, opt := range opts {
		if clientOpt, ok := opt.(__ipc.Client); ok {
			client = clientOpt
		}
	}
	return implExpClientStub{name, client}
}

type implExpClientStub struct {
	name   string
	client __ipc.Client
}

func (c implExpClientStub) c(ctx *__context.T) __ipc.Client {
	if c.client != nil {
		return c.client
	}
	return __veyron2.GetClient(ctx)
}

func (c implExpClientStub) Exp(ctx *__context.T, i0 float64, opts ...__ipc.CallOpt) (o0 float64, err error) {
	var call __ipc.Call
	if call, err = c.c(ctx).StartCall(ctx, c.name, "Exp", []interface{}{i0}, opts...); err != nil {
		return
	}
	if ierr := call.Finish(&o0, &err); ierr != nil {
		err = ierr
	}
	return
}

func (c implExpClientStub) Signature(ctx *__context.T, opts ...__ipc.CallOpt) (o0 __ipc.ServiceSignature, err error) {
	var call __ipc.Call
	if call, err = c.c(ctx).StartCall(ctx, c.name, "Signature", nil, opts...); err != nil {
		return
	}
	if ierr := call.Finish(&o0, &err); ierr != nil {
		err = ierr
	}
	return
}

// ExpServerMethods is the interface a server writer
// implements for Exp.
type ExpServerMethods interface {
	Exp(ctx __ipc.ServerContext, x float64) (float64, error)
}

// ExpServerStubMethods is the server interface containing
// Exp methods, as expected by ipc.Server.
// There is no difference between this interface and ExpServerMethods
// since there are no streaming methods.
type ExpServerStubMethods ExpServerMethods

// ExpServerStub adds universal methods to ExpServerStubMethods.
type ExpServerStub interface {
	ExpServerStubMethods
	// Describe the Exp interfaces.
	Describe__() []__ipc.InterfaceDesc
	// Signature will be replaced with Describe__.
	Signature(ctx __ipc.ServerContext) (__ipc.ServiceSignature, error)
}

// ExpServer returns a server stub for Exp.
// It converts an implementation of ExpServerMethods into
// an object that may be used by ipc.Server.
func ExpServer(impl ExpServerMethods) ExpServerStub {
	stub := implExpServerStub{
		impl: impl,
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

type implExpServerStub struct {
	impl ExpServerMethods
	gs   *__ipc.GlobState
}

func (s implExpServerStub) Exp(ctx __ipc.ServerContext, i0 float64) (float64, error) {
	return s.impl.Exp(ctx, i0)
}

func (s implExpServerStub) Globber() *__ipc.GlobState {
	return s.gs
}

func (s implExpServerStub) Describe__() []__ipc.InterfaceDesc {
	return []__ipc.InterfaceDesc{ExpDesc}
}

// ExpDesc describes the Exp interface.
var ExpDesc __ipc.InterfaceDesc = descExp

// descExp hides the desc to keep godoc clean.
var descExp = __ipc.InterfaceDesc{
	Name:    "Exp",
	PkgPath: "v.io/core/veyron2/vdl/testdata/arith/exp",
	Methods: []__ipc.MethodDesc{
		{
			Name: "Exp",
			InArgs: []__ipc.ArgDesc{
				{"x", ``}, // float64
			},
			OutArgs: []__ipc.ArgDesc{
				{"", ``}, // float64
				{"", ``}, // error
			},
		},
	},
}

func (s implExpServerStub) Signature(ctx __ipc.ServerContext) (__ipc.ServiceSignature, error) {
	// TODO(toddw): Replace with new Describe__ implementation.
	result := __ipc.ServiceSignature{Methods: make(map[string]__ipc.MethodSignature)}
	result.Methods["Exp"] = __ipc.MethodSignature{
		InArgs: []__ipc.MethodArgument{
			{Name: "x", Type: 26},
		},
		OutArgs: []__ipc.MethodArgument{
			{Name: "", Type: 26},
			{Name: "", Type: 65},
		},
	}

	result.TypeDefs = []__vdlutil.Any{
		__wiretype.NamedPrimitiveType{Type: 0x1, Name: "error", Tags: []string(nil)}}

	return result, nil
}
