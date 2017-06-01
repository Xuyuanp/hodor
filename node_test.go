package hodor

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

var defaultHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "%s", r.URL.Path)
})

func TestRouter(t *testing.T) {
	router := NewRouter()

	routes := []struct {
		method  Method
		pattern string
	}{
		{GET, "/fuck/you"},
		{GET, "/fuck/me"},
		{POST, "/fuck/you/again"},
		{GET, "/"},
		{GET, "/shit"},
	}
	for _, r := range routes {
		router.AddRoute(r.method, r.pattern, defaultHandler)
	}

	for _, r := range routes {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(r.method.String(), r.pattern, nil)
		router.ServeHTTP(w, req)
		if w.Code != 200 {
			t.Errorf("%s %s not match", r.method, r.pattern)
		}
		if got, exp := w.Body.String(), r.pattern; got != exp {
			t.Errorf("%s %s response not match. exp: %s, got: %s ",
				r.method, r.pattern, exp, got)
		}
	}

	cases := []struct {
		method  Method
		pattern string
		code    int
	}{
		{POST, "/fuck/you", 405},
		{GET, "/fuck/her", 404},
		{POST, "/fuck/her", 404},
	}

	for _, c := range cases {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(c.method.String(), c.pattern, nil)
		router.ServeHTTP(w, req)
		if got, exp := w.Code, c.code; got != exp {
			t.Errorf("%s %s code not match. exp: %d, got: %d ",
				c.method, c.pattern, exp, got)
		}
	}
}

func TestNamedNode(t *testing.T) {
	router := NewRouter()

	routes := []struct {
		method  Method
		pattern string
	}{
		{GET, "/fuck/you"},
		{POST, "/fuck/you"},
		{DELETE, "/fuck/you"},
		{GET, "/fuck/me"},
		{POST, "/fuck/:name/again"},
		{GET, "/"},
		{GET, "/shit"},
		{GET, "/:name"},
		{POST, "/:name"},
	}
	for _, r := range routes {
		router.AddRoute(r.method, r.pattern, defaultHandler)
	}

	cases := []struct {
		method  Method
		pattern string
		code    int
	}{
		{POST, "/fuck/you", 200},
		{DELETE, "/fuck/you", 200},
		{HEAD, "/fuck/you", 405},
		{GET, "/fuck/her", 404},
		{GET, "/fuck/me", 200},
		{POST, "/fuck/her", 404},
		{POST, "/fuck/her/again", 200},
		{POST, "/fuck/you/again", 200},
		{POST, "/fuck/you/once", 404},
		{GET, "/alice", 200},
		{POST, "/alice", 200},
	}

	for _, c := range cases {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(c.method.String(), c.pattern, nil)
		router.ServeHTTP(w, req)
		if got, exp := w.Code, c.code; got != exp {
			t.Errorf("%s %s code not match. exp: %d, got: %d ",
				c.method, c.pattern, exp, got)
		}
	}
}
