// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This file was auto-generated by the vanadium vdl tool.
// Package: access

// Package access defines types and interfaces for dynamic access control.
// Examples: "allow app to read this photo", "prevent user from modifying this
// file".
//
// Target Developers
//
// Developers creating functionality to share data or services between
// multiple users/devices/apps.
//
// Overview
//
// Vanadium objects provide GetPermissions and SetPermissions methods.  An
// AccessList contains the set of blessings that grant principals access to the
// object. All methods on objects can have "tags" on them and the AccessList
// used for the method is selected based on that tag (from a Permissions).
//
// An object can have multiple names, so GetPermissions and SetPermissions can
// be invoked on any of these names, but the object itself has a single
// AccessList.
//
// SetPermissions completely replaces the Permissions. To perform an atomic
// read-modify-write of the AccessList, use the version parameter.
//
// Conventions
//
// Service implementors should follow the conventions below to be consistent
// with other parts of Vanadium and with each other.
//
// All methods that create an object (e.g. Put, Mount, Link) should take an
// optional AccessList parameter.  If the AccessList is not specified, the new
// object, O, copies its AccessList from the parent.  Subsequent changes to the
// parent AccessList are not automatically propagated to O.  Instead, a client
// library must make recursive AccessList changes.
//
// Resolve access is required on all components of a name, except the last one,
// in order to access the object referenced by that name.  For example, for
// principal P to access the name "a/b/c", P must have resolve access to "a"
// and "a/b".
//
// The Resolve tag means that a principal can traverse that component of the
// name to access the child.  It does not give the principal permission to list
// the children via Glob or a similar method.  For example, a server might have
// an object named "home" with a child for each user of the system.  If these
// users were allowed to list the contents of "home", they could discover the
// other users of the system.  That could be a privacy violation.  Without
// Resolve, every user of the system would need read access to "home" to access
// "home/<user>".  If the user called Glob("home/*"), it would then be up to
// the server to filter out the names that the user could not access.  That
// could be a very expensive operation if there were a lot of children of
// "home".  Resolve protects these servers against potential denial of service
// attacks on these large, shared directories.
//
// Blessings allow for sweeping access changes. In particular, a blessing is
// useful for controlling access to objects that are always accessed together.
// For example, a document may have embedded images and comments, each with a
// unique name. When accessing a document, the server would generate a blessing
// that the client would use to fetch the images and comments; the images and
// comments would have this blessed identity in their AccessLists. Changes to
// the document's AccessLists are therefore "propagated" to the images and
// comments.
//
// In the future, we may add some sort of "groups" mechanism to provide an
// alternative way to express access control policies.
//
// Some services will want a concept of implicit access control. They are free
// to implement this as appropriate for their service. However, GetPermissions
// should respond with the correct Permissions. For example, a corporate file
// server would allow all employees to create their own directory and have full
// control within that directory. Employees should not be allowed to modify
// other employee directories. In other words, within the directory "home",
// employee E should be allowed to modify only "home/E". The file server doesn't
// know the list of all employees a priori, so it uses an
// implementation-specific rule to map employee identities to their home
// directory.
//
// Examples
//
//   client := access.ObjectClient(name)
//   for {
//     perms, version, err := client.GetPermissions()
//     if err != nil {
//       return err
//     }
//     perms[newTag] = AccessList{In: []security.BlessingPattern{newPattern}}
//     // Use the same version with the modified perms to ensure that no other
//     // client has modified the perms since GetPermissions returned.
//     if err := client.SetPermissions(perms, version); err != nil {
//       if verror.ErrorID(err) == verror.ErrBadVersion.Id {
//         // Another client replaced the Permissions after our GetPermissions
//         // returned. Try again.
//         continue
//       }
//       return err
//     }
//   }
package access

import (
	"fmt"
	"v.io/v23/context"
	"v.io/v23/i18n"
	"v.io/v23/security"
	"v.io/v23/uniqueid"
	"v.io/v23/vdl"
	"v.io/v23/verror"
)

var _ = __VDLInit() // Must be first; see __VDLInit comments for details.

//////////////////////////////////////////////////
// Type definitions

