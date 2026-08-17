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

package zapcore_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	. "go.uber.org/zap/zapcore"
)

type panickingObject struct{}

func (panickingObject) MarshalLogObject(enc ObjectEncoder) error {
	panic("simulated out-of-bounds or nil pointer panic")
}

type panickingArray struct{}

func (panickingArray) MarshalLogArray(enc ArrayEncoder) error {
	panic("simulated out-of-bounds or nil pointer panic")
}

type nestedPanickingObject struct{}

func (nestedPanickingObject) MarshalLogObject(enc ObjectEncoder) error {
	enc.AddString("foo", "bar")
	return enc.AddObject("nested", panickingObject{})
}

type nestedPanickingArray struct{}

func (nestedPanickingArray) MarshalLogArray(enc ArrayEncoder) error {
	enc.AppendString("foo")
	return enc.AppendArray(panickingArray{})
}

func TestPanickingMarshalers(t *testing.T) {
	for _, tm := range []struct {
		name  string
		field Field
		want  string
	}{
		{
			name:  "object",
			field: Field{Key: "obj", Type: ObjectMarshalerType, Interface: panickingObject{}},
			want:  `"obj":{},"objError":"panic in MarshalLogObject: simulated out-of-bounds or nil pointer panic"`,
		},
		{
			name:  "array",
			field: Field{Key: "arr", Type: ArrayMarshalerType, Interface: panickingArray{}},
			want:  `"arr":[],"arrError":"panic in MarshalLogArray: simulated out-of-bounds or nil pointer panic"`,
		},
		{
			name:  "nested object",
			field: Field{Key: "obj", Type: ObjectMarshalerType, Interface: nestedPanickingObject{}},
			want:  `"obj":{"foo":"bar","nested":{}},"objError":"panic in MarshalLogObject: simulated out-of-bounds or nil pointer panic"`,
		},
		{
			name:  "nested array",
			field: Field{Key: "arr", Type: ArrayMarshalerType, Interface: nestedPanickingArray{}},
			want:  `"arr":["foo",[]],"arrError":"panic in MarshalLogArray: simulated out-of-bounds or nil pointer panic"`,
		},
	} {
		t.Run(tm.name, func(t *testing.T) {
			enc := NewJSONEncoder(EncoderConfig{
				LineEnding: DefaultLineEnding,
			})
			tm.field.AddTo(enc)
			buf, err := enc.EncodeEntry(Entry{}, nil)
			require.NoError(t, err)
			assert.Contains(t, buf.String(), tm.want)
		})
	} 
}

func TestMemoryEncoderPanickingMarshalers(t *testing.T) {
	m := &MapObjectEncoder{Fields: make(map[string]interface{})}
	err := m.AddObject("obj", panickingObject{})
	assert.Error(t, err, "Expected error from panicking object marshaler")
	assert.Contains(t, err.Error(), "panic in MarshalLogObject: simulated out-of-bounds or nil pointer panic")

	err = m.AddArray("arr", panickingArray{})
	assert.Error(t, err, "Expected error from panicking array marshaler")
	assert.Contains(t, err.Error(), "panic in MarshalLogArray: simulated out-of-bounds or nil pointer panic")
}