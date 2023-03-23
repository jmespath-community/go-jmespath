// Code generated by "stringer -type=TokType"; DO NOT EDIT.

package parsing

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[TOKUnknown-0]
	_ = x[TOKStar-1]
	_ = x[TOKDot-2]
	_ = x[TOKFilter-3]
	_ = x[TOKFlatten-4]
	_ = x[TOKLparen-5]
	_ = x[TOKRparen-6]
	_ = x[TOKLbracket-7]
	_ = x[TOKRbracket-8]
	_ = x[TOKLbrace-9]
	_ = x[TOKRbrace-10]
	_ = x[TOKOr-11]
	_ = x[TOKPipe-12]
	_ = x[TOKNumber-13]
	_ = x[TOKUnquotedIdentifier-14]
	_ = x[TOKQuotedIdentifier-15]
	_ = x[TOKComma-16]
	_ = x[TOKColon-17]
	_ = x[TOKPlus-18]
	_ = x[TOKMinus-19]
	_ = x[TOKMultiply-20]
	_ = x[TOKDivide-21]
	_ = x[TOKModulo-22]
	_ = x[TOKDiv-23]
	_ = x[TOKLT-24]
	_ = x[TOKLTE-25]
	_ = x[TOKGT-26]
	_ = x[TOKGTE-27]
	_ = x[TOKEQ-28]
	_ = x[TOKNE-29]
	_ = x[TOKJSONLiteral-30]
	_ = x[TOKStringLiteral-31]
	_ = x[TOKCurrent-32]
	_ = x[TOKRoot-33]
	_ = x[TOKExpref-34]
	_ = x[TOKAnd-35]
	_ = x[TOKNot-36]
	_ = x[TOKLet-37]
	_ = x[TOKIn-38]
	_ = x[TOKVarref-39]
	_ = x[TOKAssign-40]
	_ = x[TOKEOF-41]
}

const _TokType_name = "TOKUnknownTOKStarTOKDotTOKFilterTOKFlattenTOKLparenTOKRparenTOKLbracketTOKRbracketTOKLbraceTOKRbraceTOKOrTOKPipeTOKNumberTOKUnquotedIdentifierTOKQuotedIdentifierTOKCommaTOKColonTOKPlusTOKMinusTOKMultiplyTOKDivideTOKModuloTOKDivTOKLTTOKLTETOKGTTOKGTETOKEQTOKNETOKJSONLiteralTOKStringLiteralTOKCurrentTOKRootTOKExprefTOKAndTOKNotTOKLetTOKInTOKVarrefTOKAssignTOKEOF"

var _TokType_index = [...]uint16{0, 10, 17, 23, 32, 42, 51, 60, 71, 82, 91, 100, 105, 112, 121, 142, 161, 169, 177, 184, 192, 203, 212, 221, 227, 232, 238, 243, 249, 254, 259, 273, 289, 299, 306, 315, 321, 327, 333, 338, 347, 356, 362}

func (i TokType) String() string {
	if i < 0 || i >= TokType(len(_TokType_index)-1) {
		return "TokType(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _TokType_name[_TokType_index[i]:_TokType_index[i+1]]
}
