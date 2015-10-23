// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package vom

import (
	"io"
	"os"
	"reflect"

	"v.io/v23/vdl"
	"v.io/v23/verror"
)

var (
	errDecodeNil                = verror.Register(pkgPath+".errDecodeNil", verror.NoRetry, "{1:}{2:} vom: invalid decode into nil interface{}{:_}")
	errDecodeNilRawValue        = verror.Register(pkgPath+".errDecodeNilRawValue", verror.NoRetry, "{1:}{2:} vom: invalid decode into nil *RawValue{:_}")
	errDecodeZeroTypeID         = verror.Register(pkgPath+".errDecodeZeroTypeID", verror.NoRetry, "{1:}{2:} vom: zero type id{:_}")
	errIndexOutOfRange          = verror.Register(pkgPath+".errIndexOutOfRange", verror.NoRetry, "{1:}{2:} vom: index out of range{:_}")
	errLeftOverBytes            = verror.Register(pkgPath+".errLeftOverBytes", verror.NoRetry, "{1:}{2:} vom: {3} leftover bytes{:_}")
	errUnexpectedControlByte    = verror.Register(pkgPath+".errUnexpectedControlByte", verror.NoRetry, "{1:}{2:} vom: unexpected control byte {3}{:_}")
	errDecodeValueUnhandledType = verror.Register(pkgPath+".errDecodeValueUnhandledType", verror.NoRetry, "{1:}{2:} vom: decodeValue unhandled type {3}{:_}")
	errIgnoreValueUnhandledType = verror.Register(pkgPath+".errIgnoreValueUnhandledType", verror.NoRetry, "{1:}{2:} vom: ignoreValue unhandled type {3}{:_}")
)

// Decoder manages the receipt and unmarshalling of typed values from the other
// side of a connection.
type Decoder struct {
	mr      *messageReader
	typeDec *TypeDecoder
}

// This is only used for debugging; add this as the first line of NewDecoder to
// dump formatted vom bytes to stdout:
//   r = teeDump(r)
func teeDump(r io.Reader) io.Reader {
	return io.TeeReader(r, NewDumper(NewDumpWriter(os.Stdout)))
}

// NewDecoder returns a new Decoder that reads from the given reader. The
// Decoder understands all formats generated by the Encoder.
func NewDecoder(r io.Reader) *Decoder {
	// When the TypeDecoder isn't shared, we always decode type messages in
	// Decoder.decodeValueType() and feed them to the TypeDecoder. That is,
	// the TypeDecoder will never read messages from the buffer. So we pass
	// a nil buffer to newTypeDecoder.
	mr := newMessageReader(newDecbuf(r))
	typeDec := newTypeDecoderInternal(mr)
	mr.SetCallbacks(typeDec.lookupType, typeDec.readSingleType)
	return &Decoder{
		mr:      mr,
		typeDec: typeDec,
	}
}

// NewDecoderWithTypeDecoder returns a new Decoder that reads from the given
// reader. Types will be decoded separately through the given typeDec.
func NewDecoderWithTypeDecoder(r io.Reader, typeDec *TypeDecoder) *Decoder {
	mr := newMessageReader(newDecbuf(r))
	mr.SetCallbacks(typeDec.lookupType, nil)
	return &Decoder{
		mr:      mr,
		typeDec: typeDec,
	}
}

