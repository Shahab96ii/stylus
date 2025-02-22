// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found [here](https://github.com/golang/go/blob/master/LICENSE),
// which applies solely to this file and nothing else.

//go:build js && wasm

package main

import (
	"fmt"
	"go-js-test/syscall"
	"math"
	"runtime"
	"syscall/js"
	"testing"
)

// Object of dummy values (misspelling is intentional and matches the official tests).
var dummys js.Value
var startHash uint64

func init() {
	startHash = syscall.PoolHash()

	// set `dummys` to a new field of the global object
	js.Global().Set("dummys", map[string]interface{}{})
	dummys = js.Global().Get("dummys")

	dummys.Set("someBool", true)
	dummys.Set("someString", "abc\u1234")
	dummys.Set("someInt", 42)
	dummys.Set("someFloat", 42.123)
	dummys.Set("someDate", js.Global().Call("Date"))
	dummys.Set("someArray", []any{41, 42, 43})
	dummys.Set("emptyArray", []any{})
	dummys.Set("emptyObj", map[string]interface{}{})

	dummys.Set("zero", 0)
	dummys.Set("stringZero", "0")
	dummys.Set("NaN", math.NaN())
	dummys.Set("Infinity", math.Inf(1))
	dummys.Set("NegInfinity", math.Inf(-1))

	dummys.Set("add", js.FuncOf(func(this js.Value, args []js.Value) any {
		return args[0].Int() + args[1].Int()
	}))
}

func TestBool(t *testing.T) {
	want := true
	o := dummys.Get("someBool")
	if got := o.Bool(); got != want {
		t.Errorf("got %#v, want %#v", got, want)
	}
	dummys.Set("otherBool", want)
	if got := dummys.Get("otherBool").Bool(); got != want {
		t.Errorf("got %#v, want %#v", got, want)
	}
	if !dummys.Get("someBool").Equal(dummys.Get("someBool")) {
		t.Errorf("same value not equal")
	}
}

func TestString(t *testing.T) {
	want := "abc\u1234"
	o := dummys.Get("someString")
	if got := o.String(); got != want {
		t.Errorf("got %#v, want %#v", got, want)
	}
	dummys.Set("otherString", want)
	if got := dummys.Get("otherString").String(); got != want {
		t.Errorf("got %#v, want %#v", got, want)
	}
	if !dummys.Get("someString").Equal(dummys.Get("someString")) {
		t.Errorf("same value not equal")
	}

	if got, want := js.Undefined().String(), "<undefined>"; got != want {
		t.Errorf("got %#v, want %#v", got, want)
	}
	if got, want := js.Null().String(), "<null>"; got != want {
		t.Errorf("got %#v, want %#v", got, want)
	}
	if got, want := js.ValueOf(true).String(), "<boolean: true>"; got != want {
		t.Errorf("got %#v, want %#v", got, want)
	}
	if got, want := js.ValueOf(42.5).String(), "<number: 42.5>"; got != want {
		t.Errorf("got %#v, want %#v", got, want)
	}
	if got, want := js.Global().String(), "<object>"; got != want {
		t.Errorf("got %#v, want %#v", got, want)
	}
	if got, want := js.Global().Get("Date").String(), "<function>"; got != want {
		t.Errorf("got %#v, want %#v", got, want)
	}
}

func TestInt(t *testing.T) {
	want := 42
	o := dummys.Get("someInt")
	if got := o.Int(); got != want {
		t.Errorf("got %#v, want %#v", got, want)
	}
	dummys.Set("otherInt", want)
	if got := dummys.Get("otherInt").Int(); got != want {
		t.Errorf("got %#v, want %#v", got, want)
	}
	if !dummys.Get("someInt").Equal(dummys.Get("someInt")) {
		t.Errorf("same value not equal")
	}
	if got := dummys.Get("zero").Int(); got != 0 {
		t.Errorf("got %#v, want %#v", got, 0)
	}
}

