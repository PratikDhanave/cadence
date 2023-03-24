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

package runtime

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/onflow/cadence"
	"github.com/onflow/cadence/runtime/common"
	"github.com/onflow/cadence/runtime/parser"
	"github.com/onflow/cadence/runtime/stdlib"
)

func TestNewLocationCoverage(t *testing.T) {

	t.Parallel()

	// Represents line numbers with statement execution count.
	// For the time being, if a line has two statements, we cannot
	// distinguish between their hits separately.
	// For example: "if let index = self.index(s, until, startIndex) {"
	lineHits := map[int]int{3: 0, 4: 0, 5: 0, 7: 0, 9: 0, 11: 0}
	locationCoverage := NewLocationCoverage(lineHits)

	assert.Equal(
		t,
		map[int]int{3: 0, 4: 0, 5: 0, 7: 0, 9: 0, 11: 0},
		locationCoverage.LineHits,
	)
	assert.Equal(
		t,
		[]int{3, 4, 5, 7, 9, 11},
		locationCoverage.MissedLines(),
	)
	assert.Equal(t, 6, locationCoverage.Statements)
	assert.Equal(t, "0.0%", locationCoverage.Percentage())
	assert.Equal(t, 0, locationCoverage.CoveredLines())
}

func TestLocationCoverageAddLineHit(t *testing.T) {

	t.Parallel()

	lineHits := map[int]int{3: 0, 4: 0, 5: 0, 7: 0, 9: 0, 11: 0}
	locationCoverage := NewLocationCoverage(lineHits)

	// Lines below 1 are dropped.
	locationCoverage.AddLineHit(0)
	locationCoverage.AddLineHit(3)
	locationCoverage.AddLineHit(3)
	locationCoverage.AddLineHit(7)
	locationCoverage.AddLineHit(9)
	// Line 15 was not included in the lineHits map, however we
	// want it to be tracked. This will help to find out about
	// cases where the inspector does not find all the statements.
	// We should also discuss if the Statements counter should be
	// increased in this case.
	// TBD
	locationCoverage.AddLineHit(15)

	assert.Equal(
		t,
		map[int]int{3: 2, 4: 0, 5: 0, 7: 1, 9: 1, 11: 0, 15: 1},
		locationCoverage.LineHits,
	)
	assert.Equal(t, 6, locationCoverage.Statements)
	assert.Equal(t, "66.7%", locationCoverage.Percentage())
}

func TestLocationCoverageCoveredLines(t *testing.T) {

	t.Parallel()

	lineHits := map[int]int{3: 0, 4: 0, 5: 0, 7: 0, 9: 0, 11: 0}
	locationCoverage := NewLocationCoverage(lineHits)

	locationCoverage.AddLineHit(3)
	locationCoverage.AddLineHit(3)
	locationCoverage.AddLineHit(7)
	locationCoverage.AddLineHit(9)
	locationCoverage.AddLineHit(15)

	assert.Equal(t, 4, locationCoverage.CoveredLines())
}

func TestLocationCoverageMissedLines(t *testing.T) {
	t.Parallel()

	lineHits := map[int]int{3: 0, 4: 0, 5: 0, 7: 0, 9: 0, 11: 0}
	locationCoverage := NewLocationCoverage(lineHits)

	locationCoverage.AddLineHit(3)
	locationCoverage.AddLineHit(3)
	locationCoverage.AddLineHit(7)
	locationCoverage.AddLineHit(9)
	locationCoverage.AddLineHit(15)

	assert.Equal(
		t,
		[]int{4, 5, 11},
		locationCoverage.MissedLines(),
	)
}

func TestLocationCoveragePercentage(t *testing.T) {

	t.Parallel()

	lineHits := map[int]int{3: 0, 4: 0, 5: 0}
	locationCoverage := NewLocationCoverage(lineHits)

	locationCoverage.AddLineHit(3)
	locationCoverage.AddLineHit(4)
	locationCoverage.AddLineHit(5)
	// Note: Line 15 was not included in the lineHits map,
	// but we saturate the percentage at 100%.
	locationCoverage.AddLineHit(15)

	assert.Equal(t, "100.0%", locationCoverage.Percentage())
}

