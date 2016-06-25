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

import (
	"net/http"
	"runtime"
	"time"
)

// Filter interface
type Filter interface {
	Do(http.Handler) http.Handler
}

// FilterFunc type
type FilterFunc func(next http.Handler) http.Handler

// Do calls FilterFunc
func (f FilterFunc) Do(next http.Handler) http.Handler {
	return f(next)
}

var emptyFilter FilterFunc = func(next http.Handler) http.Handler {
	return next
}

// MergeFilters merges multi fileers into a single one.
func MergeFilters(filters ...Filter) Filter {
	return FilterFunc(
		func(next http.Handler) http.Handler {
			for i := len(filters) - 1; i >= 0; i-- {
				next = filters[i].Do(next)
			}
			return next
		})
}

// Logger interface
type Logger interface {
	Printf(string, ...interface{})
}

// LogFilter new filter
func LogFilter(l Logger) FilterFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				start := time.Now()
				l.Printf("%s %s", r.Method, r.RemoteAddr)
				next.ServeHTTP(w, r)
				status := w.(ResponseWriter).Status()
				l.Printf("%d %s %s", status, http.StatusText(status), time.Since(start))
			})
	}
}

func RecoveryFilter(l Logger) FilterFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				defer func() {
					if err := recover(); err != nil {
						trace := make([]byte, 1<<16)
						n := runtime.Stack(trace, true)
						stack := trace[:n]
						l.Printf("PANIC: %v\n%s", err, stack)
						http.Error(w, http.StatusText(http.StatusInternalServerError),
							http.StatusInternalServerError)
					}
				}()
				next.ServeHTTP(w, r)
			})
	}
}