// AccessList represents a set of blessings that should be granted access.
//
// See also: https://vanadium.github.io/glossary.html#access-list
type AccessList struct {
	// In denotes the set of blessings (represented as BlessingPatterns) that
	// should be granted access, unless blacklisted by an entry in NotIn.
	//
	// For example:
	//   In: {"alice:family"}
	// grants access to a principal that presents at least one of
	// "alice:family", "alice:family:friend", "alice:family:friend:spouse" etc.
	// as a blessing.
	In []security.BlessingPattern
	// NotIn denotes the set of blessings (and their delegates) that
	// have been explicitly blacklisted from the In set.
	//
	// For example:
	//   In: {"alice:friend"}, NotIn: {"alice:friend:bob"}
	// grants access to principals that present "alice:friend",
	// "alice:friend:carol" etc. but NOT to a principal that presents
	// "alice:friend:bob" or "alice:friend:bob:spouse" etc.
	NotIn []string
}

func (AccessList) __VDLReflect(struct {
	Name string `vdl:"v.io/v23/security/access.AccessList"`
}) {
}

func (m *AccessList) FillVDLTarget(t vdl.Target, tt *vdl.Type) error {
	fieldsTarget1, err := t.StartFields(tt)
	if err != nil {
		return err
	}
	var var4 bool
	if len(m.In) == 0 {
		var4 = true
	}
	if var4 {
		if err := fieldsTarget1.ZeroField("In"); err != nil && err != vdl.ErrFieldNoExist {
			return err
		}
	} else {
		keyTarget2, fieldTarget3, err := fieldsTarget1.StartField("In")
		if err != vdl.ErrFieldNoExist {
			if err != nil {
				return err
			}

			listTarget5, err := fieldTarget3.StartList(tt.NonOptional().Field(0).Type, len(m.In))
			if err != nil {
				return err
			}
			for i, elem7 := range m.In {
				elemTarget6, err := listTarget5.StartElem(i)
				if err != nil {
					return err
				}

				if err := elem7.FillVDLTarget(elemTarget6, tt.NonOptional().Field(0).Type.Elem()); err != nil {
					return err
				}
				if err := listTarget5.FinishElem(elemTarget6); err != nil {
					return err
				}
			}
			if err := fieldTarget3.FinishList(listTarget5); err != nil {
				return err
			}
			if err := fieldsTarget1.FinishField(keyTarget2, fieldTarget3); err != nil {
				return err
			}
		}
	}
	var var10 bool
	if len(m.NotIn) == 0 {
		var10 = true
	}
	if var10 {
		if err := fieldsTarget1.ZeroField("NotIn"); err != nil && err != vdl.ErrFieldNoExist {
			return err
		}
	} else {
		keyTarget8, fieldTarget9, err := fieldsTarget1.StartField("NotIn")
		if err != vdl.ErrFieldNoExist {
			if err != nil {
				return err
			}

			listTarget11, err := fieldTarget9.StartList(tt.NonOptional().Field(1).Type, len(m.NotIn))
			if err != nil {
				return err
			}
			for i, elem13 := range m.NotIn {
				elemTarget12, err := listTarget11.StartElem(i)
				if err != nil {
					return err
				}
				if err := elemTarget12.FromString(string(elem13), tt.NonOptional().Field(1).Type.Elem()); err != nil {
					return err
				}
				if err := listTarget11.FinishElem(elemTarget12); err != nil {
					return err
				}
			}
			if err := fieldTarget9.FinishList(listTarget11); err != nil {
				return err
			}
			if err := fieldsTarget1.FinishField(keyTarget8, fieldTarget9); err != nil {
				return err
			}
		}
	}
	if err := t.FinishFields(fieldsTarget1); err != nil {
		return err
	}
	return nil
}

func (m *AccessList) MakeVDLTarget() vdl.Target {
	return &AccessListTarget{Value: m}
}

type AccessListTarget struct {
	Value       *AccessList
	inTarget    __VDLTarget1_list
	notInTarget vdl.StringSliceTarget
	vdl.TargetBase
	vdl.FieldsTargetBase
}

