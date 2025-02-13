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

type RuntimeTypeConstructor struct {
	Name      string
	Value     *FunctionType
	DocString string
}

const OptionalTypeFunctionName = "OptionalType"

var OptionalTypeFunctionType = &FunctionType{
	Parameters: []Parameter{
		{
			Label:          ArgumentLabelNotRequired,
			Identifier:     "type",
			TypeAnnotation: NewTypeAnnotation(MetaType),
		},
	},
	ReturnTypeAnnotation: NewTypeAnnotation(MetaType),
}

const VariableSizedArrayTypeFunctionName = "VariableSizedArrayType"

var VariableSizedArrayTypeFunctionType = &FunctionType{
	Parameters: []Parameter{
		{
			Label:          ArgumentLabelNotRequired,
			Identifier:     "type",
			TypeAnnotation: NewTypeAnnotation(MetaType),
		},
	},
	ReturnTypeAnnotation: NewTypeAnnotation(MetaType),
}

const ConstantSizedArrayTypeFunctionName = "ConstantSizedArrayType"

var ConstantSizedArrayTypeFunctionType = &FunctionType{
	Parameters: []Parameter{
		{
			Identifier:     "type",
			TypeAnnotation: NewTypeAnnotation(MetaType),
		},
		{
			Identifier:     "size",
			TypeAnnotation: NewTypeAnnotation(IntType),
		},
	},
	ReturnTypeAnnotation: NewTypeAnnotation(MetaType),
}

const DictionaryTypeFunctionName = "DictionaryType"

var DictionaryTypeFunctionType = &FunctionType{
	Parameters: []Parameter{
		{
			Identifier:     "key",
			TypeAnnotation: NewTypeAnnotation(MetaType),
		},
		{
			Identifier:     "value",
			TypeAnnotation: NewTypeAnnotation(MetaType),
		},
	},
	ReturnTypeAnnotation: NewTypeAnnotation(
		&OptionalType{
			Type: MetaType,
		},
	),
}

const CompositeTypeFunctionName = "CompositeType"

var CompositeTypeFunctionType = &FunctionType{
	Parameters: []Parameter{
		{
			Label:          ArgumentLabelNotRequired,
			Identifier:     "identifier",
			TypeAnnotation: NewTypeAnnotation(StringType),
		},
	},
	ReturnTypeAnnotation: NewTypeAnnotation(
		&OptionalType{
			Type: MetaType,
		},
	),
}

const InterfaceTypeFunctionName = "InterfaceType"

var InterfaceTypeFunctionType = &FunctionType{
	Parameters: []Parameter{
		{
			Label:          ArgumentLabelNotRequired,
			Identifier:     "identifier",
			TypeAnnotation: NewTypeAnnotation(StringType),
		},
	},
	ReturnTypeAnnotation: NewTypeAnnotation(
		&OptionalType{
			Type: MetaType,
		},
	),
}

const FunctionTypeFunctionName = "FunctionType"

var FunctionTypeFunctionType = &FunctionType{
	Parameters: []Parameter{
		{
			Identifier: "parameters",
			TypeAnnotation: NewTypeAnnotation(
				&VariableSizedType{
					Type: MetaType,
				},
			),
		},
		{
			Identifier:     "return",
			TypeAnnotation: NewTypeAnnotation(MetaType),
		},
	},
	ReturnTypeAnnotation: NewTypeAnnotation(MetaType),
}

const RestrictedTypeFunctionName = "RestrictedType"

var RestrictedTypeFunctionType = &FunctionType{
	Parameters: []Parameter{
		{
			Identifier: "identifier",
			TypeAnnotation: NewTypeAnnotation(
				&OptionalType{
					Type: StringType,
				},
			),
		},
		{
			Identifier: "restrictions",
			TypeAnnotation: NewTypeAnnotation(
				&VariableSizedType{
					Type: StringType,
				},
			),
		},
	},
	ReturnTypeAnnotation: NewTypeAnnotation(
		&OptionalType{
			Type: MetaType,
		},
	),
}

const ReferenceTypeFunctionName = "ReferenceType"

var ReferenceTypeFunctionType = &FunctionType{
	Parameters: []Parameter{
		{
			Identifier:     "authorized",
			TypeAnnotation: NewTypeAnnotation(BoolType),
		},
		{
			Identifier:     "type",
			TypeAnnotation: NewTypeAnnotation(MetaType),
		},
	},
	ReturnTypeAnnotation: NewTypeAnnotation(MetaType),
}

const CapabilityTypeFunctionName = "CapabilityType"

var CapabilityTypeFunctionType = &FunctionType{
	Parameters: []Parameter{
		{
			Label:          ArgumentLabelNotRequired,
			Identifier:     "type",
			TypeAnnotation: NewTypeAnnotation(MetaType),
		},
	},
	ReturnTypeAnnotation: NewTypeAnnotation(
		&OptionalType{
			Type: MetaType,
		},
	),
}

var runtimeTypeConstructors = []*RuntimeTypeConstructor{
	{
		Name:      OptionalTypeFunctionName,
		Value:     OptionalTypeFunctionType,
		DocString: "Creates a run-time type representing an optional version of the given run-time type.",
	},

	{
		Name:      VariableSizedArrayTypeFunctionName,
		Value:     VariableSizedArrayTypeFunctionType,
		DocString: "Creates a run-time type representing a variable-sized array type of the given run-time type.",
	},

	{
		Name:      ConstantSizedArrayTypeFunctionName,
		Value:     ConstantSizedArrayTypeFunctionType,
		DocString: "Creates a run-time type representing a constant-sized array type of the given run-time type with the specified size.",
	},

	{
		Name:  DictionaryTypeFunctionName,
		Value: DictionaryTypeFunctionType,
		DocString: `Creates a run-time type representing a dictionary type of the given run-time key and value types. 
		Returns nil if the key type is not a valid dictionary key.`,
	},

	{
		Name:  CompositeTypeFunctionName,
		Value: CompositeTypeFunctionType,
		DocString: `Creates a run-time type representing the composite type associated with the given type identifier. 
		Returns nil if the identifier does not correspond to any composite type.`,
	},

	{
		Name:  InterfaceTypeFunctionName,
		Value: InterfaceTypeFunctionType,
		DocString: `Creates a run-time type representing the interface type associated with the given type identifier. 
		Returns nil if the identifier does not correspond to any interface type.`,
	},

	{
		Name:      FunctionTypeFunctionName,
		Value:     FunctionTypeFunctionType,
		DocString: "Creates a run-time type representing a function type associated with the given parameters and return type.",
	},

	{
		Name:      ReferenceTypeFunctionName,
		Value:     ReferenceTypeFunctionType,
		DocString: "Creates a run-time type representing a reference type of the given type, with authorization provided by the first argument.",
	},

	{
		Name:  RestrictedTypeFunctionName,
		Value: RestrictedTypeFunctionType,
		DocString: `Creates a run-time type representing a restricted type of the first argument, restricted by the interface identifiers in the second argument. 
		Returns nil if the restriction is not valid.`,
	},

	{
		Name:      CapabilityTypeFunctionName,
		Value:     CapabilityTypeFunctionType,
		DocString: "Creates a run-time type representing a capability type of the given reference type. Returns nil if the type is not a reference.",
	},
}
