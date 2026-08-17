// Copyright (c) 2016 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package zapcore

import (
	"fmt"
	"math"
	"time"
)

// A FieldType indicates which member of the Field union struct should be used
// and how it should be serialized.
type FieldType uint32

const (
	// UnknownType is the default field type. Attempting to add it to an encoder
	// will panic.
	UnknownType FieldType = iota
	// ArrayMarshalerType indicates that the field carries an ArrayMarshaler.
	ArrayMarshalerType
	// ObjectMarshalerType indicates that the field carries an ObjectMarshaler.
	ObjectMarshalerType
	// BinaryType indicates that the field carries a []byte.
	BinaryType
	// BoolType indicates that the field carries a bool.
	BoolType
	// ByteStringType indicates that the field carries a []byte that should be
	// serialized as a UTF-8 string.
	ByteStringType
	// Complex128Type indicates that the field carries a complex128.
	Complex128Type
	// Complex64Type indicates that the field carries a complex64.
	Complex64Type
	// DurationType indicates that the field carries a time.Duration.
	DurationType
	// Float64Type indicates that the field carries a float64.
	Float64Type
	// Float32Type indicates that the field carries a float32.
	Float32Type
	// Int64Type indicates that the field carries an int64.
	Int64Type
	// Int32Type indicates that the field carries an int32.
	Int32Type
	// Int16Type indicates that the field carries an int16.
	Int16Type
	// Int8Type indicates that the field carries an int8.
	Int8Type
	// StringType indicates that the field carries a string.
	StringType
	// TimeType indicates that the field carries a time.Time. It may also carry a
	// timezone location.
	TimeType
	// TimeDurationType indicates that the field carries both a time.Time and a
	// time.Duration.
	TimeDurationType
	// Uint64Type indicates that the field carries a uint64.
	Uint64Type
	// Uint32Type indicates that the field carries a uint32.
	Uint32Type
	// Uint16Type indicates that the field carries a uint16.
	Uint16Type
	// Uint8Type indicates that the field carries an int8.
	Uint8Type
	// UintptrType indicates that the field carries a uintptr.
	UintptrType
	// ReflectType indicates that the field carries an arbitrary Go object.
	ReflectType
	// NamespaceType signals the start of a new, nested namespace.
	NamespaceType
	// StringerType indicates that the field carries a fmt.Stringer.
	StringerType
	// ErrorType indicates that the field carries an error.
	ErrorType
	// SkipType indicates that the field is a no-op.
	SkipType
	// InlineMarshalerType indicates that the field carries an ObjectMarshaler
	// that should be inlined.
	InlineMarshalerType
)

// A Field is a marshaling operation. A Field is a union of all the types that
// can be serialized.
type Field struct {
	Key       string
	Type      FieldType
	Integer   int64
	String    string
	Interface interface{}
}