func (t *AccessListTarget) StartFields(tt *vdl.Type) (vdl.FieldsTarget, error) {

	if ttWant := vdl.TypeOf((*AccessList)(nil)).Elem(); !vdl.Compatible(tt, ttWant) {
		return nil, fmt.Errorf("type %v incompatible with %v", tt, ttWant)
	}
	return t, nil
}
func (t *AccessListTarget) StartField(name string) (key, field vdl.Target, _ error) {
	switch name {
	case "In":
		t.inTarget.Value = &t.Value.In
		target, err := &t.inTarget, error(nil)
		return nil, target, err
	case "NotIn":
		t.notInTarget.Value = &t.Value.NotIn
		target, err := &t.notInTarget, error(nil)
		return nil, target, err
	default:
		return nil, nil, vdl.ErrFieldNoExist
	}
}
func (t *AccessListTarget) FinishField(_, _ vdl.Target) error {
	return nil
}
func (t *AccessListTarget) ZeroField(name string) error {
	switch name {
	case "In":
		t.Value.In = []security.BlessingPattern(nil)
		return nil
	case "NotIn":
		t.Value.NotIn = []string(nil)
		return nil
	default:
		return vdl.ErrFieldNoExist
	}
}
func (t *AccessListTarget) FinishFields(_ vdl.FieldsTarget) error {

	return nil
}

// []security.BlessingPattern
type __VDLTarget1_list struct {
	Value      *[]security.BlessingPattern
	elemTarget security.BlessingPatternTarget
	vdl.TargetBase
	vdl.ListTargetBase
}

func (t *__VDLTarget1_list) StartList(tt *vdl.Type, len int) (vdl.ListTarget, error) {

	if ttWant := vdl.TypeOf((*[]security.BlessingPattern)(nil)); !vdl.Compatible(tt, ttWant) {
		return nil, fmt.Errorf("type %v incompatible with %v", tt, ttWant)
	}
	if cap(*t.Value) < len {
		*t.Value = make([]security.BlessingPattern, len)
	} else {
		*t.Value = (*t.Value)[:len]
	}
	return t, nil
}
func (t *__VDLTarget1_list) StartElem(index int) (elem vdl.Target, _ error) {
	t.elemTarget.Value = &(*t.Value)[index]
	target, err := &t.elemTarget, error(nil)
	return target, err
}
func (t *__VDLTarget1_list) FinishElem(elem vdl.Target) error {
	return nil
}
func (t *__VDLTarget1_list) FinishList(elem vdl.ListTarget) error {

	return nil
}

func (x AccessList) VDLIsZero() bool {
	if len(x.In) != 0 {
		return false
	}
	if len(x.NotIn) != 0 {
		return false
	}
	return true
}

func (x AccessList) VDLWrite(enc vdl.Encoder) error {
	if err := enc.StartValue(vdl.TypeOf((*AccessList)(nil)).Elem()); err != nil {
		return err
	}
	if len(x.In) != 0 {
		if err := enc.NextField("In"); err != nil {
			return err
		}
		if err := __VDLWriteAnon_list_1(enc, x.In); err != nil {
			return err
		}
	}
	if len(x.NotIn) != 0 {
		if err := enc.NextField("NotIn"); err != nil {
			return err
		}
		if err := __VDLWriteAnon_list_2(enc, x.NotIn); err != nil {
			return err
		}
	}
	if err := enc.NextField(""); err != nil {
		return err
	}
	return enc.FinishValue()
}

func __VDLWriteAnon_list_1(enc vdl.Encoder, x []security.BlessingPattern) error {
	if err := enc.StartValue(vdl.TypeOf((*[]security.BlessingPattern)(nil))); err != nil {
		return err
	}
	if err := enc.SetLenHint(len(x)); err != nil {
		return err
	}
	for i := 0; i < len(x); i++ {
		if err := enc.NextEntry(false); err != nil {
			return err
		}
		if err := x[i].VDLWrite(enc); err != nil {
			return err
		}
	}
	if err := enc.NextEntry(true); err != nil {
		return err
	}
	return enc.FinishValue()
}