// Decode reads the next value from the reader(s) and stores it in value v.
// The type of v need not exactly match the type of the originally encoded
// value; decoding succeeds as long as the values are compatible.
//
//   Types that are special-cased, only for v:
//     *RawValue  - Store raw (uninterpreted) bytes in v.
//
//   Types that are special-cased, recursively throughout v:
//     *vdl.Value    - Decode into v.
//     reflect.Value - Decode into v, which must be settable.
//
// Decoding into a RawValue captures the value in a raw form, which may be
// subsequently passed to an Encoder for transcoding.
//
// Decode(nil) always returns an error.  Use Ignore() to ignore the next value.
func (d *Decoder) Decode(v interface{}) error {
	switch tv := v.(type) {
	case nil:
		return verror.New(errDecodeNil, nil)
	case *RawValue:
		if tv == nil {
			return verror.New(errDecodeNilRawValue, nil)
		}
		return d.decodeRaw(tv)
	}
	tid, err := d.mr.StartValueMessage()
	if err != nil {
		return err
	}
	valType, err := d.typeDec.lookupType(tid)
	if err != nil {
		return err
	}
	if err := d.decodeValueMsg(valType, v); err != nil {
		return err
	}
	return d.mr.EndMessage()
}

// Ignore ignores the next value from the reader.
func (d *Decoder) Ignore() error {
	tid, err := d.mr.StartValueMessage()
	if err != nil {
		return err
	}
	valType, err := d.typeDec.lookupType(tid)
	if err != nil {
		return err
	}
	valLen, err := d.decodeValueByteLen(valType)
	if err != nil {
		return err
	}

	if err := d.mr.Skip(valLen); err != nil {
		return err
	}
	return d.mr.EndMessage()
}

func (d *Decoder) decodeRaw(raw *RawValue) error {
	tid, err := d.mr.StartValueMessage()
	if err != nil {
		return err
	}
	valType, err := d.typeDec.lookupType(tid)
	if err != nil {
		return err
	}
	valLen, err := d.decodeValueByteLen(valType)
	if err != nil {
		return err
	}
	raw.typeDec = d.typeDec
	raw.valType = valType
	if cap(raw.data) >= valLen {
		raw.data = raw.data[:valLen]
	} else {
		raw.data = make([]byte, valLen)
	}
	if err := d.mr.ReadIntoBuf(raw.data); err != nil {
		return err
	}
	return d.mr.EndMessage()
}

// decodeWireType decodes the next type definition message and returns its
// type id.
func (d *Decoder) decodeWireType(wt *wireType) (typeId, error) {
	tid, err := d.mr.StartTypeMessage()
	if err != nil {
		return 0, err
	}
	// Decode the wire type like a regular value.
	if err := d.decodeValueMsg(wireTypeType, wt); err != nil {
		return 0, err
	}
	return tid, d.mr.EndMessage()
}

// decodeValueByteLen returns the byte length of the next value.
func (d *Decoder) decodeValueByteLen(tt *vdl.Type) (int, error) {
	if hasChunkLen(tt) {
		// Use the explicit message length.
		if d.mr.version == Version81 {
			// TODO(bprosnitz) Implement this for version 81
			panic("not yet implemented for version 81")
		}
		return d.mr.buf.lim, nil
	}
	// No explicit message length, but the length can be computed.
	switch {
	case tt.Kind() == vdl.Byte:
		// Single byte is always encoded as 1 byte.
		return 1, nil
	case tt.Kind() == vdl.Array && tt.IsBytes():
		// Byte arrays are exactly their length and encoded with 1-byte header.
		return tt.Len() + 1, nil
	case tt.Kind() == vdl.String || tt.IsBytes():
		// Strings and byte lists are encoded with a length header.
		strlen, bytelen, err := binaryPeekUint(d.mr)
		switch {
		case err != nil:
			return 0, err
		case strlen > maxBinaryMsgLen:
			return 0, verror.New(errMsgLen, nil)
		}
		return int(strlen) + bytelen, nil
	default:
		// Must be a primitive, which is encoded as an underlying uint.
		return binaryPeekUintByteLen(d.mr)
	}
}

// decodeValueMsg decodes the rest of the message assuming type tt
func (d *Decoder) decodeValueMsg(tt *vdl.Type, v interface{}) error {
	target, err := vdl.ReflectTarget(reflect.ValueOf(v))
	if err != nil {
		return err
	}
	return d.decodeValue(tt, target)
}

