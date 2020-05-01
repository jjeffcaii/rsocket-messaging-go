package internal

import (
	"bufio"
	"bytes"
	"fmt"
	"regexp"
	"strings"

	"github.com/pkg/errors"
)

var errMakeString = errors.New("make string failed")
var _regParam = regexp.MustCompilePOSIX("^\\{([a-zA-Z_][a-zA-Z0-9_]*)}$")

const _any = "{}"

func MkString(format string, args ...interface{}) (str string, err error) {
	defer func() {
		if e := recover(); e != nil {
			err = errMakeString
		}
	}()
	str = fmt.Sprintf(format, args...)
	return
}

type TrieNodeFlag uint8

type TrieNodeLeaf struct {
	value        interface{}
	paramIndices []int
	paramNames   []string
}

type TrieNode struct {
	name     string
	children map[string]*TrieNode
	parent   *TrieNode
	leaf     *TrieNodeLeaf
}

func (t *TrieNode) IsLeaf() bool {
	return t.leaf != nil
}

func (t *TrieNode) Value() interface{} {
	if t.leaf == nil {
		return nil
	}
	return t.leaf.value
}

func (t *TrieNode) Name() string {
	return t.name
}

func (t *TrieNode) getChild(path string) (*TrieNode, bool) {
	found, ok := t.children[path]
	return found, ok
}

func (t *TrieNode) addChild(path string, child *TrieNode) {
	if _, ok := t.children[path]; !ok {
		t.children[path] = child
	}
}

type PathTrie struct {
	rootNode *TrieNode
}

type PathVariables struct {
	names     []string
	variables []string
}

func (p *PathVariables) Get(name string) (variable string, ok bool) {
	if p == nil {
		return
	}
	for i := 0; i < len(p.names); i++ {
		if p.names[i] == name {
			variable = p.variables[i]
			ok = true
			return
		}
	}
	return
}

func (p *PathVariables) GetOrDefault(name string, defaultValue string) string {
	if p == nil {
		return defaultValue
	}
	v, ok := p.Get(name)
	if !ok {
		v = defaultValue
	}
	return v
}

func (p *PathVariables) GetOrCompute(name string, compute func() string) string {
	if p == nil {
		return compute()
	}
	v, ok := p.Get(name)
	if !ok && compute != nil {
		v = compute()
	}
	return v
}

func (p *PathTrie) Find(path string) (variables *PathVariables, value interface{}, ok bool) {
	var m map[int]string
	scanner := bufio.NewScanner(strings.NewReader(path))
	scanner.Split(SplitPath)
	parent := p.rootNode
	count := 0
	for scanner.Scan() {
		part := scanner.Text()
		child, exist := parent.getChild(part)
		if !exist {
			child, exist = parent.getChild(_any)
		}
		if !exist {
			return
		}
		parent = child
		if parent.name == _any {
			if m == nil {
				m = make(map[int]string)
			}
			m[count] = part
		}
		count++
	}

	leaf := parent.leaf
	if leaf == nil {
		return
	}
	ok = true
	value = leaf.value
	variablesAmount := len(leaf.paramIndices)
	if variablesAmount < 1 {
		return
	}
	variables = &PathVariables{}
	for i := 0; i < variablesAmount; i++ {
		index := leaf.paramIndices[i]
		name := leaf.paramNames[i]
		v, ok := m[index]
		if !ok {
			continue
		}
		variables.names = append(variables.names, name)
		variables.variables = append(variables.variables, v)
	}
	return
}

func (p *PathTrie) Load(path string) (*TrieNode, bool) {
	scanner := bufio.NewScanner(strings.NewReader(path))
	scanner.Split(SplitPath)
	parent := p.rootNode
	for scanner.Scan() {
		part := scanner.Text()
		child, ok := parent.getChild(part)
		if !ok {
			child, ok = parent.getChild(_any)
		}
		if !ok {
			return nil, false
		}
		parent = child
	}
	return parent, true
}

func (p *PathTrie) AddPath(path string, value interface{}) (err error) {
	var (
		indices []int
		names   []string
	)
	scanner := bufio.NewScanner(strings.NewReader(path))
	scanner.Split(SplitPath)
	parent := p.rootNode
	count := 0
	for scanner.Scan() {
		part := scanner.Text()

		// match {...} pattern
		groups := _regParam.FindStringSubmatch(part)
		if len(groups) == 2 {
			indices = append(indices, count)
			names = append(names, groups[1])
			part = _any
		}

		child, ok := parent.getChild(part)
		if !ok {
			child = newTrieNode(parent, part)
			parent.addChild(part, child)
		}
		parent = child
		count++
	}
	if parent.leaf != nil {
		return errors.Errorf("conflict path %s", path)
	}
	parent.leaf = &TrieNodeLeaf{
		value:        value,
		paramIndices: indices,
		paramNames:   names,
	}
	return
}

func NewPathTrie() *PathTrie {
	return &PathTrie{
		rootNode: newTrieNode(nil, ""),
	}
}

func SplitPath(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	i := -1
	if n := bytes.IndexByte(data, '/'); n > 0 {
		i = n
	}
	if n := bytes.IndexByte(data, '.'); n > 0 && (i < 0 || n < i) {
		i = n
	}
	if i >= 0 {
		return i + 1, data[0:i], nil
	}
	if atEOF {
		return len(data), data, nil
	}
	return 0, nil, nil
}

func newTrieNode(parent *TrieNode, path string) *TrieNode {
	return &TrieNode{
		parent:   parent,
		name:     path,
		children: make(map[string]*TrieNode),
	}
}
