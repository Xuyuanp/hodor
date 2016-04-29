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
	return func(method Method) PatternSetter {
		return func(pattern string) HandlerSetter {
			return func(handler http.Handler, filters ...Filter) {
				router.AddRoute(method, pattern, handler, filters...)
			}
		}
	}
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