func __VDLWriteAnon_list_2(enc vdl.Encoder, x []string) error {
	if err := enc.StartValue(vdl.TypeOf((*[]string)(nil))); err != nil {
		return err
	}
	if err := enc.SetLenHint(len(x)); err != nil {
		return err
	}
	for i := 0; i < len(x); i++ {
		if err := enc.NextEntry(false); err != nil {
			return err
		}
		if err := enc.StartValue(vdl.StringType); err != nil {
			return err
		}
		if err := enc.EncodeString(x[i]); err != nil {
			return err
		}
		if err := enc.FinishValue(); err != nil {
			return err
		}
	}
	if err := enc.NextEntry(true); err != nil {
		return err
	}
	return enc.FinishValue()
}

func (x *AccessList) VDLRead(dec vdl.Decoder) error {
	*x = AccessList{}
	if err := dec.StartValue(); err != nil {
		return err
	}
	if (dec.StackDepth() == 1 || dec.IsAny()) && !vdl.Compatible(vdl.TypeOf(*x), dec.Type()) {
		return fmt.Errorf("incompatible struct %T, from %v", *x, dec.Type())
	}
	for {
		f, err := dec.NextField()
		if err != nil {
			return err
		}
		switch f {
		case "":
			return dec.FinishValue()
		case "In":
			if err := __VDLReadAnon_list_1(dec, &x.In); err != nil {
				return err
			}
		case "NotIn":
			if err := __VDLReadAnon_list_2(dec, &x.NotIn); err != nil {
				return err
			}
		default:
			if err := dec.SkipValue(); err != nil {
				return err
			}
		}
	}
}

func __VDLReadAnon_list_1(dec vdl.Decoder, x *[]security.BlessingPattern) error {
	if err := dec.StartValue(); err != nil {
		return err
	}
	if (dec.StackDepth() == 1 || dec.IsAny()) && !vdl.Compatible(vdl.TypeOf(*x), dec.Type()) {
		return fmt.Errorf("incompatible list %T, from %v", *x, dec.Type())
	}
	switch len := dec.LenHint(); {
	case len > 0:
		*x = make([]security.BlessingPattern, 0, len)
	default:
		*x = nil
	}
	for {
		switch done, err := dec.NextEntry(); {
		case err != nil:
			return err
		case done:
			return dec.FinishValue()
		}
		var elem security.BlessingPattern
		if err := elem.VDLRead(dec); err != nil {
			return err
		}
		*x = append(*x, elem)
	}
}

func __VDLReadAnon_list_2(dec vdl.Decoder, x *[]string) error {
	if err := dec.StartValue(); err != nil {
		return err
	}
	if (dec.StackDepth() == 1 || dec.IsAny()) && !vdl.Compatible(vdl.TypeOf(*x), dec.Type()) {
		return fmt.Errorf("incompatible list %T, from %v", *x, dec.Type())
	}
	switch len := dec.LenHint(); {
	case len > 0:
		*x = make([]string, 0, len)
	default:
		*x = nil
	}
	for {
		switch done, err := dec.NextEntry(); {
		case err != nil:
			return err
		case done:
			return dec.FinishValue()
		}
		var elem string
		if err := dec.StartValue(); err != nil {
			return err
		}
		var err error
		if elem, err = dec.DecodeString(); err != nil {
			return err
		}
		if err := dec.FinishValue(); err != nil {
			return err
		}
		*x = append(*x, elem)
	}
}

// Permissions maps string tags to access lists specifying the blessings
// required to invoke methods with that tag.
//
// These tags are meant to add a layer of interposition between the set of
// users (blessings, specifically) and the set of methods, much like "Roles" do
// in Role Based Access Control.
// (http://en.wikipedia.org/wiki/Role-based_access_control)
type Permissions map[string]AccessList

func (Permissions) __VDLReflect(struct {
	Name string `vdl:"v.io/v23/security/access.Permissions"`
}) {
}

