//go:build wasmtime
// +build wasmtime

/*
 * Cadence - The resource-oriented smart contract programming language
 *
 * Copyright Dapper Labs, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package vm

import (
	"fmt"
	"math/big"

	"C"

	"github.com/bytecodealliance/wasmtime-go/v12"

	"github.com/onflow/cadence/runtime/interpreter"
)

type VM interface {
	Invoke(name string, arguments ...interpreter.Value) (interpreter.Value, error)
}

type vm struct {
	instance *wasmtime.Instance
	store    *wasmtime.Store
}

func (m *vm) Invoke(name string, arguments ...interpreter.Value) (interpreter.Value, error) {

	// GetExport attempts to find an export on this instance by 'name'
	// May return `nil` if this instance has no export named `name`

	f := m.instance.GetExport(m.store, name).Func()

	rawArguments := make([]any, len(arguments))
	for i, argument := range arguments {
		rawArguments[i] = argument
	}

	// Call invokes this function with the provided `args`.

	res, err := f.Call(m.store, rawArguments...)
	if err != nil {
		return nil, err
	}

	if res == nil {
		return nil, nil
	}

	return res.(interpreter.Value), nil
}

func NewVM(wasm []byte) (VM, error) {

	inter, err := interpreter.NewInterpreter(nil, nil, &interpreter.Config{})
	if err != nil {
		return nil, err
	}

	// NewConfig creates a new `Config` with all default options configured.

	config := wasmtime.NewConfig()

	// SetWasmReferenceTypes configures whether the wasm reference types proposal is enabled.
	config.SetWasmReferenceTypes(true)

	// NewEngineWithConfig creates a new `Engine` with the `Config` provided
	// Note that once a `Config` is passed to this method it cannot be used again.

	engine := wasmtime.NewEngineWithConfig(config)

	store := wasmtime.NewStore(engine)

	// Module is module which collects
	// definations for types, functions, tables, memories and globals.
	// In addition ,it can declare imports and exports
	// and provide initialization logic
	// in the form of data and element segments or a start function.
	// Module organized WebAssembly programs as the unit of deployment,
	// loading and compilation.

	module, err := wasmtime.NewModule(store.Engine, wasm)
	if err != nil {
		return nil, err
	}

	// WrapFunc wraps a native go function, `f` as a wasm `func`.

	// This function differs from `NewFunc` in that it will determine
	// the type signature of the wasm function given the
	// input value of `f`.
	// The value `f` provided must be a Go function.
	// It may take any number of the following type as arguments :

	// `int32` - a wasm `i32`

	// `int64` a wasm `i64`

	// `float32`

	// `float64`

	// `*Caller`

	//	`*Func`

	// anything else - a wasm `extenref`

	// The go function may return  any number of values.

	// It can return any number of primitive wasm values (integers/floats),
	// and the last return value may optionally be `*Trap` returned is nil
	// then the others values are returned from the wasm function.
	// Otherwise the `*Trap` is returned and
	// it is consider as if the host function traped

	// if the function `f` panics then the panic will be propagated to the caller.

	initfn := func(caller *wasmtime.Caller, offset int32, length int32) (any, *wasmtime.Trap) {
		if offset < 0 {
			return nil, wasmtime.NewTrap(fmt.Sprintf("Int: invalid offset: %d", offset))
		}

		if length < 2 {
			return nil, wasmtime.NewTrap(fmt.Sprintf("Int: invalid length: %d", length))
		}

		mem := caller.GetExport("mem").Memory()

		bytes := C.GoBytes(mem.Data(store), C.int(length))

		value := new(big.Int).SetBytes(bytes[1:])
		if bytes[0] == 0 {
			value = value.Neg(value)
		}

		return interpreter.NewUnmeteredIntValueFromBigInt(value), nil
	}

	intFunc := wasmtime.WrapFunc(
		store,
		initfn,
	)

	stringfn := func(caller *wasmtime.Caller, offset int32, length int32) (any, *wasmtime.Trap) {
		if offset < 0 {
			return nil, wasmtime.NewTrap(fmt.Sprintf("String: invalid offset: %d", offset))
		}

		if length < 0 {
			return nil, wasmtime.NewTrap(fmt.Sprintf("String: invalid length: %d", length))
		}

		mem := caller.GetExport("mem").Memory()

		bytes := C.GoBytes(mem.Data(store), C.int(length))

		return interpreter.NewUnmeteredStringValue(string(bytes)), nil
	}

	stringFunc := wasmtime.WrapFunc(
		store,
		stringstringfn,
	)

	addfn := func(left, right any) (any, *wasmtime.Trap) {
		leftNumber, ok := left.(interpreter.NumberValue)
		if !ok {
			return nil, wasmtime.NewTrap(fmt.Sprintf("add: invalid left: %#+v", left))
		}

		rightNumber, ok := right.(interpreter.NumberValue)
		if !ok {
			return nil, wasmtime.NewTrap(fmt.Sprintf("add: invalid right: %#+v", right))
		}

		return leftNumber.Plus(inter, rightNumber, interpreter.EmptyLocationRange), nil
	}

	addFunc := wasmtime.WrapFunc(
		store,
		addfn,
	)

	// NOTE: wasmtime currently does not support specifying imports by name,
	// unlike other WebAssembly APIs like wasmer, JavaScript, etc.,
	// i.e. imports are imported in the order they are given.

	imports := []wasmtime.AsExtern{
		intFunc,
		stringFunc,
		addFunc,
	}

	instance, err := wasmtime.NewInstance(
		store,
		module,
		imports,
	)
	if err != nil {
		return nil, err
	}

	return &vm{
		instance: instance,
		store:    store,
	}, nil
}