func TestIntConversion(t *testing.T) {
	testIntConversion(t, 0)
	testIntConversion(t, 1)
	testIntConversion(t, -1)
	testIntConversion(t, 1<<20)
	testIntConversion(t, -1<<20)
	testIntConversion(t, 1<<40)
	testIntConversion(t, -1<<40)
	testIntConversion(t, 1<<60)
	testIntConversion(t, -1<<60)
}

func testIntConversion(t *testing.T, want int) {
	if got := js.ValueOf(want).Int(); got != want {
		t.Errorf("got %#v, want %#v", got, want)
	}
}

func TestFloat(t *testing.T) {
	want := 42.123
	o := dummys.Get("someFloat")
	if got := o.Float(); got != want {
		t.Errorf("got %#v, want %#v", got, want)
	}
	dummys.Set("otherFloat", want)
	if got := dummys.Get("otherFloat").Float(); got != want {
		t.Errorf("got %#v, want %#v", got, want)
	}
	if !dummys.Get("someFloat").Equal(dummys.Get("someFloat")) {
		t.Errorf("same value not equal")
	}
}

func TestObject(t *testing.T) {
	if !dummys.Get("someArray").Equal(dummys.Get("someArray")) {
		t.Errorf("same value not equal")
	}
}

func TestEqual(t *testing.T) {
	if !dummys.Get("someFloat").Equal(dummys.Get("someFloat")) {
		t.Errorf("same float is not equal")
	}
	if !dummys.Get("emptyObj").Equal(dummys.Get("emptyObj")) {
		t.Errorf("same object is not equal")
	}
	if dummys.Get("someFloat").Equal(dummys.Get("someInt")) {
		t.Errorf("different values are not unequal")
	}
}

func TestNaN(t *testing.T) {
	if !dummys.Get("NaN").IsNaN() {
		t.Errorf("JS NaN is not NaN")
	}
	if !js.ValueOf(math.NaN()).IsNaN() {
		t.Errorf("Go NaN is not NaN")
	}
	if dummys.Get("NaN").Equal(dummys.Get("NaN")) {
		t.Errorf("NaN is equal to NaN")
	}
}

func TestUndefined(t *testing.T) {
	if !js.Undefined().IsUndefined() {
		t.Errorf("undefined is not undefined")
	}
	if !js.Undefined().Equal(js.Undefined()) {
		t.Errorf("undefined is not equal to undefined")
	}
	if dummys.IsUndefined() {
		t.Errorf("object is undefined")
	}
	if js.Undefined().IsNull() {
		t.Errorf("undefined is null")
	}
	if dummys.Set("test", js.Undefined()); !dummys.Get("test").IsUndefined() {
		t.Errorf("could not set undefined")
	}
}

func TestNull(t *testing.T) {
	if !js.Null().IsNull() {
		t.Errorf("null is not null")
	}
	if !js.Null().Equal(js.Null()) {
		t.Errorf("null is not equal to null")
	}
	if dummys.IsNull() {
		t.Errorf("object is null")
	}
	if js.Null().IsUndefined() {
		t.Errorf("null is undefined")
	}
	if dummys.Set("test", js.Null()); !dummys.Get("test").IsNull() {
		t.Errorf("could not set null")
	}
	if dummys.Set("test", nil); !dummys.Get("test").IsNull() {
		t.Errorf("could not set nil")
	}
}

func TestLength(t *testing.T) {
	if got := dummys.Get("someArray").Length(); got != 3 {
		t.Errorf("got %#v, want %#v", got, 3)
	}
}

func TestGet(t *testing.T) {
	// positive cases get tested per type

	expectValueError(t, func() {
		dummys.Get("zero").Get("badField")
	})
}

func TestSet(t *testing.T) {
	// positive cases get tested per type

	expectValueError(t, func() {
		dummys.Get("zero").Set("badField", 42)
	})
}

