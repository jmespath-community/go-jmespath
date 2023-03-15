package jmespath

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/jmespath-community/go-jmespath/pkg/parsing"
	"github.com/stretchr/testify/assert"
)

type TestSuite struct {
	Given     interface{}
	TestCases []TestCase `json:"cases"`
	Comment   string
}

type TestCase struct {
	Comment    string
	Expression string
	Result     interface{}
	Error      string
}

var whiteListed = []string{
	"compliance/tests/jep-12/jep-12-literal.json",
	"compliance/tests/arithmetic.json",
	"compliance/tests/basic.json",
	"compliance/tests/boolean.json",
	"compliance/tests/current.json",
	"compliance/tests/escape.json",
	"compliance/tests/filters.json",
	"compliance/tests/functions.json",
	"compliance/tests/function_group_by.json",
	// "compliance/tests/function_let.json",
	"compliance/tests/function_strings.json",
	"compliance/tests/identifiers.json",
	"compliance/tests/indices.json",
	// "compliance/tests/lexical_scoping.json",
	"compliance/tests/literal.json",
	"compliance/tests/multiselect.json",
	"compliance/tests/ormatch.json",
	"compliance/tests/pipe.json",
	"compliance/tests/slice.json",
	"compliance/tests/syntax.json",
	"compliance/tests/unicode.json",
	"compliance/tests/wildcard.json",
}

func allowed(path string) bool {
	for _, el := range whiteListed {
		if el == path {
			return true
		}
	}
	return false
}

func TestCompliance(t *testing.T) {
	assert := assert.New(t)

	var complianceFiles []string
	err := filepath.Walk("compliance", func(path string, _ os.FileInfo, _ error) error {
		// if strings.HasSuffix(path, ".json") {
		if allowed(path) {
			complianceFiles = append(complianceFiles, path)
		}
		return nil
	})
	if assert.Nil(err) {
		for _, filename := range complianceFiles {
			runComplianceTest(assert, filename)
		}
	}
}

func runComplianceTest(assert *assert.Assertions, filename string) {
	var testSuites []TestSuite
	data, err := os.ReadFile(filename)
	if assert.Nil(err) {
		err := json.Unmarshal(data, &testSuites)
		if assert.Nil(err) {
			for _, testsuite := range testSuites {
				runTestSuite(assert, testsuite, filename)
			}
		}
	}
}

func runTestSuite(assert *assert.Assertions, testsuite TestSuite, filename string) {
	for _, testcase := range testsuite.TestCases {
		if testcase.Error != "" {
			// This is a test case that verifies we error out properly.
			runSyntaxTestCase(assert, testsuite.Given, testcase, filename)
		} else {
			runTestCase(assert, testsuite.Given, testcase, filename)
		}
	}
}

func runSyntaxTestCase(assert *assert.Assertions, given interface{}, testcase TestCase, filename string) {
	// Anything with an .Error means that we expect that JMESPath should return
	// an error when we try to evaluate the expression.
	// fmt.Println(fmt.Sprintf("%s: %s", filename, testcase.Expression))
	_, err := Search(testcase.Expression, given, nil)
	assert.NotNil(err, fmt.Sprintf("Expression: %s", testcase.Expression))
}

func runTestCase(assert *assert.Assertions, given interface{}, testcase TestCase, filename string) {
	lexer := parsing.NewLexer()
	var err error
	_, err = lexer.Tokenize(testcase.Expression)
	if err != nil {
		errMsg := fmt.Sprintf("(%s) Could not lex expression: %s -- %s", filename, testcase.Expression, err.Error())
		assert.Fail(errMsg)
		return
	}
	parser := parsing.NewParser()
	_, err = parser.Parse(testcase.Expression)
	if err != nil {
		errMsg := fmt.Sprintf("(%s) Could not parse expression: %s -- %s", filename, testcase.Expression, err.Error())
		assert.Fail(errMsg)
		return
	}
	actual, err := Search(testcase.Expression, given, nil)
	if assert.Nil(err, fmt.Sprintf("Expression: %s", testcase.Expression)) {
		assert.Equal(testcase.Result, actual, fmt.Sprintf("Expression: %s", testcase.Expression))
	}
}
