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
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

type paramsKeyType string

const paramsKey paramsKeyType = "HodorParam"

func ParamOfReq(r *http.Request, name string) (string, bool) {
	return ParamOfCtx(r.Context(), name)
}

func ParamOfCtx(ctx context.Context, name string) (string, bool) {
	if p, ok := ctx.Value(paramsKey).(Params); ok {
		return p.Get(name)
	}
	return "", false
}

type Params interface {
	Get(string) (string, bool)
}

type emptyParams struct{}

func (p *emptyParams) Get(_ string) (string, bool) {
	return "", false
}

var background = new(emptyParams)

type valueParam struct {
	parent Params
	name   string
	value  string
}

func (p *valueParam) Get(name string) (string, bool) {
	if p.name == name {
		return p.value, true
	}
	return p.parent.Get(name)
}

func withValue(p Params, name, value string) Params {
	return &valueParam{
		parent: p,
		name:   name,
		value:  value,
	}
}

type node struct {
	parent   *node
	named    bool
	empty    bool
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
		empty:    true,
		handlers: map[Method]http.Handler{},
		children: map[byte]*node{},
	}
}

func (n *node) addRoute(method Method, pattern string, handler http.Handler) {
	if n.empty {
		n.init(method, pattern, handler)
		return
	}
	if n.named {
		n.addNamedRoute(method, pattern, handler)
		return
	}
	i := longestPrefix(n.pattern, pattern)
	if i < len(n.pattern) {
		n.splitAt(i)
	}
	if i == len(pattern) {
		n.handle(method, handler)
		return
	}
	n.getChildMust(pattern[i]).addRoute(method, pattern[i:], handler)
}

func (n *node) init(method Method, pattern string, handler http.Handler) {
	n.empty = false
	i := longestPrefix(pattern, pattern)
	if i == 0 {
		n.addNamedRoute(method, pattern, handler)
		return
	}
	if i == len(pattern) {
		n.pattern = pattern
		n.handle(method, handler)
		return
	}
	n.pattern = pattern[:i]
	n.getChildMust(pattern[i]).addRoute(method, pattern[i:], handler)
}

func (n *node) addNamedRoute(method Method, pattern string, handler http.Handler) {
	index := strings.Index(pattern, "/")
	var name string
	if index == -1 {
		name = pattern[1:]
		n.name = name
		n.handle(method, handler)
	} else {
		name = pattern[1:index]
		n.name = name
		n.getChildMust(pattern[index]).addRoute(method, pattern[index:], handler)
	}

}

func longestPrefix(p1, p2 string) int {
	i := 0
	for i < len(p1) && i < len(p2) && p1[i] == p2[i] && p2[i] != ':' {
		i++
	}
	return i
}

func (n *node) handle(method Method, handler http.Handler) {
	if _, ok := n.handlers[method]; ok {
		panic("duplicated handlers for same method")
	}
	n.handlers[method] = handler
}

func (n *node) splitAt(index int) {
	child := newNode(n, n.pattern[index:])
	child.handlers = n.handlers
	child.children = n.children
	child.named = false
	child.empty = false

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

func (n *node) match(p Params, method Method, pattern string) (Params, http.Handler, error) {
	if n.named {
		return n.matchNamed(p, method, pattern)
	}
	i := longestPrefix(n.pattern, pattern)
	if i < len(n.pattern) {
		return p, nil, errNotFound
	}
	if i < len(pattern) {
		if child, ok := n.children[pattern[i]]; ok {
			if p, h, err := child.match(p, method, pattern[i:]); err == nil || err == errMethodNotAllowed {
				return p, h, err
			}
			if child, ok := n.children[':']; ok {
				return child.match(p, method, pattern[i:])
			}
			return p, nil, errNotFound
		}
		if child, ok := n.children[':']; ok {
			return child.match(p, method, pattern[i:])
		}
		return p, nil, errNotFound
	}
	return n.handleMethod(p, method)
}

func (n *node) matchNamed(p Params, method Method, pattern string) (Params, http.Handler, error) {
	index := strings.Index(pattern, "/")
	if index == -1 {
		value := pattern
		return n.handleMethod(withValue(p, n.name, value), method)
	}
	value := pattern[:index]
	if child, ok := n.children['/']; ok {
		return child.match(withValue(p, n.name, value), method, pattern[index:])
	}
	return p, nil, errNotFound
}

func (n *node) handleMethod(p Params, method Method) (Params, http.Handler, error) {
	if len(n.handlers) == 0 {
		return p, nil, errNotFound
	}
	if h, ok := n.handlers[method]; ok {
		return p, h, nil
	}
	return p, nil, errMethodNotAllowed
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
