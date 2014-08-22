// This file was auto-generated by the veyron vdl tool.
// Source: service.vdl

// Package access defines the service and types for dynamic access control
// in Veyron.  Examples: "allow app to read this photo", "prevent user
// from modifying this file".
//
// Target Developers
//
// Developers creating functionality to share data or services between
// multiple users/devices/apps.
//
// Overview
//
// Every Veyron object supports GetACL and SetACL methods.  An ACL (Access
// Control List) contains principals, groups, and the labels that these
// principals and groups can access for that object.
//
// An object can have multiple names, so GetACL and SetACL can be invoked on
// any of these names, but the object itself has a single ACL.
//
// SetACL completely replaces the ACL.  To perform an atomic read-modify-write
// of the ACL, use the etag parameter.
//   n, err := access.BindObject(name)
//   if err != nil {
//      return err
//   }
//   for {
//     acl, etag, err := n.GetACL()
//     if err != nil {
//       return err
//     }
//     // Add newLabel to the LabelSet.
//     // TODO(kash): Update when we switch labels to strings instead of ints.
//     acl.Principals[newPattern] = acl.Principals[newPattern] | newLabel
//     // Use the same etag with the modified acl to ensure that no other client
//     // has modified the acl since GetACL returned.
//     if err := n.SetACL(acl, etag); err != nil {
//       if verror.Is(err, access.ErrBadEtag) {
//         // Another client replaced the ACL after our GetACL returned.
//         // Try again.
//         continue
//       }
//       return err
//     }
//   }
//
// Conventions
//
// Service implementors should follow the conventions below to be consistent
// with other parts of Veyron and with each other.
//
// All methods that create an object (e.g. Put, Mount, Link) should take an
// optional ACL parameter.  If the ACL is not specified, the new object, O,
// copies its ACL from the parent.  Subsequent changes to the parent ACL are
// not automatically propagated to O.  Instead, a client library could do
// recursive ACL changes if desired.
//
// security.ResolveLabel is required on all components of a name, except the
// last one, in order to access the object referenced by that name.  For
// example, for principal P to access the name "a/b/c", P must have resolve
// access to "a" and "a/b".
//
// security.ResolveLabel means that a principal can traverse that component of
// the name to access the child.  It does not give the principal permission to
// list the children via Glob or a similar method.  For example, a server
// might have an object named "home" with a child for each user of the system.
// If these users were allowed to list the contents of "home", they could
// discover the other users of the system.  That could be a privacy violation.
// Without ResolveLabel, every user of the system would need read access to
// "home" to access "home/<user>".  If the user called Glob("home/*"), it
// would then be up to the server to filter out the names that the user could
// not access.  That could be a very expensive operation if there were a lot
// of children of "home".  ResolveLabel protects these servers against
// potential denial of service attacks on these large, shared directories.
//
// Groups and blessings allow for sweeping access changes.  A group is
// suitable for saying that the same set of principals have access to a set of
// unrelated resources (e.g. docs, VMs, images).  See the Group API for a
// complete description.  A blessing is useful for controlling access to objects
// that are always accessed together.  For example, a document may have
// embedded images and comments, each with a unique name.  When accessing a
// document, the server would generate a blessing that the client would use to
// fetch the images and comments; the images and comments would have this
// blessed identity in their ACLs.  Changes to the document’s ACL are
// therefore “propagated” to the images and comments.
//
// Some services will want a concept of implicit access control.  They are
// free to implement this as is best for their service.  However, GetACL
// should respond with the correct ACL.  For example, a corporate file server
// would allow all employees to create their own directory and have full
// control within that directory.  Employees should not be allowed to modify
// other employee directories.  In other words, within the directory "home",
// employee E should be allowed to modify only "home/E".  The file server
// doesn't know the list of all employees a priori, so it uses an
// implementation-specific rule to map employee identities to their home
// directory.
package access

