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

package interpreter

import (
	"fmt"

	"github.com/onflow/cadence/runtime/common"
	"github.com/onflow/cadence/runtime/errors"
	"github.com/onflow/cadence/runtime/sema"
)

// AuthAccount

var authAccountTypeID = sema.AuthAccountType.ID()
var authAccountStaticType StaticType = PrimitiveStaticTypeAuthAccount // unmetered
var authAccountFieldNames = []string{
	sema.AuthAccountTypeAddressFieldName,
	sema.AuthAccountTypeContractsFieldName,
	sema.AuthAccountTypeKeysFieldName,
	sema.AuthAccountTypeInboxFieldName,
	sema.AuthAccountTypeCapabilitiesFieldName,
}

// NewAuthAccountValue constructs an auth account value.
func NewAuthAccountValue(
	gauge common.MemoryGauge,
	address AddressValue,
	accountBalanceGet func() UFix64Value,
	accountAvailableBalanceGet func() UFix64Value,
	storageUsedGet func(interpreter *Interpreter) UInt64Value,
	storageCapacityGet func(interpreter *Interpreter) UInt64Value,
	addPublicKeyFunction FunctionValue,
	removePublicKeyFunction FunctionValue,
	contractsConstructor func() Value,
	keysConstructor func() Value,
	inboxConstructor func() Value,
	capabilitiesConstructor func() Value,
) Value {

	fields := map[string]Value{
		sema.AuthAccountTypeAddressFieldName:            address,
		sema.AuthAccountTypeAddPublicKeyFunctionName:    addPublicKeyFunction,
		sema.AuthAccountTypeRemovePublicKeyFunctionName: removePublicKeyFunction,
	}

	var contracts Value
	var keys Value
	var inbox Value
	var capabilities Value
	var forEachStoredFunction *HostFunctionValue
	var forEachPublicFunction *HostFunctionValue
	var forEachPrivateFunction *HostFunctionValue
	var typeFunction *HostFunctionValue
	var loadFunction *HostFunctionValue
	var copyFunction *HostFunctionValue
	var saveFunction *HostFunctionValue
	var borrowFunction *HostFunctionValue
	var checkFunction *HostFunctionValue
	var linkFunction *HostFunctionValue
	var linkAccountFunction *HostFunctionValue
	var unlinkFunction *HostFunctionValue
	var getLinkTargetFunction *HostFunctionValue
	var getCapabilityFunction *HostFunctionValue

	computeField := func(name string, inter *Interpreter, locationRange LocationRange) Value {
		switch name {
		case sema.AuthAccountTypeContractsFieldName:
			if contracts == nil {
				contracts = contractsConstructor()
			}
			return contracts

		case sema.AuthAccountTypeKeysFieldName:
			if keys == nil {
				keys = keysConstructor()
			}
			return keys

		case sema.AuthAccountTypeInboxFieldName:
			if inbox == nil {
				inbox = inboxConstructor()
			}
			return inbox

		case sema.AuthAccountTypeCapabilitiesFieldName:
			if capabilities == nil {
				capabilities = capabilitiesConstructor()
			}
			return capabilities

		case sema.AuthAccountTypePublicPathsFieldName:
			return inter.publicAccountPaths(address, locationRange)

		case sema.AuthAccountTypePrivatePathsFieldName:
			return inter.privateAccountPaths(address, locationRange)

		case sema.AuthAccountTypeStoragePathsFieldName:
			return inter.storageAccountPaths(address, locationRange)

		case sema.AuthAccountTypeForEachPublicFunctionName:
			if forEachPublicFunction == nil {
				forEachPublicFunction = inter.newStorageIterationFunction(
					sema.AuthAccountTypeForEachPublicFunctionType,
					address,
					common.PathDomainPublic,
					sema.PublicPathType,
				)
			}
			return forEachPublicFunction

		case sema.AuthAccountTypeForEachPrivateFunctionName:
			if forEachPrivateFunction == nil {
				forEachPrivateFunction = inter.newStorageIterationFunction(
					sema.AuthAccountTypeForEachPrivateFunctionType,
					address,
					common.PathDomainPrivate,
					sema.PrivatePathType,
				)
			}
			return forEachPrivateFunction

		case sema.AuthAccountTypeForEachStoredFunctionName:
			if forEachStoredFunction == nil {
				forEachStoredFunction = inter.newStorageIterationFunction(
					sema.AuthAccountTypeForEachStoredFunctionType,
					address,
					common.PathDomainStorage,
					sema.StoragePathType,
				)
			}
			return forEachStoredFunction

		case sema.AuthAccountTypeBalanceFieldName:
			return accountBalanceGet()

		case sema.AuthAccountTypeAvailableBalanceFieldName:
			return accountAvailableBalanceGet()

		case sema.AuthAccountTypeStorageUsedFieldName:
			return storageUsedGet(inter)

		case sema.AuthAccountTypeStorageCapacityFieldName:
			return storageCapacityGet(inter)

		case sema.AuthAccountTypeTypeFunctionName:
			if typeFunction == nil {
				typeFunction = inter.authAccountTypeFunction(address)
			}
			return typeFunction

		case sema.AuthAccountTypeLoadFunctionName:
			if loadFunction == nil {
				loadFunction = inter.authAccountLoadFunction(address)
			}
			return loadFunction

		case sema.AuthAccountTypeCopyFunctionName:
			if copyFunction == nil {
				copyFunction = inter.authAccountCopyFunction(address)
			}
			return copyFunction

		case sema.AuthAccountTypeSaveFunctionName:
			if saveFunction == nil {
				saveFunction = inter.authAccountSaveFunction(address)
			}
			return saveFunction

		case sema.AuthAccountTypeBorrowFunctionName:
			if borrowFunction == nil {
				borrowFunction = inter.authAccountBorrowFunction(address)
			}
			return borrowFunction

		case sema.AuthAccountTypeCheckFunctionName:
			if checkFunction == nil {
				checkFunction = inter.authAccountCheckFunction(address)
			}
			return checkFunction

		case sema.AuthAccountTypeLinkFunctionName:
			if linkFunction == nil {
				linkFunction = inter.authAccountLinkFunction(address)
			}
			return linkFunction

		case sema.AuthAccountTypeLinkAccountFunctionName:
			if linkAccountFunction == nil {
				linkAccountFunction = inter.authAccountLinkAccountFunction(address)
			}
			return linkAccountFunction

		case sema.AuthAccountTypeUnlinkFunctionName:
			if unlinkFunction == nil {
				unlinkFunction = inter.authAccountUnlinkFunction(address)
			}
			return unlinkFunction

		case sema.AuthAccountTypeGetLinkTargetFunctionName:
			if getLinkTargetFunction == nil {
				getLinkTargetFunction = inter.accountGetLinkTargetFunction(
					sema.AuthAccountTypeGetLinkTargetFunctionType,
					address,
				)
			}
			return getLinkTargetFunction

		case sema.AuthAccountTypeGetCapabilityFunctionName:
			if getCapabilityFunction == nil {
				getCapabilityFunction = accountGetCapabilityFunction(
					gauge,
					address,
					sema.CapabilityPathType,
					sema.AuthAccountTypeGetCapabilityFunctionType,
				)
			}
			return getCapabilityFunction
		}

		return nil
	}

	var str string
	stringer := func(memoryGauge common.MemoryGauge, seenReferences SeenReferences) string {
		if str == "" {
			common.UseMemory(memoryGauge, common.AuthAccountValueStringMemoryUsage)
			addressStr := address.MeteredString(memoryGauge, seenReferences)
			str = fmt.Sprintf("AuthAccount(%s)", addressStr)
		}
		return str
	}

	return NewSimpleCompositeValue(
		gauge,
		authAccountTypeID,
		authAccountStaticType,
		authAccountFieldNames,
		fields,
		computeField,
		nil,
		stringer,
	)
}

