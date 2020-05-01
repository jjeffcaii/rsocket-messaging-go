package internal_test

import (
	"bufio"
	"fmt"
	"regexp"
	"strings"
	"testing"

	. "github.com/jjeffcaii/rsocket-messaging-go/internal"
	"github.com/stretchr/testify/assert"
)

func TestRegexp(t *testing.T) {
	var r = regexp.MustCompilePOSIX("^\\{([a-zA-Z_][a-zA-Z0-9_]*)}$")

	fmt.Println(r.FindStringSubmatch("xxs{abc}"))

}

func TestAny(t *testing.T) {
	pt := NewPathTrie()
	pt.AddPath("students.{id}.v1", 123)
	node, ok := pt.Load("students.777.v1")
	assert.True(t, ok)
	assert.True(t, node.IsLeaf())
	assert.Equal(t, 123, node.Value())
}

func TestVariables(t *testing.T) {
	pt := NewPathTrie()
	pt.AddPath("students.{id}.courses.{course}.score", 100)
	variables, value, ok := pt.Find("students.123.courses.cs.score")
	assert.True(t, ok, "find failed")
	assert.Equal(t, 100, value, "bad value")
	assert.Equal(t, "123", variables.GetOrDefault("id", ""), "bad path var")
	assert.Equal(t, "cs", variables.GetOrDefault("course", ""), "bad path var")
}

func TestPathTrie(t *testing.T) {

	var (
		node *TrieNode
		ok   bool
	)

	pt := NewPathTrie()
	pt.AddPath("foo.bar", 111)
	pt.AddPath("foo/bar/1", 222)
	pt.AddPath("bar", 333)

	node, ok = pt.Load("foo/bar")
	assert.True(t, ok)
	assert.True(t, node.IsLeaf())
	assert.Equal(t, 111, node.Value())

	node, ok = pt.Load("foo/bar/1")
	assert.True(t, ok)
	assert.True(t, node.IsLeaf())
	assert.Equal(t, 222, node.Value())

	node, ok = pt.Load("bar")
	assert.True(t, ok)
	assert.True(t, node.IsLeaf())
	assert.Equal(t, 333, node.Value())

	node, ok = pt.Load("foo")
	assert.True(t, ok)
	assert.False(t, node.IsLeaf())
}

func TestSplitPath(t *testing.T) {
	scanner := bufio.NewScanner(strings.NewReader("foo.bar"))
	scanner.Split(SplitPath)
	var results []string
	for scanner.Scan() {
		results = append(results, scanner.Text())
	}
	assert.Equal(t, "foo,bar", strings.Join(results, ","))
}