import (
	"veyron2/security"

	// The non-user imports are prefixed with "_gen_" to prevent collisions.
	_gen_veyron2 "veyron2"
	_gen_context "veyron2/context"
	_gen_ipc "veyron2/ipc"
	_gen_naming "veyron2/naming"
	_gen_vdlutil "veyron2/vdl/vdlutil"
	_gen_verror "veyron2/verror"
	_gen_wiretype "veyron2/wiretype"
)

// TODO(bprosnitz) Remove this line once signatures are updated to use typevals.
// It corrects a bug where _gen_wiretype is unused in VDL pacakges where only bootstrap types are used on interfaces.
const _ = _gen_wiretype.TypeIDInvalid

// The etag passed to SetACL is invalid.  Likely, another client set
// the ACL already and invalidated the etag.  Use GetACL to fetch a
// fresh etag.
const ErrBadEtag = _gen_verror.ID("veyron2/services/security/access.ErrBadEtag")

// The ACL is too big.  Use groups to represent large sets of principals.
const ErrTooBig = _gen_verror.ID("veyron2/services/security/access.ErrTooBig")

// Object provides access control for Veyron objects.
// Object is the interface the client binds and uses.
// Object_ExcludingUniversal is the interface without internal framework-added methods
// to enable embedding without method collisions.  Not to be used directly by clients.
type Object_ExcludingUniversal interface {
	// SetACL replaces the current ACL for an object.  etag allows for optional,
	// optimistic concurrency control.  If non-empty, etag's value must come
	// from GetACL.  If any client has successfully called SetACL in the
	// meantime, the etag will be stale and SetACL will fail.
	//
	// ACL objects are expected to be small.  It is up to the implementation to
	// define the exact limit, though it should probably be around 100KB.  Large
	// lists of principals should use the Group API or blessings.
	//
	// There is some ambiguity when calling SetACL on a mount point.  Does it
	// affect the mount itself or does it affect the service endpoint that the
	// mount points to?  The chosen behavior is that it affects the service
	// endpoint.  To modify the mount point's ACL, use ResolveToMountTable
	// to get an endpoint and call SetACL on that.  This means that clients
	// must know when a name refers to a mount point to change its ACL.
	SetACL(ctx _gen_context.T, acl security.ACL, etag string, opts ..._gen_ipc.CallOpt) (err error)
	// GetACL returns the complete, current ACL for an object.  The returned etag
	// can be passed to a subsequent call to SetACL for optimistic concurrency
	// control. A successful call to SetACL will invalidate etag, and the client
	// must call GetACL again to get the current etag.
	GetACL(ctx _gen_context.T, opts ..._gen_ipc.CallOpt) (acl security.ACL, etag string, err error)
}
type Object interface {
	_gen_ipc.UniversalServiceMethods
	Object_ExcludingUniversal
}

// ObjectService is the interface the server implements.
type ObjectService interface {

	// SetACL replaces the current ACL for an object.  etag allows for optional,
	// optimistic concurrency control.  If non-empty, etag's value must come
	// from GetACL.  If any client has successfully called SetACL in the
	// meantime, the etag will be stale and SetACL will fail.
	//
	// ACL objects are expected to be small.  It is up to the implementation to
	// define the exact limit, though it should probably be around 100KB.  Large
	// lists of principals should use the Group API or blessings.
	//
	// There is some ambiguity when calling SetACL on a mount point.  Does it
	// affect the mount itself or does it affect the service endpoint that the
	// mount points to?  The chosen behavior is that it affects the service
	// endpoint.  To modify the mount point's ACL, use ResolveToMountTable
	// to get an endpoint and call SetACL on that.  This means that clients
	// must know when a name refers to a mount point to change its ACL.
	SetACL(context _gen_ipc.ServerContext, acl security.ACL, etag string) (err error)
	// GetACL returns the complete, current ACL for an object.  The returned etag
	// can be passed to a subsequent call to SetACL for optimistic concurrency
	// control. A successful call to SetACL will invalidate etag, and the client
	// must call GetACL again to get the current etag.
	GetACL(context _gen_ipc.ServerContext) (acl security.ACL, etag string, err error)
}