func TestNewCoverageReport(t *testing.T) {

	t.Parallel()

	coverageReport := NewCoverageReport()

	assert.Equal(t, 0, len(coverageReport.Coverage))
	assert.Equal(t, 0, len(coverageReport.Programs))
	assert.Equal(t, 0, len(coverageReport.ExcludedLocations))
}

func TestCoverageReportExcludeLocation(t *testing.T) {

	t.Parallel()

	coverageReport := NewCoverageReport()

	location := common.StringLocation("FooContract")
	coverageReport.ExcludeLocation(location)
	// We do not allow duplicate locations
	coverageReport.ExcludeLocation(location)

	assert.Equal(t, 1, len(coverageReport.ExcludedLocations))
	assert.Equal(t, true, coverageReport.IsLocationExcluded(location))
}

func TestCoverageReportInspectProgram(t *testing.T) {

	t.Parallel()

	script := []byte(`
	  pub fun answer(): Int {
	    var i = 0
	    while i < 42 {
	      i = i + 1
	    }
	    return i
	  }
	`)

	program, err := parser.ParseProgram(nil, script, parser.Config{})
	require.NoError(t, err)

	coverageReport := NewCoverageReport()

	location := common.StringLocation("AnswerScript")
	coverageReport.InspectProgram(location, program)

	assert.Equal(t, 1, len(coverageReport.Coverage))
	assert.Equal(t, 1, len(coverageReport.Programs))
	assert.Equal(t, true, coverageReport.IsProgramInspected(location))
}

func TestCoverageReportInspectProgramForExcludedLocation(t *testing.T) {

	t.Parallel()

	script := []byte(`
	  pub fun answer(): Int {
	    var i = 0
	    while i < 42 {
	      i = i + 1
	    }
	    return i
	  }
	`)

	program, err := parser.ParseProgram(nil, script, parser.Config{})
	require.NoError(t, err)

	coverageReport := NewCoverageReport()

	location := common.StringLocation("AnswerScript")
	coverageReport.ExcludeLocation(location)
	coverageReport.InspectProgram(location, program)

	assert.Equal(t, 0, len(coverageReport.Coverage))
	assert.Equal(t, 0, len(coverageReport.Programs))
	assert.Equal(t, false, coverageReport.IsProgramInspected(location))
}

func TestCoverageReportAddLineHit(t *testing.T) {

	t.Parallel()

	script := []byte(`
	  pub fun answer(): Int {
	    var i = 0
	    while i < 42 {
	      i = i + 1
	    }
	    return i
	  }
	`)

	program, err := parser.ParseProgram(nil, script, parser.Config{})
	require.NoError(t, err)

	coverageReport := NewCoverageReport()

	location := common.StringLocation("AnswerScript")
	coverageReport.InspectProgram(location, program)

	coverageReport.AddLineHit(location, 3)
	coverageReport.AddLineHit(location, 3)
	coverageReport.AddLineHit(location, 5)

	locationCoverage := coverageReport.Coverage[location]

	assert.Equal(
		t,
		map[int]int{3: 2, 4: 0, 5: 1, 7: 0},
		locationCoverage.LineHits,
	)
	assert.Equal(
		t,
		[]int{4, 7},
		locationCoverage.MissedLines(),
	)
	assert.Equal(t, 4, locationCoverage.Statements)
	assert.Equal(t, "50.0%", locationCoverage.Percentage())
	assert.Equal(t, 2, locationCoverage.CoveredLines())
}

func TestCoverageReportWithFlowLocation(t *testing.T) {

	t.Parallel()

	script := []byte(`
	  pub fun answer(): Int {
	    var i = 0
	    while i < 42 {
	      i = i + 1
	    }
	    return i
	  }
	`)

	program, err := parser.ParseProgram(nil, script, parser.Config{})
	require.NoError(t, err)

	coverageReport := NewCoverageReport()

	location := stdlib.FlowLocation{}
	coverageReport.InspectProgram(location, program)

	actual, err := json.Marshal(coverageReport)
	require.NoError(t, err)

	expected := `
	  {
	    "coverage": {
	      "flow": {
	        "line_hits": {
	          "3": 0,
	          "4": 0,
	          "5": 0,
	          "7": 0
	        },
	        "missed_lines": [3, 4, 5, 7],
	        "statements": 4,
	        "percentage": "0.0%"
	      }
	    }
	  }
	`
	require.JSONEq(t, expected, string(actual))
}