// PublicAccount

var publicAccountTypeID = sema.PublicAccountType.ID()
var publicAccountStaticType StaticType = PrimitiveStaticTypePublicAccount // unmetered
var publicAccountFieldNames = []string{
	sema.PublicAccountTypeAddressFieldName,
	sema.PublicAccountTypeContractsFieldName,
	sema.PublicAccountTypeKeysFieldName,
	sema.PublicAccountTypeCapabilitiesFieldName,
}

// NewPublicAccountValue constructs a public account value.
func NewPublicAccountValue(
	gauge common.MemoryGauge,
	address AddressValue,
	accountBalanceGet func() UFix64Value,
	accountAvailableBalanceGet func() UFix64Value,
	storageUsedGet func(interpreter *Interpreter) UInt64Value,
	storageCapacityGet func(interpreter *Interpreter) UInt64Value,
	keysConstructor func() Value,
	contractsConstructor func() Value,
	capabilitiesConstructor func() Value,
) Value {

	fields := map[string]Value{
		sema.PublicAccountTypeAddressFieldName: address,
	}

	var keys Value
	var contracts Value
	var capabilities Value
	var forEachPublicFunction *HostFunctionValue
	var getLinkTargetFunction *HostFunctionValue
	var getCapabilityFunction *HostFunctionValue

	computeField := func(name string, inter *Interpreter, locationRange LocationRange) Value {
		switch name {
		case sema.PublicAccountTypeKeysFieldName:
			if keys == nil {
				keys = keysConstructor()
			}
			return keys

		case sema.PublicAccountTypeContractsFieldName:
			if contracts == nil {
				contracts = contractsConstructor()
			}
			return contracts

		case sema.PublicAccountTypeCapabilitiesFieldName:
			if capabilities == nil {
				capabilities = capabilitiesConstructor()
			}
			return capabilities

		case sema.PublicAccountTypePublicPathsFieldName:
			return inter.publicAccountPaths(address, locationRange)

		case sema.PublicAccountTypeForEachPublicFunctionName:
			if forEachPublicFunction == nil {
				forEachPublicFunction = inter.newStorageIterationFunction(
					sema.PublicAccountTypeForEachPublicFunctionType,
					address,
					common.PathDomainPublic,
					sema.PublicPathType,
				)
			}
			return forEachPublicFunction

		case sema.PublicAccountTypeBalanceFieldName:
			return accountBalanceGet()

		case sema.PublicAccountTypeAvailableBalanceFieldName:
			return accountAvailableBalanceGet()

		case sema.PublicAccountTypeStorageUsedFieldName:
			return storageUsedGet(inter)

		case sema.PublicAccountTypeStorageCapacityFieldName:
			return storageCapacityGet(inter)

		case sema.PublicAccountTypeGetLinkTargetFunctionName:
			if getLinkTargetFunction == nil {
				getLinkTargetFunction = inter.accountGetLinkTargetFunction(
					sema.PublicAccountTypeGetLinkTargetFunctionType,
					address,
				)
			}
			return getLinkTargetFunction

		case sema.PublicAccountTypeGetCapabilityFunctionName:
			if getCapabilityFunction == nil {
				getCapabilityFunction = accountGetCapabilityFunction(
					gauge,
					address,
					sema.PublicPathType,
					sema.PublicAccountTypeGetCapabilityFunctionType,
				)
			}
			return getCapabilityFunction
		}

		return nil
	}

	var str string
	stringer := func(memoryGauge common.MemoryGauge, seenReferences SeenReferences) string {
		if str == "" {
			common.UseMemory(memoryGauge, common.PublicAccountValueStringMemoryUsage)
			addressStr := address.MeteredString(memoryGauge, seenReferences)
			str = fmt.Sprintf("PublicAccount(%s)", addressStr)
		}
		return str
	}

	return NewSimpleCompositeValue(
		gauge,
		publicAccountTypeID,
		publicAccountStaticType,
		publicAccountFieldNames,
		fields,
		computeField,
		nil,
		stringer,
	)
}

