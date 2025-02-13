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

package checker

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/onflow/cadence/runtime/sema"
)

func TestCheckPragmaInvalidExpr(t *testing.T) {

	t.Parallel()

	_, err := ParseAndCheck(t, `
      #"string"
    `)

	errs := RequireCheckerErrors(t, err, 1)
	assert.IsType(t, &sema.InvalidPragmaError{}, errs[0])
}

func TestCheckPragmaValidIdentifierExpr(t *testing.T) {

	t.Parallel()
	_, err := ParseAndCheck(t, `
        #pedantic
    `)

	require.NoError(t, err)
}

func TestCheckPragmaValidInvocationExpr(t *testing.T) {

	t.Parallel()

	_, err := ParseAndCheck(t, `
        #version("1.0")
    `)

	require.NoError(t, err)
}

func TestCheckPragmaInvalidLocation(t *testing.T) {

	t.Parallel()

	_, err := ParseAndCheck(t, `
      fun test() {
          #version
      }
    `)

	errs := RequireCheckerErrors(t, err, 1)
	assert.IsType(t, &sema.InvalidDeclarationError{}, errs[0])
}

func TestCheckPragmaInvalidInvocationExprNonStringExprArgument(t *testing.T) {

	t.Parallel()

	_, err := ParseAndCheck(t, `
      #version(y)
    `)

	errs := RequireCheckerErrors(t, err, 1)
	assert.IsType(t, &sema.InvalidPragmaError{}, errs[0])
}

func TestCheckPragmaInvalidInvocationExprTypeArgs(t *testing.T) {

	t.Parallel()

	_, err := ParseAndCheck(t, `
      #version<X>()
    `)

	errs := RequireCheckerErrors(t, err, 1)
	assert.IsType(t, &sema.InvalidPragmaError{}, errs[0])
}

func TestCheckAllowAccountLinkingPragma(t *testing.T) {

	t.Parallel()

	t.Run("top-level, before other declarations", func(t *testing.T) {
		t.Parallel()

		_, err := ParseAndCheck(t, `
          #allowAccountLinking

          let x = 1
        `)
		require.NoError(t, err)
	})

	t.Run("top-level, after other pragmas", func(t *testing.T) {
		t.Parallel()

		_, err := ParseAndCheck(t, `
          #someOtherPragma
          #allowAccountLinking

          let x = 1
        `)
		errs := RequireCheckerErrors(t, err, 1)
		assert.IsType(t, &sema.InvalidPragmaError{}, errs[0])
	})

	t.Run("top-level, after other declarations", func(t *testing.T) {
		t.Parallel()

		_, err := ParseAndCheck(t, `
          let x = 1

          #allowAccountLinking
        `)

		errs := RequireCheckerErrors(t, err, 1)
		assert.IsType(t, &sema.InvalidPragmaError{}, errs[0])
	})

	t.Run("nested", func(t *testing.T) {
		t.Parallel()

		_, err := ParseAndCheck(t, `
          fun test() {
              #allowAccountLinking
          }
        `)

		errs := RequireCheckerErrors(t, err, 1)
		assert.IsType(t, &sema.InvalidDeclarationError{}, errs[0])
	})
}