// decodeValue decodes the rest of the message assuming type tt.
func (d *Decoder) decodeValue(tt *vdl.Type, target vdl.Target) error {
	ttFrom := tt
	if tt.Kind() == vdl.Optional {
		// If the type is optional, we expect to see either WireCtrlNil or the actual
		// value, but not both.  And thus, we can just peek for the WireCtrlNil here.
		switch ctrl, err := binaryPeekControl(d.mr); {
		case err != nil:
			return err
		case ctrl == WireCtrlNil:
			d.mr.Skip(1)
			return target.FromNil(ttFrom)
		}
		tt = tt.Elem()
	}
	if tt.IsBytes() {
		len, err := binaryDecodeLenOrArrayLen(d.mr, tt)
		if err != nil {
			return err
		}
		// TODO(toddw): remove allocation
		buf := make([]byte, len)
		if err := d.mr.ReadIntoBuf(buf); err != nil {
			return err
		}
		return target.FromBytes(buf, ttFrom)
	}
	switch kind := tt.Kind(); kind {
	case vdl.Bool:
		v, err := binaryDecodeBool(d.mr)
		if err != nil {
			return err
		}
		return target.FromBool(v, ttFrom)
	case vdl.Byte:
		v, err := d.mr.ReadByte()
		if err != nil {
			return err
		}
		return target.FromUint(uint64(v), ttFrom)
	case vdl.Uint16, vdl.Uint32, vdl.Uint64:
		v, err := binaryDecodeUint(d.mr)
		if err != nil {
			return err
		}
		return target.FromUint(v, ttFrom)
	case vdl.Int16, vdl.Int32, vdl.Int64:
		v, err := binaryDecodeInt(d.mr)
		if err != nil {
			return err
		}
		return target.FromInt(v, ttFrom)
	case vdl.Float32, vdl.Float64:
		v, err := binaryDecodeFloat(d.mr)
		if err != nil {
			return err
		}
		return target.FromFloat(v, ttFrom)
	case vdl.Complex64, vdl.Complex128:
		re, err := binaryDecodeFloat(d.mr)
		if err != nil {
			return err
		}
		im, err := binaryDecodeFloat(d.mr)
		if err != nil {
			return err
		}
		return target.FromComplex(complex(re, im), ttFrom)
	case vdl.String:
		v, err := binaryDecodeString(d.mr)
		if err != nil {
			return err
		}
		return target.FromString(v, ttFrom)
	case vdl.Enum:
		index, err := binaryDecodeUint(d.mr)
		switch {
		case err != nil:
			return err
		case index >= uint64(tt.NumEnumLabel()):
			return verror.New(errIndexOutOfRange, nil)
		}
		return target.FromEnumLabel(tt.EnumLabel(int(index)), ttFrom)
	case vdl.TypeObject:
		x, err := binaryDecodeUint(d.mr)
		if err != nil {
			return err
		}
		var typeobject *vdl.Type
		if d.mr.version == Version80 {
			typeobject, err = d.typeDec.lookupType(typeId(x))
		} else {
			typeobject, err = d.mr.ReferencedType(x)
		}
		if err != nil {
			return err
		}
		return target.FromTypeObject(typeobject)
	case vdl.Array, vdl.List:
		len, err := binaryDecodeLenOrArrayLen(d.mr, tt)
		if err != nil {
			return err
		}
		listTarget, err := target.StartList(ttFrom, len)
		if err != nil {
			return err
		}
		for ix := 0; ix < len; ix++ {
			elem, err := listTarget.StartElem(ix)
			if err != nil {
				return err
			}
			if err := d.decodeValue(tt.Elem(), elem); err != nil {
				return err
			}
			if err := listTarget.FinishElem(elem); err != nil {
				return err
			}
		}
		return target.FinishList(listTarget)
	case vdl.Set:
		len, err := binaryDecodeLen(d.mr)
		if err != nil {
			return err
		}
		setTarget, err := target.StartSet(ttFrom, len)
		if err != nil {
			return err
		}
		for ix := 0; ix < len; ix++ {
			key, err := setTarget.StartKey()
			if err != nil {
				return err
			}
			if err := d.decodeValue(tt.Key(), key); err != nil {
				return err
			}
			switch err := setTarget.FinishKey(key); {
			case err == vdl.ErrFieldNoExist:
				continue
			case err != nil:
				return err
			}
		}
		return target.FinishSet(setTarget)
	case vdl.Map:
		len, err := binaryDecodeLen(d.mr)
		if err != nil {
			return err
		}
		mapTarget, err := target.StartMap(ttFrom, len)
		if err != nil {
			return err
		}
		for ix := 0; ix < len; ix++ {
			key, err := mapTarget.StartKey()
			if err != nil {
				return err
			}
			if err := d.decodeValue(tt.Key(), key); err != nil {
				return err
			}
			switch field, err := mapTarget.FinishKeyStartField(key); {
			case err == vdl.ErrFieldNoExist:
				if err := d.ignoreValue(tt.Elem()); err != nil {
					return err
				}
			case err != nil:
				return err
			default:
				if err := d.decodeValue(tt.Elem(), field); err != nil {
					return err
				}
				if err := mapTarget.FinishField(key, field); err != nil {
					return err
				}
			}
		}
		return target.FinishMap(mapTarget)
	case vdl.Struct:
		fieldsTarget, err := target.StartFields(ttFrom)
		if err != nil {
			return err
		}
		// Loop through decoding the 0-based field index and corresponding field.
		decodedFields := make([]bool, tt.NumField())
		for {
			index, ctrl, err := binaryDecodeUintWithControl(d.mr)
			switch {
			case err != nil:
				return err
			case ctrl == WireCtrlEnd:
				// Fill not-yet-decoded fields with their zero values.
				for index, decoded := range decodedFields {
					if decoded {
						continue
					}
					ttfield := tt.Field(index)
					switch key, field, err := fieldsTarget.StartField(ttfield.Name); {
					case err == vdl.ErrFieldNoExist:
						// Ignore it.
					case err != nil:
						return err
					default:
						if err := vdl.FromValue(field, vdl.ZeroValue(ttfield.Type)); err != nil {
							return err
						}
						if err := fieldsTarget.FinishField(key, field); err != nil {
							return err
						}
					}
				}
				return target.FinishFields(fieldsTarget)
			case ctrl != 0:
				return verror.New(errUnexpectedControlByte, nil, ctrl)
			case index >= uint64(tt.NumField()):
				return verror.New(errIndexOutOfRange, nil)
			}
			ttfield := tt.Field(int(index))
			switch key, field, err := fieldsTarget.StartField(ttfield.Name); {
			case err == vdl.ErrFieldNoExist:
				if err := d.ignoreValue(ttfield.Type); err != nil {
					return err
				}
			case err != nil:
				return err
			default:
				if err := d.decodeValue(ttfield.Type, field); err != nil {
					return err
				}
				if err := fieldsTarget.FinishField(key, field); err != nil {
					return err
				}
			}
			decodedFields[index] = true
		}
	case vdl.Union:
		fieldsTarget, err := target.StartFields(ttFrom)
		if err != nil {
			return err
		}
		index, err := binaryDecodeUint(d.mr)
		switch {
		case err != nil:
			return err
		case index >= uint64(tt.NumField()):
			return verror.New(errIndexOutOfRange, nil)
		}
		ttfield := tt.Field(int(index))
		key, field, err := fieldsTarget.StartField(ttfield.Name)
		if err != nil {
			return err
		}
		if err := d.decodeValue(ttfield.Type, field); err != nil {
			return err
		}
		if err := fieldsTarget.FinishField(key, field); err != nil {
			return err
		}
		return target.FinishFields(fieldsTarget)
	case vdl.Any:
		var elemType *vdl.Type
		switch x, ctrl, err := binaryDecodeUintWithControl(d.mr); {
		case err != nil:
			return err
		case ctrl == WireCtrlNil:
			return target.FromNil(tt)
		case ctrl != 0:
			return verror.New(errUnexpectedControlByte, nil, ctrl)
		case d.mr.version == Version80:
			if elemType, err = d.typeDec.lookupType(typeId(x)); err != nil {
				return err
			}
		default:
			if elemType, err = d.mr.ReferencedType(x); err != nil {
				return err
			}
		}
		return d.decodeValue(elemType, target)
	default:
		panic(verror.New(errDecodeValueUnhandledType, nil, tt))
	}
}

