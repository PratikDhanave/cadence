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

package sema

import "github.com/onflow/cadence/runtime/ast"

func (checker *Checker) VisitPragmaDeclaration(declaration *ast.PragmaDeclaration) (_ struct{}) {

	switch expression := declaration.Expression.(type) {
	case *ast.InvocationExpression:
		// Type arguments are not supported for pragmas
		if len(expression.TypeArguments) > 0 {
			checker.report(&InvalidPragmaError{
				Message: "type arguments are not supported",
				Range: ast.NewRangeFromPositioned(
					checker.memoryGauge,
					expression.TypeArguments[0],
				),
			})
		}

		// Ensure arguments are string expressions
		for _, arg := range expression.Arguments {
			_, ok := arg.Expression.(*ast.StringExpression)
			if !ok {
				checker.report(&InvalidPragmaError{
					Message: "invalid non-string argument",
					Range: ast.NewRangeFromPositioned(
						checker.memoryGauge,
						arg.Expression,
					),
				})
			}
		}

	case *ast.IdentifierExpression:
		if IsAllowAccountLinkingPragma(declaration) {
			checker.reportInvalidNonHeaderPragma(declaration)
		}

	default:
		checker.report(&InvalidPragmaError{
			Message: "pragma must be identifier or invocation expression",
			Range: ast.NewRangeFromPositioned(
				checker.memoryGauge,
				declaration,
			),
		})
	}

	return
}

func (checker *Checker) reportInvalidNonHeaderPragma(declaration *ast.PragmaDeclaration) {
	checker.report(&InvalidPragmaError{
		Message: "pragma must appear at top-level, before all other declarations",
		Range: ast.NewRangeFromPositioned(
			checker.memoryGauge,
			declaration,
		),
	})
}

// allowAccountLinkingPragmaIdentifier is the identifier that needs to be used in a pragma to allow account linking.
// This is a temporary feature.
const allowAccountLinkingPragmaIdentifier = "allowAccountLinking"

func IsAllowAccountLinkingPragma(declaration *ast.PragmaDeclaration) bool {
	identifierExpression, ok := declaration.Expression.(*ast.IdentifierExpression)
	if !ok {
		return false
	}

	return identifierExpression.Identifier.Identifier ==
		allowAccountLinkingPragmaIdentifier
}