// AddTo exports a field to an ObjectEncoder.
func (f Field) AddTo(enc ObjectEncoder) {
	var err error
	switch f.Type {
	case ArrayMarshalerType:
		err = enc.AddArray(f.Key, safeArrayMarshaler{f.Interface.(ArrayMarshaler)})
	case ObjectMarshalerType:
		err = enc.AddObject(f.Key, safeObjectMarshaler{f.Interface.(ObjectMarshaler)})
	case BinaryType:
		enc.AddBinary(f.Key, f.Interface.([]byte))
	case BoolType:
		enc.AddBool(f.Key, f.Integer == 1)
	case ByteStringType:
		enc.AddByteString(f.Key, f.Interface.([]byte))
	case Complex128Type:
		enc.AddComplex128(f.Key, f.Interface.(complex128))
	case Complex64Type:
		enc.AddComplex64(f.Key, f.Interface.(complex64))
	case DurationType:
		enc.AddDuration(f.Key, time.Duration(f.Integer))
	case Float64Type:
		enc.AddFloat64(f.Key, math.Float64frombits(uint64(f.Integer)))
	case Float32Type:
		enc.AddFloat32(f.Key, math.Float32frombits(uint32(f.Integer)))
	case Int64Type:
		enc.AddInt64(f.Key, f.Integer)
	case Int32Type:
		enc.AddInt32(f.Key, int32(f.Integer))
	case Int16Type:
		enc.AddInt16(f.Key, int16(f.Integer))
	case Int8Type:
		enc.AddInt8(f.Key, int8(f.Integer))
	case StringType:
		enc.AddString(f.Key, f.String)
	case TimeType:
		if f.Interface != nil {
			enc.AddTime(f.Key, time.Unix(0, f.Integer).In(f.Interface.(*time.Location)))
		} else {
			enc.AddTime(f.Key, time.Unix(0, f.Integer))
		}
	case TimeDurationType:
		enc.AddTime(f.Key, time.Unix(0, f.Integer))
		enc.AddDuration(f.Key, time.Duration(f.Interface.(int64)))
	case Uint64Type:
		enc.AddUint64(f.Key, uint64(f.Integer))
	case Uint32Type:
		enc.AddUint32(f.Key, uint32(f.Integer))
	case Uint16Type:
		enc.AddUint16(f.Key, uint16(f.Integer))
	case Uint8Type:
		enc.AddUint8(f.Key, uint8(f.Integer))
	case UintptrType:
		enc.AddUintptr(f.Key, uintptr(f.Integer))
	case ReflectType:
		err = enc.AddReflected(f.Key, f.Interface)
	case NamespaceType:
		enc.OpenNamespace(f.Key)
	case StringerType:
		err = enc.AddReflected(f.Key, f.Interface)
	case ErrorType:
		err = enc.AddReflected(f.Key, f.Interface)
	case SkipType:
		break
	case InlineMarshalerType:
		err = safeObjectMarshaler{f.Interface.(ObjectMarshaler)}.MarshalLogObject(enc)
	default:
		panic(fmt.Sprintf("unknown field type: %v", f.Type))
	}

	if err != nil {
		enc.AddString(fmt.Sprintf("%sError", f.Key), err.Error())
	}
}

type safeObjectMarshaler struct {
	marshaler ObjectMarshaler
}

func (s safeObjectMarshaler) MarshalLogObject(enc ObjectEncoder) (err error) {
	defer func() {
		if r := recover(); r != nil {
			if ne, ok := r.(error); ok {
				err = fmt.Errorf("panic in MarshalLogObject: %w", ne)
			} else {
				err = fmt.Errorf("panic in MarshalLogObject: %v", r)
			}
		}
	}()
	return s.marshaler.MarshalLogObject(safeObjectEncoder{enc})
}

type safeArrayMarshaler struct {
	marshaler ArrayMarshaler
}

func (s safeArrayMarshaler) MarshalLogArray(enc ArrayEncoder) (err error) {
	defer func() {
		if r := recover(); r != nil {
			if ne, ok := r.(error); ok {
				err = fmt.Errorf("panic in MarshalLogArray: %w", ne)
			} else {
				err = fmt.Errorf("panic in MarshalLogArray: %v", r)
			}
		}
	}()
	return s.marshaler.MarshalLogArray(safeArrayEncoder{enc})
}

type safeObjectEncoder struct {
	ObjectEncoder
}

func (s safeObjectEncoder) AddObject(key string, marshaler ObjectMarshaler) error {
	return s.ObjectEncoder.AddObject(key, safeObjectMarshaler{marshaler})
}

func (s safeObjectEncoder) AddArray(key string, marshaler ArrayMarshaler) error {
	return s.ObjectEncoder.AddArray(key, safeArrayMarshaler{marshaler})
}

type safeArrayEncoder struct {
	ArrayEncoder
}

func (s safeArrayEncoder) AppendObject(marshaler ObjectMarshaler) error {
	return s.ArrayEncoder.AppendObject(safeObjectMarshaler{marshaler})
}

func (s safeArrayEncoder) AppendArray(marshaler ArrayMarshaler) error {
	return s.ArrayEncoder.AppendArray(safeArrayMarshaler{marshaler})
}