// ignoreValue ignores the rest of the value of type t. This is used to ignore
// unknown struct fields.
func (d *Decoder) ignoreValue(tt *vdl.Type) error {
	if tt.IsBytes() {
		len, err := binaryDecodeLenOrArrayLen(d.mr, tt)
		if err != nil {
			return err
		}
		return d.mr.Skip(len)
	}
	switch kind := tt.Kind(); kind {
	case vdl.Bool, vdl.Byte:
		return d.mr.Skip(1)
	case vdl.Uint16, vdl.Uint32, vdl.Uint64, vdl.Int16, vdl.Int32, vdl.Int64, vdl.Float32, vdl.Float64, vdl.Enum, vdl.TypeObject:
		// The underlying encoding of all these types is based on uint.
		return binaryIgnoreUint(d.mr)
	case vdl.Complex64, vdl.Complex128:
		// Complex is encoded as two floats, so we can simply ignore two uints.
		if err := binaryIgnoreUint(d.mr); err != nil {
			return err
		}
		return binaryIgnoreUint(d.mr)
	case vdl.String:
		return binaryIgnoreString(d.mr)
	case vdl.Array, vdl.List, vdl.Set, vdl.Map:
		len, err := binaryDecodeLenOrArrayLen(d.mr, tt)
		if err != nil {
			return err
		}
		for ix := 0; ix < len; ix++ {
			if kind == vdl.Set || kind == vdl.Map {
				if err := d.ignoreValue(tt.Key()); err != nil {
					return err
				}
			}
			if kind == vdl.Array || kind == vdl.List || kind == vdl.Map {
				if err := d.ignoreValue(tt.Elem()); err != nil {
					return err
				}
			}
		}
		return nil
	case vdl.Struct:
		// Loop through decoding the 0-based field index and corresponding field.
		for {
			switch index, ctrl, err := binaryDecodeUintWithControl(d.mr); {
			case err != nil:
				return err
			case ctrl == WireCtrlEnd:
				return nil
			case ctrl != 0:
				return verror.New(errUnexpectedControlByte, nil, ctrl)
			case index >= uint64(tt.NumField()):
				return verror.New(errIndexOutOfRange, nil)
			default:
				ttfield := tt.Field(int(index))
				if err := d.ignoreValue(ttfield.Type); err != nil {
					return err
				}
			}
		}
	case vdl.Union:
		switch index, err := binaryDecodeUint(d.mr); {
		case err != nil:
			return err
		case index >= uint64(tt.NumField()):
			return verror.New(errIndexOutOfRange, nil)
		default:
			ttfield := tt.Field(int(index))
			return d.ignoreValue(ttfield.Type)
		}
	case vdl.Any:
		var elemType *vdl.Type
		switch x, ctrl, err := binaryDecodeUintWithControl(d.mr); {
		case err != nil:
			return err
		case ctrl == WireCtrlNil:
			return nil
		case ctrl != 0:
			return verror.New(errUnexpectedControlByte, nil, ctrl)
		case d.mr.version == Version80:
			if elemType, err = d.typeDec.lookupType(typeId(x)); err != nil {
				return err
			}
		default:
			if elemType, err = d.mr.ReferencedType(x); err != nil {
				return err
			}
		}
		return d.ignoreValue(elemType)
	default:
		panic(verror.New(errIgnoreValueUnhandledType, nil, tt))
	}
}
