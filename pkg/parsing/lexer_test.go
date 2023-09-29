package parsing

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

var lexingTests = []struct {
	expression string
	expected   []token
}{
	{"*", []token{{TOKStar, "*", 0, 1}}},
	{".", []token{{TOKDot, ".", 0, 1}}},
	{"[?", []token{{TOKFilter, "[?", 0, 2}}},
	{"[]", []token{{TOKFlatten, "[]", 0, 2}}},
	{"(", []token{{TOKLparen, "(", 0, 1}}},
	{")", []token{{TOKRparen, ")", 0, 1}}},
	{"[", []token{{TOKLbracket, "[", 0, 1}}},
	{"]", []token{{TOKRbracket, "]", 0, 1}}},
	{"{", []token{{TOKLbrace, "{", 0, 1}}},
	{"}", []token{{TOKRbrace, "}", 0, 1}}},
	{"||", []token{{TOKOr, "||", 0, 2}}},
	{"|", []token{{TOKPipe, "|", 0, 1}}},
	{"29", []token{{TOKNumber, "29", 0, 2}}},
	{"2", []token{{TOKNumber, "2", 0, 1}}},
	{"0", []token{{TOKNumber, "0", 0, 1}}},
	{"-20", []token{{TOKNumber, "-20", 0, 3}}},
	{"foo", []token{{TOKUnquotedIdentifier, "foo", 0, 3}}},
	{`"bar"`, []token{{TOKQuotedIdentifier, "bar", 0, 3}}},
	// Arithmetic operators
	{"+", []token{{TOKPlus, "+", 0, 1}}},
	{"/", []token{{TOKDivide, "/", 0, 1}}},
	{"\u2212", []token{{TOKMinus, "\u2212", 0, 1}}},
	{"\u00d7", []token{{TOKMultiply, "\u00d7", 0, 1}}},
	{"\u00f7", []token{{TOKDivide, "\u00f7", 0, 1}}},
	{"%", []token{{TOKModulo, "%", 0, 1}}},
	{"//", []token{{TOKDiv, "//", 0, 2}}},
	{"- 20", []token{
		{TOKMinus, "-", 0, 1},
		{TOKNumber, "20", 2, 2},
	}},
	// Escaping the delimiter
	{`"bar\"baz"`, []token{{TOKQuotedIdentifier, `bar"baz`, 0, 7}}},
	{",", []token{{TOKComma, ",", 0, 1}}},
	{":", []token{{TOKColon, ":", 0, 1}}},
	{"<", []token{{TOKLT, "<", 0, 1}}},
	{"<=", []token{{TOKLTE, "<=", 0, 2}}},
	{">", []token{{TOKGT, ">", 0, 1}}},
	{">=", []token{{TOKGTE, ">=", 0, 2}}},
	{"==", []token{{TOKEQ, "==", 0, 2}}},
	{"!=", []token{{TOKNE, "!=", 0, 2}}},
	{"`[0, 1, 2]`", []token{{TOKJSONLiteral, "[0, 1, 2]", 1, 9}}},
	{"'foo'", []token{{TOKStringLiteral, "foo", 1, 3}}},
	{"'\\\\'", []token{{TOKStringLiteral, `\`, 1, 1}}},
	{"'a'", []token{{TOKStringLiteral, "a", 1, 1}}},
	{`'foo\'bar'`, []token{{TOKStringLiteral, "foo'bar", 1, 7}}},
	{"@", []token{{TOKCurrent, "@", 0, 1}}},
	{"$", []token{{TOKRoot, "$", 0, 1}}},
	{"&", []token{{TOKExpref, "&", 0, 1}}},
	// Quoted identifier unicode escape sequences
	{`"\u2713"`, []token{{TOKQuotedIdentifier, "✓", 0, 3}}},
	{`"\\"`, []token{{TOKQuotedIdentifier, `\`, 0, 1}}},
	{"`\"foo\"`", []token{{TOKJSONLiteral, "\"foo\"", 1, 5}}},
	// Combinations of tokens.
	{"foo.bar", []token{
		{TOKUnquotedIdentifier, "foo", 0, 3},
		{TOKDot, ".", 3, 1},
		{TOKUnquotedIdentifier, "bar", 4, 3},
	}},
	{"foo[0]", []token{
		{TOKUnquotedIdentifier, "foo", 0, 3},
		{TOKLbracket, "[", 3, 1},
		{TOKNumber, "0", 4, 1},
		{TOKRbracket, "]", 5, 1},
	}},
	{"foo[?a<b]", []token{
		{TOKUnquotedIdentifier, "foo", 0, 3},
		{TOKFilter, "[?", 3, 2},
		{TOKUnquotedIdentifier, "a", 5, 1},
		{TOKLT, "<", 6, 1},
		{TOKUnquotedIdentifier, "b", 7, 1},
		{TOKRbracket, "]", 8, 1},
	}},
	// let expressions
	{"$root", []token{{TOKVarref, "$root", 0, 5}}},
	{"$root = @", []token{
		{TOKVarref, "$root", 0, 5},
		{TOKAssign, "=", 6, 1},
		{TOKCurrent, "@", 8, 1},
	}},
}

func TestCanLexTokens(t *testing.T) {
	assert := assert.New(t)
	lexer := NewLexer()
	for _, tt := range lexingTests {
		tokens, err := lexer.Tokenize(tt.expression)
		if assert.Nil(err) {
			errMsg := fmt.Sprintf("Mismatch expected number of tokens: (expected: %s, actual: %s)",
				tt.expected, tokens)
			tt.expected = append(tt.expected, token{TOKEOF, "", len(tt.expression), 0})
			if assert.Equal(len(tt.expected), len(tokens), errMsg) {
				for i, token := range tokens {
					expected := tt.expected[i]
					assert.Equal(expected, token, "Token not equal")
				}
			}
		}
	}
}

var lexingErrorTests = []struct {
	expression string
	msg        string
}{
	{"'foo", "Missing closing single quote"},
	{"[?foo==bar?]", "Unknown char '?'"},
}

func TestLexingErrors(t *testing.T) {
	assert := assert.New(t)
	lexer := NewLexer()
	for _, tt := range lexingErrorTests {
		_, err := lexer.Tokenize(tt.expression)
		assert.NotNil(err, fmt.Sprintf("Expected lexing error: %s", tt.msg))
	}
}

var (
	exprIdentifier          = "abcdefghijklmnopqrstuvwxyz"
	exprSubexpr             = "abcdefghijklmnopqrstuvwxyz.abcdefghijklmnopqrstuvwxyz"
	deeplyNested50          = "j49.j48.j47.j46.j45.j44.j43.j42.j41.j40.j39.j38.j37.j36.j35.j34.j33.j32.j31.j30.j29.j28.j27.j26.j25.j24.j23.j22.j21.j20.j19.j18.j17.j16.j15.j14.j13.j12.j11.j10.j9.j8.j7.j6.j5.j4.j3.j2.j1.j0"
	deeplyNested50Pipe      = "j49|j48|j47|j46|j45|j44|j43|j42|j41|j40|j39|j38|j37|j36|j35|j34|j33|j32|j31|j30|j29|j28|j27|j26|j25|j24|j23|j22|j21|j20|j19|j18|j17|j16|j15|j14|j13|j12|j11|j10|j9|j8|j7|j6|j5|j4|j3|j2|j1|j0"
	deeplyNested50Index     = "[49][48][47][46][45][44][43][42][41][40][39][38][37][36][35][34][33][32][31][30][29][28][27][26][25][24][23][22][21][20][19][18][17][16][15][14][13][12][11][10][9][8][7][6][5][4][3][2][1][0]"
	deepProjection104       = "a[*].b[*].c[*].d[*].e[*].f[*].g[*].h[*].i[*].j[*].k[*].l[*].m[*].n[*].o[*].p[*].q[*].r[*].s[*].t[*].u[*].v[*].w[*].x[*].y[*].z[*].a[*].b[*].c[*].d[*].e[*].f[*].g[*].h[*].i[*].j[*].k[*].l[*].m[*].n[*].o[*].p[*].q[*].r[*].s[*].t[*].u[*].v[*].w[*].x[*].y[*].z[*].a[*].b[*].c[*].d[*].e[*].f[*].g[*].h[*].i[*].j[*].k[*].l[*].m[*].n[*].o[*].p[*].q[*].r[*].s[*].t[*].u[*].v[*].w[*].x[*].y[*].z[*].a[*].b[*].c[*].d[*].e[*].f[*].g[*].h[*].i[*].j[*].k[*].l[*].m[*].n[*].o[*].p[*].q[*].r[*].s[*].t[*].u[*].v[*].w[*].x[*].y[*].z[*]"
	exprQuotedIdentifier    = `"abcdefghijklmnopqrstuvwxyz.abcdefghijklmnopqrstuvwxyz"`
	quotedIdentifierEscapes = `"\n\r\b\t\n\r\b\t\n\r\b\t\n\r\b\t\n\r\b\t\n\r\b\t\n\r\b\t"`
	rawStringLiteral        = `'abcdefghijklmnopqrstuvwxyz.abcdefghijklmnopqrstuvwxyz'`
)

func BenchmarkLexIdentifier(b *testing.B) {
	runLexBenchmark(b, exprIdentifier)
}

func BenchmarkLexSubexpression(b *testing.B) {
	runLexBenchmark(b, exprSubexpr)
}

func BenchmarkLexDeeplyNested50(b *testing.B) {
	runLexBenchmark(b, deeplyNested50)
}

func BenchmarkLexDeepNested50Pipe(b *testing.B) {
	runLexBenchmark(b, deeplyNested50Pipe)
}

func BenchmarkLexDeepNested50Index(b *testing.B) {
	runLexBenchmark(b, deeplyNested50Index)
}

func BenchmarkLexQuotedIdentifier(b *testing.B) {
	runLexBenchmark(b, exprQuotedIdentifier)
}

func BenchmarkLexQuotedIdentifierEscapes(b *testing.B) {
	runLexBenchmark(b, quotedIdentifierEscapes)
}

func BenchmarkLexRawStringLiteral(b *testing.B) {
	runLexBenchmark(b, rawStringLiteral)
}

func BenchmarkLexDeepProjection104(b *testing.B) {
	runLexBenchmark(b, deepProjection104)
}

func runLexBenchmark(b *testing.B, expression string) {
	b.Helper()
	assert := assert.New(b)
	lexer := NewLexer()
	for i := 0; i < b.N; i++ {
		_, err := lexer.Tokenize(expression)
		if err != nil {
			assert.Fail("Could not lex expression")
		}
	}
}
