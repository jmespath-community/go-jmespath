/*
Basic command line interface for debug and testing purposes.

Examples:

Only print the AST for the expression:

	jp.go -ast "foo.bar.baz"

Evaluate the JMESPath expression against JSON data from a file:

	jp.go -input /tmp/data.json "foo.bar.baz"

This program can also be used as an executable to the jp-compliance
runner (github.com/jmespath-community/jmespath.test).
*/
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/jmespath-community/go-jmespath/pkg/api"
	"github.com/jmespath-community/go-jmespath/pkg/parsing"
	"github.com/spf13/cobra"
)

func main() {
	var command command
	cmd := &cobra.Command{
		Use:  "jpgo",
		Args: cobra.ExactArgs(1),
		RunE: command.run,
	}
	cmd.Flags().BoolVar(&command.astOnly, "ast", false, "Print the AST for the input expression and exit.")
	cmd.Flags().StringVar(&command.inputFile, "input", "", "Filename containing JSON data to search. If not provided, data is read from stdin.")
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

type command struct {
	astOnly   bool
	inputFile string
}

func (c command) run(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return errors.New("error: expected a single argument (the JMESPath expression)")
	}
	expression := args[0]
	parser := parsing.NewParser()
	parsed, err := parser.Parse(expression)
	if err != nil {
		if syntaxError, ok := err.(parsing.SyntaxError); ok {
			return fmt.Errorf("%s\n%s", syntaxError, syntaxError.HighlightLocation())
		}
		return err
	}
	if c.astOnly {
		fmt.Println("")
		fmt.Printf("%s\n", parsed)
		return nil
	}
	var inputData []byte
	if c.inputFile != "" {
		inputData, err = os.ReadFile(c.inputFile)
		if err != nil {
			return fmt.Errorf("error loading file %s: %w", c.inputFile, err)
		}
	} else {
		// If an input data file is not provided then we read the
		// data from stdin.
		inputData, err = io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("error reading from stdin: %w", err)
		}
	}
	var data interface{}
	if err := json.Unmarshal(inputData, &data); err != nil {
		return fmt.Errorf("invalid input JSON: %w", err)
	}
	result, err := api.Search(expression, data)
	if err != nil {
		return fmt.Errorf("error executing expression: %w", err)
	}
	toJSON, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("error serializing result to JSON: %w", err)
	}
	fmt.Println(string(toJSON))
	return nil
}
