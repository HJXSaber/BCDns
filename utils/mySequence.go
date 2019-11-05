package utils

import (
	"math/big"
	"reflect"
)

var (
	bigIntType = reflect.TypeOf(new(big.Int))
)

type ConvertError struct {
	Msg string
}

func (err ConvertError) Error() string {
	return err.Msg
}

func Convert(src interface{}, t reflect.Type, top bool) (interface{}, error) {
	obj := reflect.New(t)
	srcV, srcT := reflect.ValueOf(src), reflect.TypeOf(src)
	if t.Kind() != reflect.Struct {
		return nil, ConvertError{Msg: "Can not convert none-struct object"}
	}
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		name := field.Name
		fieldV := obj.Elem().FieldByName(name)
		if f, ok := srcT.FieldByName(name); ok {
			if f.Type == field.Type {
				fieldV.Set(srcV.FieldByName(name))
			} else if f.Type == bigIntType || field.Type == bigIntType {
				if field.Type == bigIntType {
					fieldValue, err := ParseBigInt(srcV.FieldByName(name).Interface().([]byte))
					if err != nil {
						return nil, err
					}
					fieldV.Set(reflect.ValueOf(fieldValue))
				} else {
					encode, err := MakeBigInt(srcV.FieldByName(name).Interface().(*big.Int))
					if err != nil {
						return nil, err
					}
					bigIntBytes := make([]byte, encode.Len())
					encode.Encode(bigIntBytes)
					fieldV.Set(reflect.ValueOf(bigIntBytes))
				}
			} else if f.Type.Kind() == reflect.Ptr {
				if srcV.FieldByName(name).Elem().IsValid() {
					fieldValue, err := Convert(srcV.FieldByName(name).Elem().Interface(), field.Type, false)
					if err != nil {
						return nil, err
					}
					fieldV.Set(reflect.ValueOf(fieldValue).Elem())
				}
			} else if field.Type.Kind() == reflect.Ptr {
				fieldValue, err := Convert(srcV.FieldByName(name).Interface(), field.Type.Elem(), false)
				if err != nil {
					return nil, err
				}
				fieldV.Set(reflect.ValueOf(fieldValue))
			} else if field.Type.Kind() == reflect.Slice {
				eleT := field.Type.Elem()
				fieldValue := reflect.MakeSlice(field.Type, srcV.FieldByName(name).Len(), srcV.FieldByName(name).Len())
				for j := 0; j < srcV.FieldByName(name).Len(); j++ {
					if eleT.Kind() == reflect.Ptr {
						ele, err := Convert(srcV.FieldByName(name).Index(j).Interface(), eleT.Elem(), false)
						if err != nil {
							return nil, err
						}
						fieldValue.Index(j).Set(reflect.ValueOf(ele))
					} else {
						if srcV.FieldByName(name).Index(j).Elem().IsValid() {
							ele, err := Convert(srcV.FieldByName(name).Index(j).Elem().Interface(), eleT, false)
							if err != nil {
								return nil, err
							}
							fieldValue.Index(j).Set(reflect.ValueOf(ele).Elem())
						}
					}
				}
				fieldV.Set(fieldValue)
			} else if field.Type.Kind() == reflect.Map {
				eleT := field.Type.Elem()
				fieldValue := reflect.MakeMap(field.Type)
				keys := srcV.FieldByName(name).MapKeys()
				for _, key := range keys {
					if eleT.Kind() == reflect.Ptr {
						ele, err := Convert(srcV.FieldByName(name).MapIndex(key).Interface(), eleT.Elem(), false)
						if err != nil {
							return nil, err
						}
						fieldValue.MapIndex(key).Set(reflect.ValueOf(ele))
					} else {
						if srcV.FieldByName(name).MapIndex(key).Elem().IsValid() {
							ele, err := Convert(srcV.FieldByName(name).MapIndex(key).Elem().Interface(), eleT, false)
							if err != nil {
								return nil, err
							}
							fieldValue.MapIndex(key).Set(reflect.ValueOf(ele).Elem())
						}
					}
				}
				fieldV.Set(fieldValue)
			} else {
				return nil, ConvertError{Msg: "Unsupported value type"}
			}
		}
	}
	if top {
		return obj.Elem().Interface(), nil
	} else {
		return obj.Interface(), nil
	}
}