func (m *Permissions) FillVDLTarget(t vdl.Target, tt *vdl.Type) error {
	mapTarget1, err := t.StartMap(tt, len((*m)))
	if err != nil {
		return err
	}
	for key3, value5 := range *m {
		keyTarget2, err := mapTarget1.StartKey()
		if err != nil {
			return err
		}
		if err := keyTarget2.FromString(string(key3), tt.NonOptional().Key()); err != nil {
			return err
		}
		valueTarget4, err := mapTarget1.FinishKeyStartField(keyTarget2)
		if err != nil {
			return err
		}

		if err := value5.FillVDLTarget(valueTarget4, tt.NonOptional().Elem()); err != nil {
			return err
		}
		if err := mapTarget1.FinishField(keyTarget2, valueTarget4); err != nil {
			return err
		}
	}
	if err := t.FinishMap(mapTarget1); err != nil {
		return err
	}
	return nil
}

func (m *Permissions) MakeVDLTarget() vdl.Target {
	return &PermissionsTarget{Value: m}
}

type PermissionsTarget struct {
	Value      *Permissions
	currKey    string
	currElem   AccessList
	keyTarget  vdl.StringTarget
	elemTarget AccessListTarget
	vdl.TargetBase
	vdl.MapTargetBase
}

func (t *PermissionsTarget) StartMap(tt *vdl.Type, len int) (vdl.MapTarget, error) {

	if ttWant := vdl.TypeOf((*Permissions)(nil)); !vdl.Compatible(tt, ttWant) {
		return nil, fmt.Errorf("type %v incompatible with %v", tt, ttWant)
	}
	*t.Value = make(Permissions)
	return t, nil
}
func (t *PermissionsTarget) StartKey() (key vdl.Target, _ error) {
	t.currKey = ""
	t.keyTarget.Value = &t.currKey
	target, err := &t.keyTarget, error(nil)
	return target, err
}
func (t *PermissionsTarget) FinishKeyStartField(key vdl.Target) (field vdl.Target, _ error) {
	t.currElem = AccessList{}
	t.elemTarget.Value = &t.currElem
	target, err := &t.elemTarget, error(nil)
	return target, err
}
func (t *PermissionsTarget) FinishField(key, field vdl.Target) error {
	(*t.Value)[t.currKey] = t.currElem
	return nil
}
func (t *PermissionsTarget) FinishMap(elem vdl.MapTarget) error {
	if len(*t.Value) == 0 {
		*t.Value = nil
	}

	return nil
}

func (x Permissions) VDLIsZero() bool {
	return len(x) == 0
}

func (x Permissions) VDLWrite(enc vdl.Encoder) error {
	if err := enc.StartValue(vdl.TypeOf((*Permissions)(nil))); err != nil {
		return err
	}
	if err := enc.SetLenHint(len(x)); err != nil {
		return err
	}
	for key, elem := range x {
		if err := enc.NextEntry(false); err != nil {
			return err
		}
		if err := enc.StartValue(vdl.StringType); err != nil {
			return err
		}
		if err := enc.EncodeString(key); err != nil {
			return err
		}
		if err := enc.FinishValue(); err != nil {
			return err
		}
		if err := elem.VDLWrite(enc); err != nil {
			return err
		}
	}
	if err := enc.NextEntry(true); err != nil {
		return err
	}
	return enc.FinishValue()
}

func (x *Permissions) VDLRead(dec vdl.Decoder) error {
	if err := dec.StartValue(); err != nil {
		return err
	}
	if (dec.StackDepth() == 1 || dec.IsAny()) && !vdl.Compatible(vdl.TypeOf(*x), dec.Type()) {
		return fmt.Errorf("incompatible map %T, from %v", *x, dec.Type())
	}
	var tmpMap Permissions
	if len := dec.LenHint(); len > 0 {
		tmpMap = make(Permissions, len)
	}
	for {
		switch done, err := dec.NextEntry(); {
		case err != nil:
			return err
		case done:
			*x = tmpMap
			return dec.FinishValue()
		}
		var key string
		{
			if err := dec.StartValue(); err != nil {
				return err
			}
			var err error
			if key, err = dec.DecodeString(); err != nil {
				return err
			}
			if err := dec.FinishValue(); err != nil {
				return err
			}
		}
		var elem AccessList
		{
			if err := elem.VDLRead(dec); err != nil {
				return err
			}
		}
		if tmpMap == nil {
			tmpMap = make(Permissions)
		}
		tmpMap[key] = elem
	}
}

