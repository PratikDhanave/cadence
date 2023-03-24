// Code generated from storage_capability_controller.cdc. DO NOT EDIT.
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

package sema

import (
	"github.com/onflow/cadence/runtime/ast"
	"github.com/onflow/cadence/runtime/common"
)

const StorageCapabilityControllerTypeBorrowTypeFieldName = "borrowType"

var StorageCapabilityControllerTypeBorrowTypeFieldType = MetaType

const StorageCapabilityControllerTypeBorrowTypeFieldDocString = `
The type of the controlled capability, i.e. the T in ` + "`Capability<T>`" + `.
`

const StorageCapabilityControllerTypeCapabilityIDFieldName = "capabilityID"

var StorageCapabilityControllerTypeCapabilityIDFieldType = UInt64Type

const StorageCapabilityControllerTypeCapabilityIDFieldDocString = `
The identifier of the controlled capability.
All copies of a capability have the same ID.
`

const StorageCapabilityControllerTypeDeleteFunctionName = "delete"

var StorageCapabilityControllerTypeDeleteFunctionType = &FunctionType{
	ReturnTypeAnnotation: NewTypeAnnotation(
		VoidType,
	),
}

const StorageCapabilityControllerTypeDeleteFunctionDocString = `
Delete this capability controller,
and disable the controlled capability and its copies.

The controller will be deleted from storage,
but the controlled capability and its copies remain.

Once this function returns, the controller is no longer usable,
all further operations on the controller will panic.

Borrowing from the controlled capability or its copies will return nil.
`

const StorageCapabilityControllerTypeTargetFunctionName = "target"

var StorageCapabilityControllerTypeTargetFunctionType = &FunctionType{
	ReturnTypeAnnotation: NewTypeAnnotation(
		StoragePathType,
	),
}

const StorageCapabilityControllerTypeTargetFunctionDocString = `
Returns the targeted storage path of the controlled capability.
`

const StorageCapabilityControllerTypeRetargetFunctionName = "retarget"

var StorageCapabilityControllerTypeRetargetFunctionType = &FunctionType{
	Parameters: []Parameter{
		{
			Identifier:     "target",
			TypeAnnotation: NewTypeAnnotation(StoragePathType),
		},
	},
	ReturnTypeAnnotation: NewTypeAnnotation(
		VoidType,
	),
}

const StorageCapabilityControllerTypeRetargetFunctionDocString = `
Retarget the capability.
This moves the CapCon from one CapCon array to another.
`

const StorageCapabilityControllerTypeName = "StorageCapabilityController"

var StorageCapabilityControllerType = &SimpleType{
	Name:          StorageCapabilityControllerTypeName,
	QualifiedName: StorageCapabilityControllerTypeName,
	TypeID:        StorageCapabilityControllerTypeName,
	tag:           StorageCapabilityControllerTypeTag,
	IsResource:    false,
	Storable:      false,
	Equatable:     false,
	Exportable:    false,
	Importable:    false,
	Members: func(t *SimpleType) map[string]MemberResolver {
		return map[string]MemberResolver{
			StorageCapabilityControllerTypeBorrowTypeFieldName: {
				Kind: common.DeclarationKindField,
				Resolve: func(memoryGauge common.MemoryGauge,
					identifier string,
					targetRange ast.Range,
					report func(error)) *Member {

					return NewPublicConstantFieldMember(
						memoryGauge,
						t,
						identifier,
						StorageCapabilityControllerTypeBorrowTypeFieldType,
						StorageCapabilityControllerTypeBorrowTypeFieldDocString,
					)
				},
			},
			StorageCapabilityControllerTypeCapabilityIDFieldName: {
				Kind: common.DeclarationKindField,
				Resolve: func(memoryGauge common.MemoryGauge,
					identifier string,
					targetRange ast.Range,
					report func(error)) *Member {

					return NewPublicConstantFieldMember(
						memoryGauge,
						t,
						identifier,
						StorageCapabilityControllerTypeCapabilityIDFieldType,
						StorageCapabilityControllerTypeCapabilityIDFieldDocString,
					)
				},
			},
			StorageCapabilityControllerTypeDeleteFunctionName: {
				Kind: common.DeclarationKindFunction,
				Resolve: func(memoryGauge common.MemoryGauge,
					identifier string,
					targetRange ast.Range,
					report func(error)) *Member {

					return NewPublicFunctionMember(
						memoryGauge,
						t,
						identifier,
						StorageCapabilityControllerTypeDeleteFunctionType,
						StorageCapabilityControllerTypeDeleteFunctionDocString,
					)
				},
			},
			StorageCapabilityControllerTypeTargetFunctionName: {
				Kind: common.DeclarationKindFunction,
				Resolve: func(memoryGauge common.MemoryGauge,
					identifier string,
					targetRange ast.Range,
					report func(error)) *Member {

					return NewPublicFunctionMember(
						memoryGauge,
						t,
						identifier,
						StorageCapabilityControllerTypeTargetFunctionType,
						StorageCapabilityControllerTypeTargetFunctionDocString,
					)
				},
			},
			StorageCapabilityControllerTypeRetargetFunctionName: {
				Kind: common.DeclarationKindFunction,
				Resolve: func(memoryGauge common.MemoryGauge,
					identifier string,
					targetRange ast.Range,
					report func(error)) *Member {

					return NewPublicFunctionMember(
						memoryGauge,
						t,
						identifier,
						StorageCapabilityControllerTypeRetargetFunctionType,
						StorageCapabilityControllerTypeRetargetFunctionDocString,
					)
				},
			},
		}
	},
}
