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
	"log"
	"net/http"
	"os"
)

// Hodor struct
type Hodor struct {
	router Router
	filter Filter
}

// NewHodor creates new Hodor with Router
func NewHodor(router Router) *Hodor {
	h := &Hodor{
		router: router,
		filter: emptyFilter,
	}
	return h
}

// Default return a default Hodor with log and recover filters
func Default() *Hodor {
	h := NewHodor(NewRouter())
	logger := log.New(os.Stdout, "[Hodor] ", log.LstdFlags)
	h.AddFilters(
		LogFilter(logger),
		RecoveryFilter(logger),
	)
	return h
}

func (h *Hodor) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.filter.Do(h.router).ServeHTTP(NewResponseWriter(w), r)
}

// Route returns root route
func (h *Hodor) Route() Route {
	return BuildRoute(h.router)
}

// Run on addr
func (h *Hodor) Run(addr string) error {
	return http.ListenAndServe(addr, h)
}

// AddFilters appends filters to current filter
func (h *Hodor) AddFilters(filters ...Filter) {
	h.filter = MergeFilters(h.filter, MergeFilters(filters...))
}

// SetFilters replace current filter with the merged filters
func (h *Hodor) SetFilters(filters ...Filter) {
	h.filter = MergeFilters(filters...)
}