// Tag is used to associate methods with an AccessList in a Permissions.
//
// While services can define their own tag type and values, many
// services should be able to use the type and values defined in
// this package.
type Tag string

func (Tag) __VDLReflect(struct {
	Name string `vdl:"v.io/v23/security/access.Tag"`
}) {
}

func (m *Tag) FillVDLTarget(t vdl.Target, tt *vdl.Type) error {
	if err := t.FromString(string((*m)), tt); err != nil {
		return err
	}
	return nil
}

func (m *Tag) MakeVDLTarget() vdl.Target {
	return &TagTarget{Value: m}
}

type TagTarget struct {
	Value *Tag
	vdl.TargetBase
}

func (t *TagTarget) FromString(src string, tt *vdl.Type) error {

	if ttWant := vdl.TypeOf((*Tag)(nil)); !vdl.Compatible(tt, ttWant) {
		return fmt.Errorf("type %v incompatible with %v", tt, ttWant)
	}
	*t.Value = Tag(src)

	return nil
}

func (x Tag) VDLIsZero() bool {
	return x == ""
}

func (x Tag) VDLWrite(enc vdl.Encoder) error {
	if err := enc.StartValue(vdl.TypeOf((*Tag)(nil))); err != nil {
		return err
	}
	if err := enc.EncodeString(string(x)); err != nil {
		return err
	}
	return enc.FinishValue()
}

func (x *Tag) VDLRead(dec vdl.Decoder) error {
	if err := dec.StartValue(); err != nil {
		return err
	}
	tmp, err := dec.DecodeString()
	if err != nil {
		return err
	}
	*x = Tag(tmp)
	return dec.FinishValue()
}

//////////////////////////////////////////////////
// Const definitions

const Admin = Tag("Admin")     // Operations that require privileged access for object administration.
const Debug = Tag("Debug")     // Operations that return debugging information (e.g., logs, statistics etc.) about the object.
const Read = Tag("Read")       // Operations that do not mutate the state of the object.
const Write = Tag("Write")     // Operations that mutate the state of the object.
const Resolve = Tag("Resolve") // Operations involving namespace navigation.
// AccessTagCaveat represents a caveat that validates iff the method being invoked has
// at least one of the tags listed in the caveat.
var AccessTagCaveat = security.CaveatDescriptor{
	Id: uniqueid.Id{
		239,
		205,
		227,
		117,
		20,
		22,
		199,
		59,
		24,
		156,
		232,
		156,
		204,
		147,
		128,
		0,
	},
	ParamType: vdl.TypeOf((*[]Tag)(nil)),
}

//////////////////////////////////////////////////
// Error definitions

var (

	// The AccessList is too big.  Use groups to represent large sets of principals.
	ErrTooBig                    = verror.Register("v.io/v23/security/access.TooBig", verror.NoRetry, "{1:}{2:} AccessList is too big")
	ErrNoPermissions             = verror.Register("v.io/v23/security/access.NoPermissions", verror.NoRetry, "{1:}{2:} {3} does not have {5} access (rejected blessings: {4})")
	ErrAccessListMatch           = verror.Register("v.io/v23/security/access.AccessListMatch", verror.NoRetry, "{1:}{2:} {3} does not match the access list (rejected blessings: {4})")
	ErrUnenforceablePatterns     = verror.Register("v.io/v23/security/access.UnenforceablePatterns", verror.NoRetry, "{1:}{2:} AccessList contains the following invalid or unrecognized patterns in the In list: {3}")
	ErrInvalidOpenAccessList     = verror.Register("v.io/v23/security/access.InvalidOpenAccessList", verror.NoRetry, "{1:}{2:} AccessList with the pattern ... in its In list must have no other patterns in the In or NotIn lists")
	ErrAccessTagCaveatValidation = verror.Register("v.io/v23/security/access.AccessTagCaveatValidation", verror.NoRetry, "{1:}{2:} access tags on method ({3}) do not include any of the ones in the caveat ({4}), or the method is using a different tag type")
)

