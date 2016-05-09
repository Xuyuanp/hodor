/*
 * Copyright 2016 Xuyuan Pang
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

import "net/http"

// BuildRoute creates a route builder from router.
func BuildRoute(router Router) Route {
	return RouteFunc(router.AddRoute).Route()
}

// RouteFunc is a function type implemented Router interface.
type RouteFunc func(method Method, pattern string, handler http.Handler, filters ...Filter)

// Route returns a setter-chain to add a new route step-by-step.
func (f RouteFunc) Route() Route {
	return func(method Method) PatternSetter {
		return func(pattern string) HandlerSetter {
			return func(handler http.Handler, filters ...Filter) {
				f(method, pattern, handler, filters...)
			}
		}
	}
}

// Route easy way to set Method for a route.
type Route func(method Method) PatternSetter

// Method calls MethodSetter function.
func (ms Route) Method(method Method) PatternSetter {
	return ms(method)
}

// Options short for Method(OPTIONS)
func (ms Route) Options() PatternSetter {
	return ms.Method(OPTIONS)
}

// Get short for Method(GET)
func (ms Route) Get() PatternSetter {
	return ms.Method(GET)
}

// Head short for Method(HEAD)
func (ms Route) Head() PatternSetter {
	return ms.Method(HEAD)
}

// Post short for Method(POST)
func (ms Route) Post() PatternSetter {
	return ms.Method(POST)
}

// Put short for Method(PUT)
func (ms Route) Put() PatternSetter {
	return ms.Method(PUT)
}

// Delete short for Method(DELETE)
func (ms Route) Delete() PatternSetter {
	return ms.Method(DELETE)
}

// Trace short for Method(TRACE)
func (ms Route) Trace() PatternSetter {
	return ms.Method(TRACE)
}

// Connect short for Method(CONNECT)
func (ms Route) Connect() PatternSetter {
	return ms.Method(CONNECT)
}

// Patch short for Method(PATCH)
func (ms Route) Patch() PatternSetter {
	return ms.Method(PATCH)
}

// Group creates a Group with root
func (ms Route) Group(root string) Grouper {
	return func(fn func(Route), fs ...Filter) {
		fn(RouteFunc(
			func(method Method, subpattern string, handler http.Handler, subfilters ...Filter) {
				ms.Method(method).
					Pattern(root + subpattern).
					Filters(fs...).
					Filters(subfilters...).
					Handler(handler)
			}).Route())
	}
}

// HandlerSetter easy way to set Handler for a route.
type HandlerSetter func(handler http.Handler, filters ...Filter)

// Handler calls HandlerSetter function.
func (hs HandlerSetter) Handler(handler http.Handler) {
	hs(handler)
}

// HandlerFunc wraps HandlerFunc to Handler
func (hs HandlerSetter) HandlerFunc(hf func(http.ResponseWriter, *http.Request)) {
	hs.Handler(http.HandlerFunc(hf))
}

// Filters returns a new HandlerSetter
func (hs HandlerSetter) Filters(filters ...Filter) HandlerSetter {
	return func(handler http.Handler, fs ...Filter) {
		hs(handler, append(filters, fs...)...)
	}
}

// PatternSetter easy way to set Path for a route.
type PatternSetter func(pattern string) HandlerSetter

// Pattern calls Pattern function.
func (ps PatternSetter) Pattern(pattern string) HandlerSetter {
	return ps(pattern)
}

// Grouper is to add routes grouply.
type Grouper func(func(Route), ...Filter)

// For applies func
func (g Grouper) For(fn func(Route)) {
	g(fn)
}

// Filters creates new Grouper wrapped filters
func (g Grouper) Filters(filters ...Filter) Grouper {
	return func(fn func(Route), fs ...Filter) {
		g(fn, append(filters, fs...)...)
	}
}