func TestIndex(t *testing.T) {
	if got := dummys.Get("someArray").Index(1).Int(); got != 42 {
		t.Errorf("got %#v, want %#v", got, 42)
	}

	expectValueError(t, func() {
		dummys.Get("zero").Index(1)
	})
}

func TestSetIndex(t *testing.T) {
	dummys.Get("someArray").SetIndex(2, 99)
	if got := dummys.Get("someArray").Index(2).Int(); got != 99 {
		t.Errorf("got %#v, want %#v", got, 99)
	}

	expectValueError(t, func() {
		dummys.Get("zero").SetIndex(2, 99)
	})
}

func TestCall(t *testing.T) {
	var i int64 = 40
	if got := dummys.Call("add", i, 2).Int(); got != 42 {
		t.Errorf("got %#v, want %#v", got, 42)
	}
	if got := dummys.Call("add", 40, 2).Int(); got != 42 {
		t.Errorf("got %#v, want %#v", got, 42)
	}
}

func TestInvoke(t *testing.T) {
	var i int64 = 40
	if got := dummys.Get("add").Invoke(i, 2).Int(); got != 42 {
		t.Errorf("got %#v, want %#v", got, 42)
	}

	expectValueError(t, func() {
		dummys.Get("zero").Invoke()
	})
}

func TestNew(t *testing.T) {
	if got := js.Global().Get("Array").New(42).Length(); got != 42 {
		t.Errorf("got %#v, want %#v", got, 42)
	}
}

func TestType(t *testing.T) {
	if got, want := js.Undefined().Type(), js.TypeUndefined; got != want {
		t.Errorf("got %s, want %s", got, want)
	}
	if got, want := js.Null().Type(), js.TypeNull; got != want {
		t.Errorf("got %s, want %s", got, want)
	}
	if got, want := js.ValueOf(true).Type(), js.TypeBoolean; got != want {
		t.Errorf("got %s, want %s", got, want)
	}
	if got, want := js.ValueOf(0).Type(), js.TypeNumber; got != want {
		t.Errorf("got %s, want %s", got, want)
	}
	if got, want := js.ValueOf(42).Type(), js.TypeNumber; got != want {
		t.Errorf("got %s, want %s", got, want)
	}
	if got, want := js.ValueOf("test").Type(), js.TypeString; got != want {
		t.Errorf("got %s, want %s", got, want)
	}
	if got, want := js.Global().Get("Array").New().Type(), js.TypeObject; got != want {
		t.Errorf("got %s, want %s", got, want)
	}
	if got, want := js.Global().Get("Array").Type(), js.TypeFunction; got != want {
		t.Errorf("got %s, want %s", got, want)
	}
}

type object = map[string]any
type array = []any

func TestValueOf(t *testing.T) {
	a := js.ValueOf(array{0, array{0, 42, 0}, 0})
	if got := a.Index(1).Index(1).Int(); got != 42 {
		t.Errorf("got %v, want %v", got, 42)
	}

	o := js.ValueOf(object{"x": object{"y": 42}})
	if got := o.Get("x").Get("y").Int(); got != 42 {
		t.Errorf("got %v, want %v", got, 42)
	}
}

func TestZeroValue(t *testing.T) {
	var v js.Value
	if !v.IsUndefined() {
		t.Error("zero js.Value is not js.Undefined()")
	}
}

func TestFuncOf(t *testing.T) {
	cb := js.FuncOf(func(this js.Value, args []js.Value) any {
		if got := args[0].Int(); got != 42 {
			t.Errorf("got %#v, want %#v", got, 42)
		}
		return nil
	})
	defer cb.Release()
	cb.Invoke(42)
}