// NewErrTooBig returns an error with the ErrTooBig ID.
func NewErrTooBig(ctx *context.T) error {
	return verror.New(ErrTooBig, ctx)
}

// NewErrNoPermissions returns an error with the ErrNoPermissions ID.
func NewErrNoPermissions(ctx *context.T, validBlessings []string, rejectedBlessings []security.RejectedBlessing, tag string) error {
	return verror.New(ErrNoPermissions, ctx, validBlessings, rejectedBlessings, tag)
}

// NewErrAccessListMatch returns an error with the ErrAccessListMatch ID.
func NewErrAccessListMatch(ctx *context.T, validBlessings []string, rejectedBlessings []security.RejectedBlessing) error {
	return verror.New(ErrAccessListMatch, ctx, validBlessings, rejectedBlessings)
}

// NewErrUnenforceablePatterns returns an error with the ErrUnenforceablePatterns ID.
func NewErrUnenforceablePatterns(ctx *context.T, rejectedPatterns []security.BlessingPattern) error {
	return verror.New(ErrUnenforceablePatterns, ctx, rejectedPatterns)
}

// NewErrInvalidOpenAccessList returns an error with the ErrInvalidOpenAccessList ID.
func NewErrInvalidOpenAccessList(ctx *context.T) error {
	return verror.New(ErrInvalidOpenAccessList, ctx)
}

// NewErrAccessTagCaveatValidation returns an error with the ErrAccessTagCaveatValidation ID.
func NewErrAccessTagCaveatValidation(ctx *context.T, methodTags []string, caveatTags []Tag) error {
	return verror.New(ErrAccessTagCaveatValidation, ctx, methodTags, caveatTags)
}

var __VDLInitCalled bool

// __VDLInit performs vdl initialization.  It is safe to call multiple times.
// If you have an init ordering issue, just insert the following line verbatim
// into your source files in this package, right after the "package foo" clause:
//
//    var _ = __VDLInit()
//
// The purpose of this function is to ensure that vdl initialization occurs in
// the right order, and very early in the init sequence.  In particular, vdl
// registration and package variable initialization needs to occur before
// functions like vdl.TypeOf will work properly.
//
// This function returns a dummy value, so that it can be used to initialize the
// first var in the file, to take advantage of Go's defined init order.
func __VDLInit() struct{} {
	if __VDLInitCalled {
		return struct{}{}
	}
	__VDLInitCalled = true

	// Register types.
	vdl.Register((*AccessList)(nil))
	vdl.Register((*Permissions)(nil))
	vdl.Register((*Tag)(nil))

	// Set error format strings.
	i18n.Cat().SetWithBase(i18n.LangID("en"), i18n.MsgID(ErrTooBig.ID), "{1:}{2:} AccessList is too big")
	i18n.Cat().SetWithBase(i18n.LangID("en"), i18n.MsgID(ErrNoPermissions.ID), "{1:}{2:} {3} does not have {5} access (rejected blessings: {4})")
	i18n.Cat().SetWithBase(i18n.LangID("en"), i18n.MsgID(ErrAccessListMatch.ID), "{1:}{2:} {3} does not match the access list (rejected blessings: {4})")
	i18n.Cat().SetWithBase(i18n.LangID("en"), i18n.MsgID(ErrUnenforceablePatterns.ID), "{1:}{2:} AccessList contains the following invalid or unrecognized patterns in the In list: {3}")
	i18n.Cat().SetWithBase(i18n.LangID("en"), i18n.MsgID(ErrInvalidOpenAccessList.ID), "{1:}{2:} AccessList with the pattern ... in its In list must have no other patterns in the In or NotIn lists")
	i18n.Cat().SetWithBase(i18n.LangID("en"), i18n.MsgID(ErrAccessTagCaveatValidation.ID), "{1:}{2:} access tags on method ({3}) do not include any of the ones in the caveat ({4}), or the method is using a different tag type")

	return struct{}{}
}