// BindObject returns the client stub implementing the Object
// interface.
//
// If no _gen_ipc.Client is specified, the default _gen_ipc.Client in the
// global Runtime is used.
func BindObject(name string, opts ..._gen_ipc.BindOpt) (Object, error) {
	var client _gen_ipc.Client
	switch len(opts) {
	case 0:
		// Do nothing.
	case 1:
		if clientOpt, ok := opts[0].(_gen_ipc.Client); opts[0] == nil || ok {
			client = clientOpt
		} else {
			return nil, _gen_vdlutil.ErrUnrecognizedOption
		}
	default:
		return nil, _gen_vdlutil.ErrTooManyOptionsToBind
	}
	stub := &clientStubObject{defaultClient: client, name: name}

	return stub, nil
}

// NewServerObject creates a new server stub.
//
// It takes a regular server implementing the ObjectService
// interface, and returns a new server stub.
func NewServerObject(server ObjectService) interface{} {
	return &ServerStubObject{
		service: server,
	}
}

// clientStubObject implements Object.
type clientStubObject struct {
	defaultClient _gen_ipc.Client
	name          string
}

func (__gen_c *clientStubObject) client(ctx _gen_context.T) _gen_ipc.Client {
	if __gen_c.defaultClient != nil {
		return __gen_c.defaultClient
	}
	return _gen_veyron2.RuntimeFromContext(ctx).Client()
}

func (__gen_c *clientStubObject) SetACL(ctx _gen_context.T, acl security.ACL, etag string, opts ..._gen_ipc.CallOpt) (err error) {
	var call _gen_ipc.Call
	if call, err = __gen_c.client(ctx).StartCall(ctx, __gen_c.name, "SetACL", []interface{}{acl, etag}, opts...); err != nil {
		return
	}
	if ierr := call.Finish(&err); ierr != nil {
		err = ierr
	}
	return
}

func (__gen_c *clientStubObject) GetACL(ctx _gen_context.T, opts ..._gen_ipc.CallOpt) (acl security.ACL, etag string, err error) {
	var call _gen_ipc.Call
	if call, err = __gen_c.client(ctx).StartCall(ctx, __gen_c.name, "GetACL", nil, opts...); err != nil {
		return
	}
	if ierr := call.Finish(&acl, &etag, &err); ierr != nil {
		err = ierr
	}
	return
}

func (__gen_c *clientStubObject) UnresolveStep(ctx _gen_context.T, opts ..._gen_ipc.CallOpt) (reply []string, err error) {
	var call _gen_ipc.Call
	if call, err = __gen_c.client(ctx).StartCall(ctx, __gen_c.name, "UnresolveStep", nil, opts...); err != nil {
		return
	}
	if ierr := call.Finish(&reply, &err); ierr != nil {
		err = ierr
	}
	return
}

func (__gen_c *clientStubObject) Signature(ctx _gen_context.T, opts ..._gen_ipc.CallOpt) (reply _gen_ipc.ServiceSignature, err error) {
	var call _gen_ipc.Call
	if call, err = __gen_c.client(ctx).StartCall(ctx, __gen_c.name, "Signature", nil, opts...); err != nil {
		return
	}
	if ierr := call.Finish(&reply, &err); ierr != nil {
		err = ierr
	}
	return
}

func (__gen_c *clientStubObject) GetMethodTags(ctx _gen_context.T, method string, opts ..._gen_ipc.CallOpt) (reply []interface{}, err error) {
	var call _gen_ipc.Call
	if call, err = __gen_c.client(ctx).StartCall(ctx, __gen_c.name, "GetMethodTags", []interface{}{method}, opts...); err != nil {
		return
	}
	if ierr := call.Finish(&reply, &err); ierr != nil {
		err = ierr
	}
	return
}