// See
// - https://developer.mozilla.org/en-US/docs/Glossary/Truthy
// - https://stackoverflow.com/questions/19839952/all-falsey-values-in-javascript/19839953#19839953
// - http://www.ecma-international.org/ecma-262/5.1/#sec-9.2
func TestTruthy(t *testing.T) {
	want := true
	for _, key := range []string{
		"someBool", "someString", "someInt", "someFloat", "someArray", "someDate",
		"stringZero", // "0" is truthy
		"add",        // functions are truthy
		"emptyObj", "emptyArray", "Infinity", "NegInfinity",
	} {
		if got := dummys.Get(key).Truthy(); got != want {
			t.Errorf("%s: got %#v, want %#v", key, got, want)
		}
	}

	want = false
	if got := dummys.Get("zero").Truthy(); got != want {
		t.Errorf("got %#v, want %#v", got, want)
	}
	if got := dummys.Get("NaN").Truthy(); got != want {
		t.Errorf("got %#v, want %#v", got, want)
	}
	if got := js.ValueOf("").Truthy(); got != want {
		t.Errorf("got %#v, want %#v", got, want)
	}
	if got := js.Null().Truthy(); got != want {
		t.Errorf("got %#v, want %#v", got, want)
	}
	if got := js.Undefined().Truthy(); got != want {
		t.Errorf("got %#v, want %#v", got, want)
	}
}

func expectValueError(t *testing.T, fn func()) {
	defer func() {
		err := recover()
		if _, ok := err.(*js.ValueError); !ok {
			t.Errorf("expected *js.ValueError, got %T", err)
		}
	}()
	fn()
}

func expectPanic(t *testing.T, fn func()) {
	defer func() {
		err := recover()
		if err == nil {
			t.Errorf("expected panic")
		}
	}()
	fn()
}

var copyTests = []struct {
	srcLen  int
	dstLen  int
	copyLen int
}{
	{5, 3, 3},
	{3, 5, 3},
	{0, 0, 0},
}

func TestCopyBytesToGo(t *testing.T) {
	for _, tt := range copyTests {
		t.Run(fmt.Sprintf("%d-to-%d", tt.srcLen, tt.dstLen), func(t *testing.T) {
			src := js.Global().Get("Uint8Array").New(tt.srcLen)
			if tt.srcLen >= 2 {
				src.SetIndex(1, 42)
			}
			dst := make([]byte, tt.dstLen)

			if got, want := js.CopyBytesToGo(dst, src), tt.copyLen; got != want {
				t.Errorf("copied %d, want %d", got, want)
			}
			if tt.dstLen >= 2 {
				if got, want := int(dst[1]), 42; got != want {
					t.Errorf("got %d, want %d", got, want)
				}
			}
		})
	}
}

func TestCopyBytesToJS(t *testing.T) {
	for _, tt := range copyTests {
		t.Run(fmt.Sprintf("%d-to-%d", tt.srcLen, tt.dstLen), func(t *testing.T) {
			src := make([]byte, tt.srcLen)
			if tt.srcLen >= 2 {
				src[1] = 42
			}
			dst := js.Global().Get("Uint8Array").New(tt.dstLen)

			if got, want := js.CopyBytesToJS(dst, src), tt.copyLen; got != want {
				t.Errorf("copied %d, want %d", got, want)
			}
			if tt.dstLen >= 2 {
				if got, want := dst.Index(1).Int(), 42; got != want {
					t.Errorf("got %d, want %d", got, want)
				}
			}
		})
	}
}

func TestGlobal(t *testing.T) {
	ident := js.FuncOf(func(this js.Value, args []js.Value) any {
		return args[0]
	})
	defer ident.Release()

	if got := ident.Invoke(js.Global()); !got.Equal(js.Global()) {
		t.Errorf("got %#v, want %#v", got, js.Global())
	}
}

func TestPoolHash(t *testing.T) {
	dummys = js.Undefined() // drop dummys
	runtime.GC()

	poolHash := syscall.PoolHash()
	if poolHash != startHash {
		t.Error("reference counting failure:", poolHash, startHash)
	}
}
