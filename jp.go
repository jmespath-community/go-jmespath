package jmespath

import (
	"github.com/jmespath-community/go-jmespath/pkg/api"
	"github.com/jmespath-community/go-jmespath/pkg/functions"
)

// api types

type JMESPath = api.JMESPath

var (
	Compile     = api.Compile
	MustCompile = api.MustCompile
	Search      = api.Search
)

// function types

type (
	ExecuteFunc   = functions.ExecuteFunc
	JpFunction    = functions.JpFunction
	JpType        = functions.JpType
	FunctionEntry = functions.FunctionEntry
	ArgSpec       = functions.ArgSpec
	ExpRef        = functions.ExpRef
)

const (
	JpNumber      = functions.JpNumber
	JpString      = functions.JpString
	JpArray       = functions.JpArray
	JpObject      = functions.JpObject
	JpArrayArray  = functions.JpArrayArray
	JpArrayNumber = functions.JpArrayNumber
	JpArrayString = functions.JpArrayString
	JpExpref      = functions.JpExpref
	JpAny         = functions.JpAny
)