func accountGetCapabilityFunction(
	gauge common.MemoryGauge,
	addressValue AddressValue,
	pathType sema.Type,
	funcType *sema.FunctionType,
) *HostFunctionValue {

	address := addressValue.ToAddress()

	return NewHostFunctionValue(
		gauge,
		funcType,
		func(invocation Invocation) Value {

			path, ok := invocation.Arguments[0].(PathValue)
			if !ok {
				panic(errors.NewUnreachableError())
			}

			interpreter := invocation.Interpreter

			pathStaticType := path.StaticType(interpreter)

			if !interpreter.IsSubTypeOfSemaType(pathStaticType, pathType) {
				pathSemaType := interpreter.MustConvertStaticToSemaType(pathStaticType)

				panic(TypeMismatchError{
					ExpectedType:  pathType,
					ActualType:    pathSemaType,
					LocationRange: invocation.LocationRange,
				})
			}

			// NOTE: the type parameter is optional, for backwards compatibility

			var borrowType *sema.ReferenceType
			typeParameterPair := invocation.TypeParameterTypes.Oldest()
			if typeParameterPair != nil {
				ty := typeParameterPair.Value
				// we handle the nil case for this below
				borrowType, _ = ty.(*sema.ReferenceType)
			}

			var borrowStaticType StaticType
			if borrowType != nil {
				borrowStaticType = ConvertSemaToStaticType(interpreter, borrowType)
			}

			// Read stored capability, if any

			domain := path.Domain.Identifier()
			identifier := path.Identifier

			storageMapKey := StringStorageMapKey(identifier)

			readValue := interpreter.ReadStored(address, domain, storageMapKey)
			if capabilityValue, ok := readValue.(*IDCapabilityValue); ok {
				// TODO: only if interpreter.IsSubType(capabilityValue.BorrowType, borrowStaticType) ?
				return NewIDCapabilityValue(
					gauge,
					capabilityValue.ID,
					addressValue,
					borrowStaticType,
				)
			}

			return NewPathCapabilityValue(
				gauge,
				addressValue,
				path,
				borrowStaticType,
			)
		},
	)
}