//BigInt begin
// A StructuralError suggests that the ASN.1 data is valid, but the Go type
// which is receiving it doesn't match.
type StructuralError struct {
	Msg string
}

func (e StructuralError) Error() string { return "asn1: structure error: " + e.Msg }

// encoder represents an ASN.1 element that is waiting to be marshaled.
type encoder interface {
	// Len returns the number of bytes needed to marshal this element.
	Len() int
	// Encode encodes this element by writing Len() bytes to dst.
	Encode(dst []byte)
}

type byteEncoder byte

func (c byteEncoder) Len() int {
	return 1
}

func (c byteEncoder) Encode(dst []byte) {
	dst[0] = byte(c)
}

type bytesEncoder []byte

func (b bytesEncoder) Len() int {
	return len(b)
}

func (b bytesEncoder) Encode(dst []byte) {
	if copy(dst, b) != len(b) {
		panic("internal error")
	}
}

type multiEncoder []encoder

func (m multiEncoder) Len() int {
	var size int
	for _, e := range m {
		size += e.Len()
	}
	return size
}

func (m multiEncoder) Encode(dst []byte) {
	var off int
	for _, e := range m {
		e.Encode(dst[off:])
		off += e.Len()
	}
}

var (
	byte00Encoder encoder = byteEncoder(0x00)
	byteFFEncoder encoder = byteEncoder(0xff)
)

// checkInteger returns nil if the given bytes are a valid DER-encoded
// INTEGER and an error otherwise.
func checkInteger(bytes []byte) error {
	if len(bytes) == 0 {
		return StructuralError{"empty integer"}
	}
	if len(bytes) == 1 {
		return nil
	}
	if (bytes[0] == 0 && bytes[1]&0x80 == 0) || (bytes[0] == 0xff && bytes[1]&0x80 == 0x80) {
		return StructuralError{"integer not minimally-encoded"}
	}
	return nil
}

var bigOne = big.NewInt(1)

func MakeBigInt(n *big.Int) (encoder, error) {
	if n == nil {
		return nil, StructuralError{"empty integer"}
	}

	if n.Sign() < 0 {
		// A negative number has to be converted to two's-complement
		// form. So we'll invert and subtract 1. If the
		// most-significant-bit isn't set then we'll need to pad the
		// beginning with 0xff in order to keep the number negative.
		nMinus1 := new(big.Int).Neg(n)
		nMinus1.Sub(nMinus1, bigOne)
		bytes := nMinus1.Bytes()
		for i := range bytes {
			bytes[i] ^= 0xff
		}
		if len(bytes) == 0 || bytes[0]&0x80 == 0 {
			return multiEncoder([]encoder{byteFFEncoder, bytesEncoder(bytes)}), nil
		}
		return bytesEncoder(bytes), nil
	} else if n.Sign() == 0 {
		// Zero is written as a single 0 zero rather than no bytes.
		return byte00Encoder, nil
	} else {
		bytes := n.Bytes()
		if len(bytes) > 0 && bytes[0]&0x80 != 0 {
			// We'll have to pad this with 0x00 in order to stop it
			// looking like a negative number.
			return multiEncoder([]encoder{byte00Encoder, bytesEncoder(bytes)}), nil
		}
		return bytesEncoder(bytes), nil
	}
}

// parseBigInt treats the given bytes as a big-endian, signed integer and returns
// the result.
func ParseBigInt(bytes []byte) (*big.Int, error) {
	if err := checkInteger(bytes); err != nil {
		return nil, err
	}
	ret := new(big.Int)
	if len(bytes) > 0 && bytes[0]&0x80 == 0x80 {
		// This is a negative number.
		notBytes := make([]byte, len(bytes))
		for i := range notBytes {
			notBytes[i] = ^bytes[i]
		}
		ret.SetBytes(notBytes)
		ret.Add(ret, bigOne)
		ret.Neg(ret)
		return ret, nil
	}
	ret.SetBytes(bytes)
	return ret, nil
}

//BigInt end