func TestCoverageReportWithREPLLocation(t *testing.T) {

	t.Parallel()

	script := []byte(`
	  pub fun answer(): Int {
	    var i = 0
	    while i < 42 {
	      i = i + 1
	    }
	    return i
	  }
	`)

	program, err := parser.ParseProgram(nil, script, parser.Config{})
	require.NoError(t, err)

	coverageReport := NewCoverageReport()

	location := common.REPLLocation{}
	coverageReport.InspectProgram(location, program)

	actual, err := json.Marshal(coverageReport)
	require.NoError(t, err)

	expected := `
	  {
	    "coverage": {
	      "REPL": {
	        "line_hits": {
	          "3": 0,
	          "4": 0,
	          "5": 0,
	          "7": 0
	        },
	        "missed_lines": [3, 4, 5, 7],
	        "statements": 4,
	        "percentage": "0.0%"
	      }
	    }
	  }
	`
	require.JSONEq(t, expected, string(actual))
}

func TestCoverageReportWithScriptLocation(t *testing.T) {

	t.Parallel()

	script := []byte(`
	  pub fun answer(): Int {
	    var i = 0
	    while i < 42 {
	      i = i + 1
	    }
	    return i
	  }
	`)

	program, err := parser.ParseProgram(nil, script, parser.Config{})
	require.NoError(t, err)

	coverageReport := NewCoverageReport()

	location := common.ScriptLocation{0x1, 0x2}
	coverageReport.InspectProgram(location, program)

	actual, err := json.Marshal(coverageReport)
	require.NoError(t, err)

	expected := `
	  {
	    "coverage": {
	      "s.0102000000000000000000000000000000000000000000000000000000000000": {
	        "line_hits": {
	          "3": 0,
	          "4": 0,
	          "5": 0,
	          "7": 0
	        },
	        "missed_lines": [3, 4, 5, 7],
	        "statements": 4,
	        "percentage": "0.0%"
	      }
	    }
	  }
	`
	require.JSONEq(t, expected, string(actual))
}

func TestCoverageReportWithStringLocation(t *testing.T) {

	t.Parallel()

	script := []byte(`
	  pub fun answer(): Int {
	    var i = 0
	    while i < 42 {
	      i = i + 1
	    }
	    return i
	  }
	`)

	program, err := parser.ParseProgram(nil, script, parser.Config{})
	require.NoError(t, err)

	coverageReport := NewCoverageReport()

	location := common.StringLocation("AnswerScript")
	coverageReport.InspectProgram(location, program)

	actual, err := json.Marshal(coverageReport)
	require.NoError(t, err)

	expected := `
	  {
	    "coverage": {
	      "S.AnswerScript": {
	        "line_hits": {
	          "3": 0,
	          "4": 0,
	          "5": 0,
	          "7": 0
	        },
	        "missed_lines": [3, 4, 5, 7],
	        "statements": 4,
	        "percentage": "0.0%"
	      }
	    }
	  }
	`
	require.JSONEq(t, expected, string(actual))
}

func TestCoverageReportWithIdentifierLocation(t *testing.T) {

	t.Parallel()

	script := []byte(`
	  pub fun answer(): Int {
	    var i = 0
	    while i < 42 {
	      i = i + 1
	    }
	    return i
	  }
	`)

	program, err := parser.ParseProgram(nil, script, parser.Config{})
	require.NoError(t, err)

	coverageReport := NewCoverageReport()

	location := common.IdentifierLocation("Answer")
	coverageReport.InspectProgram(location, program)

	actual, err := json.Marshal(coverageReport)
	require.NoError(t, err)

	expected := `
	  {
	    "coverage": {
	      "I.Answer": {
	        "line_hits": {
	          "3": 0,
	          "4": 0,
	          "5": 0,
	          "7": 0
	        },
	        "missed_lines": [3, 4, 5, 7],
	        "statements": 4,
	        "percentage": "0.0%"
	      }
	    }
	  }
	`
	require.JSONEq(t, expected, string(actual))
}

