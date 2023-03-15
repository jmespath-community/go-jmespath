package interpreter

// import (
// 	"testing"

// 	"github.com/stretchr/testify/assert"
// )

// func TestScopesMissing(t *testing.T) {
// 	assert := assert.New(t)
// 	scopes := newScope(nil)

// 	_, found := scopes.GetValue("foo")
// 	assert.False(found)
// }

// func TestScopesRoot(t *testing.T) {
// 	assert := assert.New(t)
// 	scopes := newScope(map[string]interface{}{"foo": "bar"})

// 	value, found := scopes.GetValue("foo")
// 	assert.True(found)
// 	assert.Equal("bar", value.(string))
// }

// func TestScopesNested(t *testing.T) {
// 	assert := assert.New(t)
// 	scopes := newScope(nil)

// 	{
// 		scopes := scopes.With(map[string]interface{}{"foo": "bar", "qux": "quux"})

// 		{
// 			scopes := scopes.With(map[string]interface{}{"foo": "baz"})

// 			value, found := scopes.GetValue("foo")
// 			assert.True(found)
// 			assert.Equal("baz", value.(string))

// 			value, found = scopes.GetValue("qux")
// 			assert.True(found)
// 			assert.Equal("quux", value.(string))
// 		}

// 		value, found := scopes.GetValue("foo")
// 		assert.True(found)
// 		assert.Equal("bar", value.(string))
// 	}

// 	_, found := scopes.GetValue("foo")
// 	assert.False(found)
// }
