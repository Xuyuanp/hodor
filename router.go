/*
 * Copyright 2015 Xuyuan Pang
 * Author: Pang Xuyuan <xuyuanp # gmail dot com>
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

import "net/http"

// Method is HTTP method.
type Method string

// available methods.
const (
	OPTIONS Method = "OPTIONS"
	GET            = "GET"
	HEAD           = "HEAD"
	POST           = "POST"
	PUT            = "PUT"
	DELETE         = "DELETE"
	TRACE          = "TRACE"
	CONNECT        = "CONNECT"
	PATCH          = "PATCH"
)

// Methods is a list of all valid methods.
var Methods = []Method{
	OPTIONS,
	GET,
	HEAD,
	POST,
	PUT,
	DELETE,
	TRACE,
	CONNECT,
	PATCH,
}

func (m Method) String() string {
	return string(m)
}

// Router interface
type Router interface {
	http.Handler
	AddRoute(method Method, pattern string, handler http.Handler, filters ...Filter)
}

// NodeRouter struct
type NodeRouter struct {
	root       *node
	Handler404 http.Handler
	Handler405 http.Handler
}

func errHandler(status int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, http.StatusText(status), status)
	}
}

// NewRouter creates a new router
func NewRouter() *NodeRouter {
	return &NodeRouter{
		root:       newNode(nil, ""),
		Handler404: errHandler(http.StatusNotFound),
		Handler405: errHandler(http.StatusMethodNotAllowed),
	}
}

func (nr *NodeRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h, err := nr.root.match(Method(r.Method), r.URL.Path)
	switch err {
	case errNotFound:
		nr.Handler404.ServeHTTP(w, r)
	case errMethodNotAllowed:
		nr.Handler405.ServeHTTP(w, r)
	case nil:
		h.ServeHTTP(w, r)
	default:
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// AddRoute implements Router interface
func (nr *NodeRouter) AddRoute(method Method, pattern string, handler http.Handler, filters ...Filter) {
	nr.root.addRoute(method, pattern, handler, filters...)
}
