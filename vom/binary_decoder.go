package vom

import (
	"errors"
	"fmt"
	"reflect"

	"v.io/v23/vdl"
)

var (
	errDecodeZeroTypeID = errors.New("vom: type ID 0")
	errIndexOutOfRange  = errors.New("vom: index out of range")
)

type binaryDecoder struct {
	buf       *decbuf
	recvTypes *decoderTypes
}

func newBinaryDecoder(buf *decbuf, types *decoderTypes) *binaryDecoder {
	return &binaryDecoder{buf, types}
}

func (d *binaryDecoder) Decode(target vdl.Target) error {
	valType, err := d.decodeValueType()
	if err != nil {
		return err
	}
	return d.decodeValueMsg(valType, target)
}

func (d *binaryDecoder) DecodeRaw(raw *RawValue) error {
	valType, err := d.decodeValueType()
	if err != nil {
		return err
	}
	valLen, err := d.decodeValueByteLen(valType)
	if err != nil {
		return err
	}
	raw.recvTypes = d.recvTypes
	raw.valType = valType
	if cap(raw.data) >= valLen {
		raw.data = raw.data[:valLen]
	} else {
		raw.data = make([]byte, valLen)
	}
	return d.buf.ReadFull(raw.data)
}

func (d *binaryDecoder) Ignore() error {
	valType, err := d.decodeValueType()
	if err != nil {
		return err
	}
	valLen, err := d.decodeValueByteLen(valType)
	if err != nil {
		return err
	}
	return d.buf.Skip(valLen)
}

// decodeValueType returns the type of the next value message.  Any type
// definition messages it encounters along the way are decoded and added to
// recvTypes.
func (d *binaryDecoder) decodeValueType() (*vdl.Type, error) {
	for {
		id, err := binaryDecodeInt(d.buf)
		if err != nil {
			return nil, err
		}
		switch {
		case id == 0:
			return nil, errDecodeZeroTypeID
		case id > 0:
			// This is a value message, the typeID is +id.
			tid := typeID(+id)
			tt, err := d.recvTypes.LookupOrBuildType(tid)
			if err != nil {
				return nil, err
			}
			return tt, nil
		}
		// This is a type message, the typeID is -id.
		tid := typeID(-id)
		// Decode the wireType like a regular value, and store it in recvTypes.  The
		// type will actually be built when a value message arrives using this tid.
		var wt wireType
		target, err := vdl.ReflectTarget(reflect.ValueOf(&wt))
		if err != nil {
			return nil, err
		}
		if err := d.decodeValueMsg(wireTypeType, target); err != nil {
			return nil, err
		}
		if err := d.recvTypes.AddWireType(tid, wt); err != nil {
			return nil, err
		}
	}
}

// decodeValueByteLen returns the byte length of the next value.
func (d *binaryDecoder) decodeValueByteLen(tt *vdl.Type) (int, error) {
	if hasBinaryMsgLen(tt) {
		// Use the explicit message length.
		msgLen, err := binaryDecodeLen(d.buf)
		if err != nil {
			return 0, err
		}
		return msgLen, nil
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
		strlen, bytelen, err := binaryPeekUint(d.buf)
		switch {
		case err != nil:
			return 0, err
		case strlen > maxBinaryMsgLen:
			return 0, errMsgLen
		}
		return int(strlen) + bytelen, nil
	default:
		// Must be a primitive, which is encoded as an underlying uint.
		return binaryPeekUintByteLen(d.buf)
	}
}

// decodeValueMsg decodes the rest of the message assuming type t, handling the
// optional message length.
func (d *binaryDecoder) decodeValueMsg(tt *vdl.Type, target vdl.Target) error {
	if hasBinaryMsgLen(tt) {
		msgLen, err := binaryDecodeLen(d.buf)
		if err != nil {
			return err
		}
		d.buf.SetLimit(msgLen)
	}
	err := d.decodeValue(tt, target)
	leftover := d.buf.RemoveLimit()
	switch {
	case err != nil:
		return err
	case leftover > 0:
		return fmt.Errorf("vom: %d leftover bytes", leftover)
	}
	return nil
}