func TestCoverageReportWithTransactionLocation(t *testing.T) {

	t.Parallel()

	script := []byte(`
	  pub fun answer(): Int {
	    var i = 0
	    while i < 42 {
	      i = i + 1
	    }
	    return i
	  }
	`)

	program, err := parser.ParseProgram(nil, script, parser.Config{})
	require.NoError(t, err)

	coverageReport := NewCoverageReport()

	location := common.TransactionLocation{0x1, 0x2}
	coverageReport.InspectProgram(location, program)

	actual, err := json.Marshal(coverageReport)
	require.NoError(t, err)

	expected := `
	  {
	    "coverage": {
	      "t.0102000000000000000000000000000000000000000000000000000000000000": {
	        "line_hits": {
	          "3": 0,
	          "4": 0,
	          "5": 0,
	          "7": 0
	        },
	        "missed_lines": [3, 4, 5, 7],
	        "statements": 4,
	        "percentage": "0.0%"
	      }
	    }
	  }
	`
	require.JSONEq(t, expected, string(actual))
}

func TestCoverageReportWithAddressLocation(t *testing.T) {

	t.Parallel()

	script := []byte(`
	  pub fun answer(): Int {
	    var i = 0
	    while i < 42 {
	      i = i + 1
	    }
	    return i
	  }
	`)

	program, err := parser.ParseProgram(nil, script, parser.Config{})
	require.NoError(t, err)

	coverageReport := NewCoverageReport()

	location := common.AddressLocation{
		Address: common.MustBytesToAddress([]byte{1, 2}),
		Name:    "Answer",
	}
	coverageReport.InspectProgram(location, program)

	actual, err := json.Marshal(coverageReport)
	require.NoError(t, err)

	expected := `
	  {
	    "coverage": {
	      "A.0000000000000102.Answer": {
	        "line_hits": {
	          "3": 0,
	          "4": 0,
	          "5": 0,
	          "7": 0
	        },
	        "missed_lines": [3, 4, 5, 7],
	        "statements": 4,
	        "percentage": "0.0%"
	      }
	    }
	  }
	`
	require.JSONEq(t, expected, string(actual))

}
func TestCoverageReportReset(t *testing.T) {

	t.Parallel()

	script := []byte(`
	  pub fun answer(): Int {
	    var i = 0
	    while i < 42 {
	      i = i + 1
	    }
	    return i
	  }
	`)

	program, err := parser.ParseProgram(nil, script, parser.Config{})
	require.NoError(t, err)

	coverageReport := NewCoverageReport()

	location := common.StringLocation("AnswerScript")
	coverageReport.InspectProgram(location, program)
	coverageReport.AddLineHit(location, 3)
	coverageReport.AddLineHit(location, 3)
	coverageReport.AddLineHit(location, 5)

	excludedLocation := common.StringLocation("XLocation")
	coverageReport.ExcludeLocation(excludedLocation)

	assert.Equal(t, 1, len(coverageReport.Coverage))
	assert.Equal(t, 1, len(coverageReport.Programs))
	assert.Equal(t, 1, len(coverageReport.ExcludedLocations))
	assert.Equal(t, true, coverageReport.IsProgramInspected(location))
	assert.Equal(t, true, coverageReport.IsLocationExcluded(excludedLocation))

	coverageReport.Reset()

	assert.Equal(t, 0, len(coverageReport.Coverage))
	assert.Equal(t, 0, len(coverageReport.Programs))
	assert.Equal(t, 1, len(coverageReport.ExcludedLocations))
	assert.Equal(t, false, coverageReport.IsProgramInspected(location))
	assert.Equal(t, true, coverageReport.IsLocationExcluded(excludedLocation))
}