// ServerStubObject wraps a server that implements
// ObjectService and provides an object that satisfies
// the requirements of veyron2/ipc.ReflectInvoker.
type ServerStubObject struct {
	service ObjectService
}

func (__gen_s *ServerStubObject) GetMethodTags(call _gen_ipc.ServerCall, method string) ([]interface{}, error) {
	// TODO(bprosnitz) GetMethodTags() will be replaces with Signature().
	// Note: This exhibits some weird behavior like returning a nil error if the method isn't found.
	// This will change when it is replaced with Signature().
	switch method {
	case "SetACL":
		return []interface{}{security.Label(8)}, nil
	case "GetACL":
		return []interface{}{security.Label(8)}, nil
	default:
		return nil, nil
	}
}

func (__gen_s *ServerStubObject) Signature(call _gen_ipc.ServerCall) (_gen_ipc.ServiceSignature, error) {
	result := _gen_ipc.ServiceSignature{Methods: make(map[string]_gen_ipc.MethodSignature)}
	result.Methods["GetACL"] = _gen_ipc.MethodSignature{
		InArgs: []_gen_ipc.MethodArgument{},
		OutArgs: []_gen_ipc.MethodArgument{
			{Name: "acl", Type: 69},
			{Name: "etag", Type: 3},
			{Name: "err", Type: 70},
		},
	}
	result.Methods["SetACL"] = _gen_ipc.MethodSignature{
		InArgs: []_gen_ipc.MethodArgument{
			{Name: "acl", Type: 69},
			{Name: "etag", Type: 3},
		},
		OutArgs: []_gen_ipc.MethodArgument{
			{Name: "", Type: 70},
		},
	}

	result.TypeDefs = []_gen_vdlutil.Any{
		_gen_wiretype.NamedPrimitiveType{Type: 0x3, Name: "veyron2/security.PrincipalPattern", Tags: []string(nil)}, _gen_wiretype.NamedPrimitiveType{Type: 0x34, Name: "veyron2/security.LabelSet", Tags: []string(nil)}, _gen_wiretype.MapType{Key: 0x41, Elem: 0x42, Name: "", Tags: []string(nil)}, _gen_wiretype.StructType{
			[]_gen_wiretype.FieldType{
				_gen_wiretype.FieldType{Type: 0x43, Name: "Principals"},
			},
			"veyron2/security.Entries", []string(nil)},
		_gen_wiretype.StructType{
			[]_gen_wiretype.FieldType{
				_gen_wiretype.FieldType{Type: 0x44, Name: "In"},
				_gen_wiretype.FieldType{Type: 0x44, Name: "NotIn"},
			},
			"veyron2/security.ACL", []string(nil)},
		_gen_wiretype.NamedPrimitiveType{Type: 0x1, Name: "error", Tags: []string(nil)}}

	return result, nil
}

func (__gen_s *ServerStubObject) UnresolveStep(call _gen_ipc.ServerCall) (reply []string, err error) {
	if unresolver, ok := __gen_s.service.(_gen_ipc.Unresolver); ok {
		return unresolver.UnresolveStep(call)
	}
	if call.Server() == nil {
		return
	}
	var published []string
	if published, err = call.Server().Published(); err != nil || published == nil {
		return
	}
	reply = make([]string, len(published))
	for i, p := range published {
		reply[i] = _gen_naming.Join(p, call.Name())
	}
	return
}

func (__gen_s *ServerStubObject) SetACL(call _gen_ipc.ServerCall, acl security.ACL, etag string) (err error) {
	err = __gen_s.service.SetACL(call, acl, etag)
	return
}

func (__gen_s *ServerStubObject) GetACL(call _gen_ipc.ServerCall) (acl security.ACL, etag string, err error) {
	acl, etag, err = __gen_s.service.GetACL(call)
	return
}