// decodeValue decodes the rest of the message assuming type tt.
func (d *binaryDecoder) decodeValue(tt *vdl.Type, target vdl.Target) error {
	ttFrom := tt
	if tt.Kind() == vdl.Optional {
		// If the type is optional, we expect to see either WireCtrlNil or the actual
		// value, but not both.  And thus, we can just peek for the WireCtrlNil here.
		switch ctrl, err := binaryPeekControl(d.buf); {
		case err != nil:
			return err
		case ctrl == WireCtrlNil:
			d.buf.Skip(1)
			return target.FromNil(ttFrom)
		}
		tt = tt.Elem()
	}
	if tt.IsBytes() {
		len, err := binaryDecodeLenOrArrayLen(d.buf, tt)
		if err != nil {
			return err
		}
		bytes, err := d.buf.ReadBuf(len)
		if err != nil {
			return err
		}
		return target.FromBytes(bytes, ttFrom)
	}
	switch kind := tt.Kind(); kind {
	case vdl.Bool:
		v, err := binaryDecodeBool(d.buf)
		if err != nil {
			return err
		}
		return target.FromBool(v, ttFrom)
	case vdl.Byte:
		v, err := d.buf.ReadByte()
		if err != nil {
			return err
		}
		return target.FromUint(uint64(v), ttFrom)
	case vdl.Uint16, vdl.Uint32, vdl.Uint64:
		v, err := binaryDecodeUint(d.buf)
		if err != nil {
			return err
		}
		return target.FromUint(v, ttFrom)
	case vdl.Int16, vdl.Int32, vdl.Int64:
		v, err := binaryDecodeInt(d.buf)
		if err != nil {
			return err
		}
		return target.FromInt(v, ttFrom)
	case vdl.Float32, vdl.Float64:
		v, err := binaryDecodeFloat(d.buf)
		if err != nil {
			return err
		}
		return target.FromFloat(v, ttFrom)
	case vdl.Complex64, vdl.Complex128:
		re, err := binaryDecodeFloat(d.buf)
		if err != nil {
			return err
		}
		im, err := binaryDecodeFloat(d.buf)
		if err != nil {
			return err
		}
		return target.FromComplex(complex(re, im), ttFrom)
	case vdl.String:
		v, err := binaryDecodeString(d.buf)
		if err != nil {
			return err
		}
		return target.FromString(v, ttFrom)
	case vdl.Enum:
		index, err := binaryDecodeUint(d.buf)
		switch {
		case err != nil:
			return err
		case index >= uint64(tt.NumEnumLabel()):
			return errIndexOutOfRange
		}
		return target.FromEnumLabel(tt.EnumLabel(int(index)), ttFrom)
	case vdl.TypeObject:
		id, err := binaryDecodeUint(d.buf)
		if err != nil {
			return err
		}
		typeobj, err := d.recvTypes.LookupOrBuildType(typeID(id))
		if err != nil {
			return err
		}
		return target.FromTypeObject(typeobj)
	case vdl.Array, vdl.List:
		len, err := binaryDecodeLenOrArrayLen(d.buf, tt)
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
		len, err := binaryDecodeLen(d.buf)
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
		len, err := binaryDecodeLen(d.buf)
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
			index, ctrl, err := binaryDecodeUintWithControl(d.buf)
			switch {
			case err != nil:
				return err
			case ctrl == WireCtrlEOF:
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
				return fmt.Errorf("vom: unexpected control byte 0x%x", ctrl)
			case index >= uint64(tt.NumField()):
				return errIndexOutOfRange
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
		index, err := binaryDecodeUint(d.buf)
		switch {
		case err != nil:
			return err
		case index >= uint64(tt.NumField()):
			return errIndexOutOfRange
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
		switch id, ctrl, err := binaryDecodeUintWithControl(d.buf); {
		case err != nil:
			return err
		case ctrl == WireCtrlNil:
			return target.FromNil(vdl.AnyType)
		case ctrl != 0:
			return fmt.Errorf("vom: unexpected control byte 0x%x", ctrl)
		default:
			elemType, err := d.recvTypes.LookupOrBuildType(typeID(id))
			if err != nil {
				return err
			}
			return d.decodeValue(elemType, target)
		}
	default:
		panic(fmt.Errorf("vom: decodeValue unhandled type %v", tt))
	}
}

// ignoreValue ignores the rest of the value of type t.  This is used to ignore
// unknown struct fields.
func (d *binaryDecoder) ignoreValue(tt *vdl.Type) error {
	if tt.IsBytes() {
		len, err := binaryDecodeLenOrArrayLen(d.buf, tt)
		if err != nil {
			return err
		}
		return d.buf.Skip(len)
	}
	switch kind := tt.Kind(); kind {
	case vdl.Bool, vdl.Byte:
		return d.buf.Skip(1)
	case vdl.Uint16, vdl.Uint32, vdl.Uint64, vdl.Int16, vdl.Int32, vdl.Int64, vdl.Float32, vdl.Float64, vdl.Enum, vdl.TypeObject:
		// The underlying encoding of all these types is based on uint.
		return binaryIgnoreUint(d.buf)
	case vdl.Complex64, vdl.Complex128:
		// Complex is encoded as two floats, so we can simply ignore two uints.
		if err := binaryIgnoreUint(d.buf); err != nil {
			return err
		}
		return binaryIgnoreUint(d.buf)
	case vdl.String:
		return binaryIgnoreString(d.buf)
	case vdl.Array, vdl.List, vdl.Set, vdl.Map:
		len, err := binaryDecodeLenOrArrayLen(d.buf, tt)
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
			switch index, ctrl, err := binaryDecodeUintWithControl(d.buf); {
			case err != nil:
				return err
			case ctrl == WireCtrlEOF:
				return nil
			case ctrl != 0:
				return fmt.Errorf("vom: unexpected control byte 0x%x", ctrl)
			case index >= uint64(tt.NumField()):
				return errIndexOutOfRange
			default:
				ttfield := tt.Field(int(index))
				if err := d.ignoreValue(ttfield.Type); err != nil {
					return err
				}
			}
		}
	case vdl.Union:
		switch index, err := binaryDecodeUint(d.buf); {
		case err != nil:
			return err
		case index >= uint64(tt.NumField()):
			return errIndexOutOfRange
		default:
			ttfield := tt.Field(int(index))
			return d.ignoreValue(ttfield.Type)
		}
	case vdl.Any:
		switch id, ctrl, err := binaryDecodeUintWithControl(d.buf); {
		case err != nil:
			return err
		case ctrl == WireCtrlNil:
			return nil
		case ctrl != 0:
			return fmt.Errorf("vom: unexpected control byte 0x%x", ctrl)
		default:
			elemType, err := d.recvTypes.LookupOrBuildType(typeID(id))
			if err != nil {
				return err
			}
			return d.ignoreValue(elemType)
		}
	default:
		panic(fmt.Errorf("vom: ignoreValue unhandled type %v", tt))
	}
}