func TestCoverageReportAddLineHitForExcludedLocation(t *testing.T) {

	t.Parallel()

	coverageReport := NewCoverageReport()

	location := common.StringLocation("AnswerScript")
	coverageReport.ExcludeLocation(location)

	coverageReport.AddLineHit(location, 3)
	coverageReport.AddLineHit(location, 5)

	assert.Equal(t, 0, len(coverageReport.Coverage))
	assert.Equal(t, 0, len(coverageReport.Programs))
	assert.Equal(t, false, coverageReport.IsProgramInspected(location))
}

func TestCoverageReportAddLineHitForNonInspectedProgram(t *testing.T) {

	t.Parallel()

	coverageReport := NewCoverageReport()

	location := common.StringLocation("AnswerScript")

	coverageReport.AddLineHit(location, 3)
	coverageReport.AddLineHit(location, 5)

	assert.Equal(t, 0, len(coverageReport.Coverage))
	assert.Equal(t, 0, len(coverageReport.Programs))
	assert.Equal(t, false, coverageReport.IsProgramInspected(location))
}

func TestRuntimeCoverage(t *testing.T) {

	t.Parallel()

	runtime := NewInterpreterRuntime(Config{
		CoverageReportingEnabled: true,
	})

	importedScript := []byte(`
	  pub let specialNumbers: {Int: String} = {
	    1729: "Harshad",
	    8128: "Harmonic",
	    41041: "Carmichael"
	  }

	  pub fun addSpecialNumber(_ n: Int, _ trait: String) {
	    specialNumbers[n] = trait
	  }

	  pub fun getIntegerTrait(_ n: Int): String {
	    if n < 0 {
	      return "Negative"
	    } else if n == 0 {
	      return "Zero"
	    } else if n < 10 {
	      return "Small"
	    } else if n < 100 {
	      return "Big"
	    } else if n < 1000 {
	      return "Huge"
	    }

	    if specialNumbers.containsKey(n) {
	      return specialNumbers[n]!
	    }

	    return "Enormous"
	  }

	  pub fun factorial(_ n: Int): Int {
	    pre {
	      n >= 0:
	        "factorial is only defined for integers greater than or equal to zero"
	    }
	    post {
	      result >= 1:
	        "the result must be greater than or equal to 1"
	    }

	    if n < 1 {
	      return 1
	    }

	    return n * factorial(n - 1)
	  }
	`)

	script := []byte(`
	  import "imported"

	  pub fun main(): Int {
	    let testInputs: {Int: String} = {
	      -1: "Negative",
	      0: "Zero",
	      9: "Small",
	      99: "Big",
	      999: "Huge",
	      1001: "Enormous",
	      1729: "Harshad",
	      8128: "Harmonic",
	      41041: "Carmichael"
	    }

	    for input in testInputs.keys {
	      let result = getIntegerTrait(input)
	      assert(result == testInputs[input])
	    }

	    addSpecialNumber(78557, "Sierpinski")
	    assert("Sierpinski" == getIntegerTrait(78557))

	    factorial(5)
	    factorial(0)

	    return 42
	  }
	`)

	runtimeInterface := &testRuntimeInterface{
		getCode: func(location Location) (bytes []byte, err error) {
			switch location {
			case common.StringLocation("imported"):
				return importedScript, nil
			default:
				return nil, fmt.Errorf("unknown import location: %s", location)
			}
		},
	}

	coverageReport := NewCoverageReport()

	value, err := runtime.ExecuteScript(
		Script{
			Source: script,
		},
		Context{
			Interface:      runtimeInterface,
			Location:       common.ScriptLocation{},
			CoverageReport: coverageReport,
		},
	)
	require.NoError(t, err)

	assert.Equal(t, cadence.NewInt(42), value)

	actual, err := json.Marshal(coverageReport)
	require.NoError(t, err)

	expected := `
	  {
	    "coverage": {
	      "S.imported": {
	        "line_hits": {
	          "13": 10,
	          "14": 1,
	          "15": 9,
	          "16": 1,
	          "17": 8,
	          "18": 1,
	          "19": 7,
	          "20": 1,
	          "21": 6,
	          "22": 1,
	          "25": 5,
	          "26": 4,
	          "29": 1,
	          "34": 7,
	          "38": 7,
	          "42": 7,
	          "43": 2,
	          "46": 5,
	          "9": 1
	        },
	        "missed_lines": [],
	        "statements": 19,
	        "percentage": "100.0%"
	      },
	      "s.0000000000000000000000000000000000000000000000000000000000000000": {
	        "line_hits": {
	          "17": 1,
	          "18": 9,
	          "19": 9,
	          "22": 1,
	          "23": 1,
	          "25": 1,
	          "26": 1,
	          "28": 1,
	          "5": 1
	        },
	        "missed_lines": [],
	        "statements": 9,
	        "percentage": "100.0%"
	      }
	    }
	  }
	`
	require.JSONEq(t, expected, string(actual))

	assert.Equal(
		t,
		"Coverage: 100.0% of statements",
		coverageReport.CoveredStatementsPercentage(),
	)
}

