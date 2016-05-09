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

// MethodHandler creates a new Handler which only accepts specific method.
func MethodHandler(method Method, handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == method.String() {
			handler.ServeHTTP(w, r)
			return
		}
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	})
}

type serverRouter struct {
	*http.ServeMux
}

func (sr *serverRouter) AddRoute(method Method, pattern string, handler http.Handler, filters ...Filter) {
	sr.Handle(pattern, MergeFilters(filters...).Do(MethodHandler(method, handler)))
}

// WithServer creates new Router with ServeMux in http package.
func WithServer(srv *http.ServeMux) Router {
	return &serverRouter{
		ServeMux: srv,
	}
}
