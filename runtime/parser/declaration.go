/*
 * Cadence - The resource-oriented smart contract programming language
 *
 * Copyright 2019-2022 Dapper Labs, Inc.
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

package parser

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

	"github.com/onflow/cadence/runtime/ast"
	"github.com/onflow/cadence/runtime/common"
	"github.com/onflow/cadence/runtime/errors"
	. "github.com/onflow/cadence/runtime/parser/constants"
	"github.com/onflow/cadence/runtime/parser/lexer"
)

func parseDeclarations(p *parser, endTokenType lexer.TokenType) (declarations []ast.Declaration, err error) {
	for {
		_, docString := p.parseTrivia(triviaOptions{
			skipNewlines:    true,
			parseDocStrings: true,
		})

		switch p.current.Type {
		case lexer.TokenSemicolon:
			// Skip the semicolon
			p.next()
			continue

		case endTokenType, lexer.TokenEOF:
			return

		default:
			var declaration ast.Declaration
			declaration, err = parseDeclaration(p, docString)
			if err != nil {
				return
			}

			if declaration == nil {
				return
			}

			declarations = append(declarations, declaration)
		}
	}
}

func parseDeclaration(p *parser, docString string) (ast.Declaration, error) {

	access := ast.AccessNotSpecified
	var accessPos *ast.Position

	for {
		p.skipSpaceAndComments(true)

		switch p.current.Type {
		case lexer.TokenPragma:
			if access != ast.AccessNotSpecified {
				return nil, p.syntaxError("invalid access modifier for pragma")
			}
			return parsePragmaDeclaration(p)
		case lexer.TokenIdentifier:
			switch p.current.Value {
			case KeywordLet, KeywordVar:
				return parseVariableDeclaration(p, access, accessPos, docString)

			case KeywordFun:
				return parseFunctionDeclaration(p, false, access, accessPos, docString)

			case KeywordImport:
				return parseImportDeclaration(p)

			case KeywordEvent:
				return parseEventDeclaration(p, access, accessPos, docString)

			case KeywordStruct, KeywordResource, KeywordContract, KeywordEnum:
				return parseCompositeOrInterfaceDeclaration(p, access, accessPos, docString)

			case KeywordTransaction:
				if access != ast.AccessNotSpecified {
					return nil, p.syntaxError("invalid access modifier for transaction")
				}
				return parseTransactionDeclaration(p, docString)

			case KeywordPriv, KeywordPub, KeywordAccess:
				if access != ast.AccessNotSpecified {
					return nil, p.syntaxError("invalid second access modifier")
				}
				pos := p.current.StartPos
				accessPos = &pos
				var err error
				access, err = parseAccess(p)
				if err != nil {
					return nil, err
				}

				continue
			}
		}

		return nil, nil
	}
}

// parseAccess parses an access modifier
//
//     access
//         : 'priv'
//         | 'pub' ( '(' 'set' ')' )?
//         | 'access' '(' ( 'self' | 'contract' | 'account' | 'all' ) ')'
//
func parseAccess(p *parser) (ast.Access, error) {

	switch p.current.Value {
	case KeywordPriv:
		// Skip the `priv` Keyword
		p.next()
		return ast.AccessPrivate, nil

	case KeywordPub:
		// Skip the `pub` Keyword
		p.next()
		p.skipSpaceAndComments(true)
		if !p.current.Is(lexer.TokenParenOpen) {
			return ast.AccessPublic, nil
		}

		// Skip the opening paren
		p.next()
		p.skipSpaceAndComments(true)

		if !p.current.Is(lexer.TokenIdentifier) {
			return ast.AccessNotSpecified, p.syntaxError(
				"expected keyword %q, got %s",
				KeywordSet,
				p.current.Type,
			)
		}
		if p.current.Value != KeywordSet {
			return ast.AccessNotSpecified, p.syntaxError(
				"expected keyword %q, got %q",
				KeywordSet,
				p.current.Value,
			)
		}

		// Skip the `set` Keyword
		p.next()
		p.skipSpaceAndComments(true)

		_, err := p.mustOne(lexer.TokenParenClose)
		if err != nil {
			return ast.AccessNotSpecified, err
		}

		return ast.AccessPublicSettable, nil

	case KeywordAccess:
		// Skip the `access` Keyword
		p.next()
		p.skipSpaceAndComments(true)

		_, err := p.mustOne(lexer.TokenParenOpen)
		if err != nil {
			return ast.AccessNotSpecified, err
		}

		p.skipSpaceAndComments(true)

		if !p.current.Is(lexer.TokenIdentifier) {
			return ast.AccessNotSpecified, p.syntaxError(
				"expected keyword %s, got %s",
				common.EnumerateWords(
					[]string{
						strconv.Quote(KeywordAll),
						strconv.Quote(KeywordAccount),
						strconv.Quote(KeywordContract),
						strconv.Quote(KeywordSelf),
					},
					"or",
				),
				p.current.Type,
			)
		}

		var access ast.Access

		switch p.current.Value {
		case KeywordAll:
			access = ast.AccessPublic

		case KeywordAccount:
			access = ast.AccessAccount

		case KeywordContract:
			access = ast.AccessContract

		case KeywordSelf:
			access = ast.AccessPrivate

		default:
			return ast.AccessNotSpecified, p.syntaxError(
				"expected keyword %s, got %q",
				common.EnumerateWords(
					[]string{
						strconv.Quote(KeywordAll),
						strconv.Quote(KeywordAccount),
						strconv.Quote(KeywordContract),
						strconv.Quote(KeywordSelf),
					},
					"or",
				),
				p.current.Value,
			)
		}

		// Skip the Keyword
		p.next()
		p.skipSpaceAndComments(true)

		_, err = p.mustOne(lexer.TokenParenClose)
		if err != nil {
			return ast.AccessNotSpecified, err
		}

		return access, nil

	default:
		return ast.AccessNotSpecified, errors.NewUnreachableError()
	}
}

// parseVariableDeclaration parses a variable declaration.
//
//     variableKind : 'var' | 'let'
//
//     variableDeclaration :
//         variableKind identifier ( ':' typeAnnotation )?
//         transfer expression
//         ( transfer expression )?
//
func parseVariableDeclaration(
	p *parser,
	access ast.Access,
	accessPos *ast.Position,
	docString string,
) (*ast.VariableDeclaration, error) {

	startPos := p.current.StartPos
	if accessPos != nil {
		startPos = *accessPos
	}

	isLet := p.current.Value == KeywordLet

	// Skip the `let` or `var` Keyword
	p.next()

	p.skipSpaceAndComments(true)

	identifier, err := p.nonReservedIdentifier("after start of variable declaration")

	if err != nil {
		return nil, err
	}

	// Skip the identifier
	p.next()
	p.skipSpaceAndComments(true)

	var typeAnnotation *ast.TypeAnnotation

	if p.current.Is(lexer.TokenColon) {
		// Skip the colon
		p.next()
		p.skipSpaceAndComments(true)

		typeAnnotation, err = parseTypeAnnotation(p)
		if err != nil {
			return nil, err
		}
	}

	p.skipSpaceAndComments(true)
	transfer := parseTransfer(p)
	if transfer == nil {
		return nil, p.syntaxError("expected transfer")
	}

	value, err := parseExpression(p, lowestBindingPower)
	if err != nil {
		return nil, err
	}

	p.skipSpaceAndComments(true)

	secondTransfer := parseTransfer(p)
	var secondValue ast.Expression
	if secondTransfer != nil {
		secondValue, err = parseExpression(p, lowestBindingPower)
		if err != nil {
			return nil, err
		}
	}

	variableDeclaration := ast.NewVariableDeclaration(
		p.memoryGauge,
		access,
		isLet,
		identifier,
		typeAnnotation,
		value,
		transfer,
		startPos,
		secondTransfer,
		secondValue,
		docString,
	)

	castingExpression, leftIsCasting := value.(*ast.CastingExpression)
	if leftIsCasting {
		castingExpression.ParentVariableDeclaration = variableDeclaration
	}

	return variableDeclaration, nil
}

// parseTransfer parses a transfer.
//
//     transfer : '=' | '<-' | '<-!'
//
func parseTransfer(p *parser) *ast.Transfer {
	var operation ast.TransferOperation

	switch p.current.Type {
	case lexer.TokenEqual:
		operation = ast.TransferOperationCopy

	case lexer.TokenLeftArrow:
		operation = ast.TransferOperationMove

	case lexer.TokenLeftArrowExclamation:
		operation = ast.TransferOperationMoveForced
	}

	if operation == ast.TransferOperationUnknown {
		return nil
	}

	pos := p.current.StartPos

	p.next()

	return ast.NewTransfer(
		p.memoryGauge,
		operation,
		pos,
	)
}

func parsePragmaDeclaration(p *parser) (*ast.PragmaDeclaration, error) {
	startPos := p.current.StartPosition()
	p.next()

	expr, err := parseExpression(p, lowestBindingPower)
	if err != nil {
		return nil, err
	}

	return ast.NewPragmaDeclaration(
		p.memoryGauge,
		expr,
		ast.NewRange(
			p.memoryGauge,
			startPos,
			expr.EndPosition(p.memoryGauge),
		),
	), nil
}

// parseImportDeclaration parses an import declaration
//
//     importDeclaration :
//         'import'
//         ( identifier (',' identifier)* 'from' )?
//         ( string | hexadecimalLiteral | identifier )
//
func parseImportDeclaration(p *parser) (*ast.ImportDeclaration, error) {

	startPosition := p.current.StartPos

	var identifiers []ast.Identifier

	var location common.Location
	var locationPos ast.Position
	var endPos ast.Position

	parseStringOrAddressLocation := func() {
		locationPos = p.current.StartPos
		endPos = p.current.EndPos

		switch p.current.Type {
		case lexer.TokenString:
			parsedString := parseStringLiteral(p, p.current.Value.(string))
			location = common.NewStringLocation(p.memoryGauge, parsedString)

		case lexer.TokenHexadecimalIntegerLiteral:
			location = parseHexadecimalLocation(p)

		default:
			panic(errors.NewUnreachableError())
		}

		// Skip the location
		p.next()
	}

	setIdentifierLocation := func(identifier ast.Identifier) {
		location = common.IdentifierLocation(identifier.Identifier)
		locationPos = identifier.Pos
		endPos = identifier.EndPosition(p.memoryGauge)
	}

	parseLocation := func() error {
		switch p.current.Type {
		case lexer.TokenString, lexer.TokenHexadecimalIntegerLiteral:
			parseStringOrAddressLocation()

		case lexer.TokenIdentifier:
			identifier := p.tokenToIdentifier(p.current)
			setIdentifierLocation(identifier)
			p.next()

		default:
			return p.syntaxError(
				"unexpected token in import declaration: got %s, expected string, address, or identifier",
				p.current.Type,
			)
		}

		return nil
	}

	parseMoreIdentifiers := func() error {
		expectCommaOrFrom := false

		atEnd := false
		for !atEnd {
			p.next()
			p.skipSpaceAndComments(true)

			switch p.current.Type {
			case lexer.TokenComma:
				if !expectCommaOrFrom {
					return p.syntaxError(
						"expected %s or Keyword %q, got %s",
						lexer.TokenIdentifier,
						KeywordFrom,
						p.current.Type,
					)
				}
				expectCommaOrFrom = false

			case lexer.TokenIdentifier:

				if p.current.Value == KeywordFrom {
					if expectCommaOrFrom {
						atEnd = true

						// Skip the `from` Keyword
						p.next()
						p.skipSpaceAndComments(true)

						err := parseLocation()
						if err != nil {
							return err
						}

						break
					}

					isCommaOrFrom, err := isNextTokenCommaOrFrom(p)
					if err != nil {
						return err
					}

					if !isCommaOrFrom {
						return p.syntaxError(
							"expected %s, got keyword %q",
							lexer.TokenIdentifier,
							p.current.Value,
						)
					}

					// If the next token is either comma or 'from' token, then fall through
					// and process the current 'from' token as an identifier.
				}

				identifier := p.tokenToIdentifier(p.current)
				identifiers = append(identifiers, identifier)

				expectCommaOrFrom = true

			case lexer.TokenEOF:
				return p.syntaxError(
					"unexpected end in import declaration: expected %s or %s",
					lexer.TokenIdentifier,
					lexer.TokenComma,
				)

			default:
				return p.syntaxError(
					"unexpected token in import declaration: got %s, expected keyword %q or %s",
					p.current.Type,
					KeywordFrom,
					lexer.TokenComma,
				)
			}
		}

		return nil
	}

	maybeParseFromIdentifier := func(identifier ast.Identifier) error {
		// The current identifier is maybe the `from` Keyword,
		// in which case the given (previous) identifier was
		// an imported identifier and not the import location.
		//
		// If it is not the `from` Keyword,
		// the given (previous) identifier is the import location.

		if p.current.Value == KeywordFrom {
			identifiers = append(identifiers, identifier)
			// Skip the `from` Keyword
			p.next()
			p.skipSpaceAndComments(true)

			err := parseLocation()
			if err != nil {
				return err
			}
		} else {
			setIdentifierLocation(identifier)
		}

		return nil
	}

	// Skip the `import` Keyword
	p.next()
	p.skipSpaceAndComments(true)

	switch p.current.Type {
	case lexer.TokenString, lexer.TokenHexadecimalIntegerLiteral:
		parseStringOrAddressLocation()

	case lexer.TokenIdentifier:
		identifier := p.tokenToIdentifier(p.current)
		// Skip the identifier
		p.next()
		p.skipSpaceAndComments(true)

		switch p.current.Type {
		case lexer.TokenComma:
			// The previous identifier is an imported identifier,
			// not the import location
			identifiers = append(identifiers, identifier)
			err := parseMoreIdentifiers()
			if err != nil {
				return nil, err
			}
		case lexer.TokenIdentifier:
			err := maybeParseFromIdentifier(identifier)
			if err != nil {
				return nil, err
			}
		case lexer.TokenEOF:
			// The previous identifier is the identifier location
			setIdentifierLocation(identifier)

		default:
			return nil, p.syntaxError(
				"unexpected token in import declaration: got %s, expected keyword %q or %s",
				p.current.Type,
				KeywordFrom,
				lexer.TokenComma,
			)
		}

	case lexer.TokenEOF:
		return nil, p.syntaxError("unexpected end in import declaration: expected string, address, or identifier")

	default:
		return nil, p.syntaxError(
			"unexpected token in import declaration: got %s, expected string, address, or identifier",
			p.current.Type,
		)
	}

	return ast.NewImportDeclaration(
		p.memoryGauge,
		identifiers,
		location,
		ast.NewRange(
			p.memoryGauge,
			startPosition,
			endPos,
		),
		locationPos,
	), nil
}

// isNextTokenCommaOrFrom check whether the token to follow is a comma or a from token.
func isNextTokenCommaOrFrom(p *parser) (b bool, err error) {
	p.startBuffering()
	defer func() {
		err = p.replayBuffered()
	}()

	// skip the current token
	p.next()
	p.skipSpaceAndComments(true)

	// Lookahead the next token
	switch p.current.Type {
	case lexer.TokenIdentifier:
		return p.current.Value == KeywordFrom, nil
	case lexer.TokenComma:
		return true, nil
	default:
		return false, nil
	}
}

func parseHexadecimalLocation(p *parser) common.AddressLocation {
	literal := p.current.Value.(string)

	bytes := []byte(strings.ReplaceAll(literal[2:], "_", ""))

	length := len(bytes)
	if length%2 == 1 {
		bytes = append([]byte{'0'}, bytes...)
		length++
	}

	rawAddress := make([]byte, hex.DecodedLen(length))
	_, err := hex.Decode(rawAddress, bytes)
	if err != nil {
		// unreachable, hex literal should always be valid
		panic(errors.NewUnexpectedErrorFromCause(err))
	}

	address, err := common.BytesToAddress(rawAddress)
	if err != nil {
		// Any returned error is a syntax error. e.g: Address too large error.
		p.reportSyntaxError(err.Error())
	}

	return common.NewAddressLocation(p.memoryGauge, address, "")
}

// parseEventDeclaration parses an event declaration.
//
//     eventDeclaration : 'event' identifier parameterList
//
func parseEventDeclaration(
	p *parser,
	access ast.Access,
	accessPos *ast.Position,
	docString string,
) (*ast.CompositeDeclaration, error) {

	startPos := p.current.StartPos
	if accessPos != nil {
		startPos = *accessPos
	}

	// Skip the `event` Keyword
	p.next()

	p.skipSpaceAndComments(true)

	identifier, err := p.nonReservedIdentifier("after start of event declaration")

	if err != nil {
		return nil, err
	}

	// Skip the identifier
	p.next()

	parameterList, err := parseParameterList(p)
	if err != nil {
		return nil, err
	}

	initializer := ast.NewSpecialFunctionDeclaration(
		p.memoryGauge,
		common.DeclarationKindInitializer,
		ast.NewFunctionDeclaration(
			p.memoryGauge,
			ast.AccessNotSpecified,
			ast.NewEmptyIdentifier(p.memoryGauge, ast.EmptyPosition),
			parameterList,
			nil,
			nil,
			parameterList.StartPos,
			"",
		),
	)

	members := ast.NewMembers(
		p.memoryGauge,
		[]ast.Declaration{
			initializer,
		},
	)

	return ast.NewCompositeDeclaration(
		p.memoryGauge,
		access,
		common.CompositeKindEvent,
		identifier,
		nil,
		members,
		docString,
		ast.NewRange(
			p.memoryGauge,
			startPos,
			parameterList.EndPos,
		),
	), nil
}

// parseCompositeKind parses a composite kind.
//
//     compositeKind : 'struct' | 'resource' | 'contract' | 'enum'
//
func parseCompositeKind(p *parser) common.CompositeKind {

	if p.current.Is(lexer.TokenIdentifier) {
		switch p.current.Value {
		case KeywordStruct:
			return common.CompositeKindStructure

		case KeywordResource:
			return common.CompositeKindResource

		case KeywordContract:
			return common.CompositeKindContract

		case KeywordEnum:
			return common.CompositeKindEnum
		}
	}

	return common.CompositeKindUnknown
}

// parseFieldWithVariableKind parses a field which has a variable kind.
//
//     variableKind : 'var' | 'let'
//
//     field : variableKind identifier ':' typeAnnotation
//
func parseFieldWithVariableKind(
	p *parser,
	access ast.Access,
	accessPos *ast.Position,
	docString string,
) (*ast.FieldDeclaration, error) {

	startPos := p.current.StartPos
	if accessPos != nil {
		startPos = *accessPos
	}

	var variableKind ast.VariableKind
	switch p.current.Value {
	case KeywordLet:
		variableKind = ast.VariableKindConstant

	case KeywordVar:
		variableKind = ast.VariableKindVariable
	}

	// Skip the `let` or `var` Keyword
	p.next()

	p.skipSpaceAndComments(true)
	if !p.current.Is(lexer.TokenIdentifier) {
		return nil, p.syntaxError(
			"expected identifier after start of field declaration, got %s",
			p.current.Type,
		)
	}

	identifier := p.tokenToIdentifier(p.current)
	// Skip the identifier
	p.next()
	p.skipSpaceAndComments(true)

	_, err := p.mustOne(lexer.TokenColon)
	if err != nil {
		return nil, err
	}

	p.skipSpaceAndComments(true)

	typeAnnotation, err := parseTypeAnnotation(p)
	if err != nil {
		return nil, err
	}

	return ast.NewFieldDeclaration(
		p.memoryGauge,
		access,
		variableKind,
		identifier,
		typeAnnotation,
		docString,
		ast.NewRange(
			p.memoryGauge,
			startPos,
			typeAnnotation.EndPosition(p.memoryGauge),
		),
	), nil
}

// parseCompositeOrInterfaceDeclaration parses an event declaration.
//
//     conformances : ':' nominalType ( ',' nominalType )*
//
//     compositeDeclaration : compositeKind identifier conformances?
//                            '{' membersAndNestedDeclarations '}'
//
//     interfaceDeclaration : compositeKind 'interface' identifier conformances?
//                            '{' membersAndNestedDeclarations '}'
//
func parseCompositeOrInterfaceDeclaration(
	p *parser,
	access ast.Access,
	accessPos *ast.Position,
	docString string,
) (ast.Declaration, error) {

	startPos := p.current.StartPos
	if accessPos != nil {
		startPos = *accessPos
	}

	compositeKind := parseCompositeKind(p)

	// Skip the composite kind Keyword
	p.next()

	var isInterface bool
	var identifier ast.Identifier

	for {
		p.skipSpaceAndComments(true)
		if !p.current.Is(lexer.TokenIdentifier) {
			return nil, p.syntaxError(
				"expected %s, got %s",
				lexer.TokenIdentifier,
				p.current.Type,
			)
		}

		wasInterface := isInterface

		if p.current.Value == KeywordInterface {
			isInterface = true
			if wasInterface {
				return nil, p.syntaxError(
					"expected interface name, got keyword %q",
					KeywordInterface,
				)
			}
			// Skip the `interface` Keyword
			p.next()
			continue
		} else {
			ctx := fmt.Sprintf("following %s declaration", compositeKind.Keyword())
			nonReserved, err := p.nonReservedIdentifier(ctx)
			if err != nil {
				return nil, err
			}

			identifier = nonReserved
			// Skip the identifier
			p.next()
			break
		}
	}

	p.skipSpaceAndComments(true)

	var conformances []*ast.NominalType
	var err error

	if p.current.Is(lexer.TokenColon) {
		// Skip the colon
		p.next()

		conformances, _, err = parseNominalTypes(p, lexer.TokenBraceOpen)
		if err != nil {
			return nil, err
		}

		if len(conformances) < 1 {
			return nil, p.syntaxError(
				"expected at least one conformance after %s",
				lexer.TokenColon,
			)
		}
	}

	p.skipSpaceAndComments(true)

	_, err = p.mustOne(lexer.TokenBraceOpen)
	if err != nil {
		return nil, err
	}

	members, err := parseMembersAndNestedDeclarations(p, lexer.TokenBraceClose)
	if err != nil {
		return nil, err
	}

	p.skipSpaceAndComments(true)

	endToken, err := p.mustOne(lexer.TokenBraceClose)
	if err != nil {
		return nil, err
	}

	declarationRange := ast.NewRange(
		p.memoryGauge,
		startPos,
		endToken.EndPos,
	)

	if isInterface {
		// TODO: remove once interface conformances are supported
		if len(conformances) > 0 {
			// TODO: improve
			return nil, p.syntaxError("unexpected conformances")
		}

		return ast.NewInterfaceDeclaration(
			p.memoryGauge,
			access,
			compositeKind,
			identifier,
			members,
			docString,
			declarationRange,
		), nil
	} else {
		return ast.NewCompositeDeclaration(
			p.memoryGauge,
			access,
			compositeKind,
			identifier,
			conformances,
			members,
			docString,
			declarationRange,
		), nil
	}
}

// parseMembersAndNestedDeclarations parses composite or interface members,
// and nested declarations.
//
//     membersAndNestedDeclarations : ( memberOrNestedDeclaration ';'* )*
//
func parseMembersAndNestedDeclarations(p *parser, endTokenType lexer.TokenType) (*ast.Members, error) {

	var declarations []ast.Declaration

	for {
		_, docString := p.parseTrivia(triviaOptions{
			skipNewlines:    true,
			parseDocStrings: true,
		})

		switch p.current.Type {
		case lexer.TokenSemicolon:
			// Skip the semicolon
			p.next()
			continue

		case endTokenType, lexer.TokenEOF:
			return ast.NewMembers(p.memoryGauge, declarations), nil

		default:
			memberOrNestedDeclaration, err := parseMemberOrNestedDeclaration(p, docString)
			if err != nil {
				return nil, err
			}

			if memberOrNestedDeclaration == nil {
				return ast.NewMembers(p.memoryGauge, declarations), nil
			}

			declarations = append(declarations, memberOrNestedDeclaration)
		}
	}
}

// parseMemberOrNestedDeclaration parses a composite or interface member,
// or a declaration nested in it.
//
//     memberOrNestedDeclaration : field
//                               | specialFunctionDeclaration
//                               | functionDeclaration
//                               | interfaceDeclaration
//                               | compositeDeclaration
//                               | eventDeclaration
//                               | enumCase
//
func parseMemberOrNestedDeclaration(p *parser, docString string) (ast.Declaration, error) {

	const functionBlockIsOptional = true

	access := ast.AccessNotSpecified
	var accessPos *ast.Position

	var previousIdentifierToken *lexer.Token

	for {
		p.skipSpaceAndComments(true)

		switch p.current.Type {
		case lexer.TokenIdentifier:
			switch p.current.Value {
			case KeywordLet, KeywordVar:
				return parseFieldWithVariableKind(p, access, accessPos, docString)

			case KeywordCase:
				return parseEnumCase(p, access, accessPos, docString)

			case KeywordFun:
				return parseFunctionDeclaration(p, functionBlockIsOptional, access, accessPos, docString)

			case KeywordEvent:
				return parseEventDeclaration(p, access, accessPos, docString)

			case KeywordStruct, KeywordResource, KeywordContract, KeywordEnum:
				return parseCompositeOrInterfaceDeclaration(p, access, accessPos, docString)

			case KeywordPriv, KeywordPub, KeywordAccess:
				if access != ast.AccessNotSpecified {
					return nil, p.syntaxError("unexpected access modifier")
				}

				pos := p.current.StartPos
				accessPos = &pos
				var err error
				access, err = parseAccess(p)
				if err != nil {
					return nil, err
				}
				continue

			default:
				if previousIdentifierToken != nil {
					return nil, p.syntaxError("unexpected %s", p.current.Type)
				}

				t := p.current
				previousIdentifierToken = &t
				// Skip the identifier
				p.next()
				continue
			}

		case lexer.TokenColon:
			if previousIdentifierToken == nil {
				return nil, p.syntaxError("unexpected %s", p.current.Type)
			}

			identifier := p.tokenToIdentifier(*previousIdentifierToken)
			return parseFieldDeclarationWithoutVariableKind(p, access, accessPos, identifier, docString)

		case lexer.TokenParenOpen:
			if previousIdentifierToken == nil {
				return nil, p.syntaxError("unexpected %s", p.current.Type)
			}

			identifier := p.tokenToIdentifier(*previousIdentifierToken)
			return parseSpecialFunctionDeclaration(p, functionBlockIsOptional, access, accessPos, identifier)
		}

		return nil, nil
	}
}

func parseFieldDeclarationWithoutVariableKind(
	p *parser,
	access ast.Access,
	accessPos *ast.Position,
	identifier ast.Identifier,
	docString string,
) (*ast.FieldDeclaration, error) {

	startPos := identifier.Pos
	if accessPos != nil {
		startPos = *accessPos
	}

	_, err := p.mustOne(lexer.TokenColon)
	if err != nil {
		return nil, err
	}

	p.skipSpaceAndComments(true)

	typeAnnotation, err := parseTypeAnnotation(p)
	if err != nil {
		return nil, err
	}

	return ast.NewFieldDeclaration(
		p.memoryGauge,
		access,
		ast.VariableKindNotSpecified,
		identifier,
		typeAnnotation,
		docString,
		ast.NewRange(
			p.memoryGauge,
			startPos,
			typeAnnotation.EndPosition(p.memoryGauge),
		),
	), nil
}

func parseSpecialFunctionDeclaration(
	p *parser,
	functionBlockIsOptional bool,
	access ast.Access,
	accessPos *ast.Position,
	identifier ast.Identifier,
) (*ast.SpecialFunctionDeclaration, error) {

	startPos := identifier.Pos
	if accessPos != nil {
		startPos = *accessPos
	}

	// TODO: switch to parseFunctionParameterListAndRest once old parser is deprecated:
	//   allow a return type annotation while parsing, but reject later.

	parameterList, err := parseParameterList(p)
	if err != nil {
		return nil, err
	}

	p.skipSpaceAndComments(true)

	var functionBlock *ast.FunctionBlock

	if !functionBlockIsOptional ||
		p.current.Is(lexer.TokenBraceOpen) {

		functionBlock, err = parseFunctionBlock(p)
		if err != nil {
			return nil, err
		}
	}

	declarationKind := common.DeclarationKindUnknown
	switch identifier.Identifier {
	case KeywordInit:
		declarationKind = common.DeclarationKindInitializer

	case KeywordDestroy:
		declarationKind = common.DeclarationKindDestructor

	case KeywordPrepare:
		declarationKind = common.DeclarationKindPrepare
	}

	return ast.NewSpecialFunctionDeclaration(
		p.memoryGauge,
		declarationKind,
		ast.NewFunctionDeclaration(
			p.memoryGauge,
			access,
			identifier,
			parameterList,
			nil,
			functionBlock,
			startPos,
			"",
		),
	), nil
}

// parseEnumCase parses a field which has a variable kind.
//
//     enumCase : 'case' identifier
//
func parseEnumCase(
	p *parser,
	access ast.Access,
	accessPos *ast.Position,
	docString string,
) (*ast.EnumCaseDeclaration, error) {

	startPos := p.current.StartPos
	if accessPos != nil {
		startPos = *accessPos
	}

	// Skip the `enum` Keyword
	p.next()

	p.skipSpaceAndComments(true)
	if !p.current.Is(lexer.TokenIdentifier) {
		return nil, p.syntaxError(
			"expected identifier after start of enum case declaration, got %s",
			p.current.Type,
		)
	}

	identifier := p.tokenToIdentifier(p.current)
	// Skip the identifier
	p.next()

	return ast.NewEnumCaseDeclaration(
		p.memoryGauge,
		access,
		identifier,
		docString,
		startPos,
	), nil
}