func TestRuntimeCoverageWithExcludedLocation(t *testing.T) {

	t.Parallel()

	runtime := NewInterpreterRuntime(Config{
		CoverageReportingEnabled: true,
	})

	importedScript := []byte(`
	  pub let specialNumbers: {Int: String} = {
	    1729: "Harshad",
	    8128: "Harmonic",
	    41041: "Carmichael"
	  }

	  pub fun addSpecialNumber(_ n: Int, _ trait: String) {
	    specialNumbers[n] = trait
	  }

	  pub fun getIntegerTrait(_ n: Int): String {
	    if n < 0 {
	      return "Negative"
	    } else if n == 0 {
	      return "Zero"
	    } else if n < 10 {
	      return "Small"
	    } else if n < 100 {
	      return "Big"
	    } else if n < 1000 {
	      return "Huge"
	    }

	    if specialNumbers.containsKey(n) {
	      return specialNumbers[n]!
	    }

	    return "Enormous"
	  }
	`)

	script := []byte(`
	  import "imported"

	  pub fun main(): Int {
	    let testInputs: {Int: String} = {
	      -1: "Negative",
	      0: "Zero",
	      9: "Small",
	      99: "Big",
	      999: "Huge",
	      1001: "Enormous",
	      1729: "Harshad",
	      8128: "Harmonic",
	      41041: "Carmichael"
	    }

	    for input in testInputs.keys {
	      let result = getIntegerTrait(input)
	      assert(result == testInputs[input])
	    }

	    addSpecialNumber(78557, "Sierpinski")
	    assert("Sierpinski" == getIntegerTrait(78557))

	    return 42
	  }
	`)

	runtimeInterface := &testRuntimeInterface{
		getCode: func(location Location) (bytes []byte, err error) {
			switch location {
			case common.StringLocation("imported"):
				return importedScript, nil
			default:
				return nil, fmt.Errorf("unknown import location: %s", location)
			}
		},
	}

	coverageReport := NewCoverageReport()
	scriptlocation := common.ScriptLocation{}
	coverageReport.ExcludeLocation(scriptlocation)

	value, err := runtime.ExecuteScript(
		Script{
			Source: script,
		},
		Context{
			Interface:      runtimeInterface,
			Location:       scriptlocation,
			CoverageReport: coverageReport,
		},
	)
	require.NoError(t, err)

	assert.Equal(t, cadence.NewInt(42), value)

	actual, err := json.Marshal(coverageReport)
	require.NoError(t, err)

	expected := `
	  {
	    "coverage": {
	      "S.imported": {
	        "line_hits": {
	          "13": 10,
	          "14": 1,
	          "15": 9,
	          "16": 1,
	          "17": 8,
	          "18": 1,
	          "19": 7,
	          "20": 1,
	          "21": 6,
	          "22": 1,
	          "25": 5,
	          "26": 4,
	          "29": 1,
	          "9": 1
	        },
	        "missed_lines": [],
	        "statements": 14,
	        "percentage": "100.0%"
	      }
	    }
	  }
	`
	require.JSONEq(t, expected, string(actual))

	assert.Equal(
		t,
		"Coverage: 100.0% of statements",
		coverageReport.CoveredStatementsPercentage(),
	)
}
