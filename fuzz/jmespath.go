package jmespath

import "github.com/jmespath-community/go-jmespath/pkg/parsing"

// Fuzz will fuzz test the JMESPath parser.
func Fuzz(data []byte) int {
	p := parsing.NewParser()
	_, err := p.Parse(string(data))
	if err != nil {
		return 1
	}
	return 0
}
