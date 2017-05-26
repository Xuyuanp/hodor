/*
 * Copyright 2017 Xuyuan Pang
 * Author: Xuyuan Pang
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

package hodor

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
)

type node struct {
	parent   *node
	named    bool
	pattern  string
	name     string
	handlers map[Method]http.Handler
	children map[byte]*node
}

func newNode(parent *node, pattern string) *node {
	return &node{
		parent:   parent,
		pattern:  pattern,
		named:    false,
		handlers: map[Method]http.Handler{},
		children: map[byte]*node{},
	}
}

func (n *node) addRoute(method Method, pattern string, handler http.Handler, filters ...Filter) {
	if !n.named && n.pattern == "" {
		n.pattern = pattern
		n.handle(method, handler, filters...)
		return
	}
	if strings.HasPrefix(pattern, ":") {
		n.addNamedRoute(method, pattern, handler, filters...)
		return
	}
	i := longestPrefix(n.pattern, pattern)
	if i < len(n.pattern) {
		n.splitAt(i)
	}
	if i == len(pattern) {
		n.handle(method, handler, filters...)
		return
	}
	n.getChildMust(pattern[i]).addRoute(method, pattern[i:], handler, filters...)
}

func (n *node) addNamedRoute(method Method, pattern string, handler http.Handler, filters ...Filter) {
	n.named = true
	index := strings.Index(pattern, "/")
	var name string
	if index == -1 {
		name = pattern[1:]
		n.name = name
		n.handle(method, handler, filters...)
	} else {
		name = pattern[1:index]
		n.name = name
		n.getChildMust(pattern[index]).addRoute(method, pattern[index:], handler, filters...)
	}

}

func longestPrefix(p1, p2 string) int {
	i := 0
	for i < len(p1) && i < len(p2) && p1[i] == p2[i] && p2[i] != ':' {
		i++
	}
	return i
}

func (n *node) handle(method Method, handler http.Handler, filters ...Filter) {
	if _, ok := n.handlers[method]; ok {
		panic("duplicated handler for same method")
	}
	n.handlers[method] = MergeFilters(filters...).Do(handler)
}

func (n *node) splitAt(index int) {
	child := newNode(n, n.pattern[index:])
	child.handlers = n.handlers
	child.children = n.children

	n.handlers = map[Method]http.Handler{}
	n.children = map[byte]*node{n.pattern[index]: child}
	n.pattern = n.pattern[:index]
}

func (n *node) getChildMust(c byte) *node {
	if child, ok := n.children[c]; ok {
		return child
	}
	child := newNode(n, "")
	n.children[c] = child
	if c == ':' {
		child.named = true
	}
	return child
}

var (
	errNotFound         = errors.New("Not Found")
	errMethodNotAllowed = errors.New("Method Not Allowed")
)

func (n *node) match(method Method, pattern string) (http.Handler, error) {
	if n.named {
		return n.matchNamed(method, pattern)
	}
	i := longestPrefix(n.pattern, pattern)
	if i < len(n.pattern) {
		return nil, errNotFound
	}
	if i < len(pattern) {
		if child, ok := n.children[pattern[i]]; ok {
			return child.match(method, pattern[i:])
		}
		if child, ok := n.children[':']; ok {
			return child.match(method, pattern[i:])
		}
		return nil, errNotFound
	}
	return n.handleMethod(method)
}

func (n *node) matchNamed(method Method, pattern string) (http.Handler, error) {
	index := strings.Index(pattern, "/")
	if index == -1 {
		// value := pattern
		// fmt.Println(value)
		return n.handleMethod(method)
	}
	// value := pattern[:index]
	// fmt.Println(value)
	if child, ok := n.children['/']; ok {
		return child.match(method, pattern[index:])
	}
	return nil, errNotFound
}

func (n *node) handleMethod(method Method) (http.Handler, error) {
	if len(n.handlers) == 0 {
		return nil, errNotFound
	}
	if h, ok := n.handlers[method]; ok {
		return h, nil
	}
	return nil, errMethodNotAllowed
}

func (n *node) printTree(prefix string) {
	pattern := n.pattern
	if n.named {
		pattern = ":" + n.name
	}
	for method := range n.handlers {
		fmt.Println(method, prefix+pattern)
	}
	for _, child := range n.children {
		child.printTree(prefix + pattern)
	}
}
