package httprouter

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestApplicationGroup_GroupReturnsWebApplication(t *testing.T) {
	app := NewApplication(http.NewServeMux())
	group := app.Group("/api")

	if _, ok := group.(WebApplication); !ok {
		t.Fatalf("expected group to implement WebApplication, got %T", group)
	}
}

func TestApplicationGroup_GroupJoinsNestedPaths(t *testing.T) {
	mux := http.NewServeMux()
	app := NewApplication(mux)

	app.Group("/api").Group("/v1").Get("/users", stringHandler("users", http.StatusOK))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rr.Code)
	}
	if rr.Body.String() != "users" {
		t.Fatalf("expected body %q, got %q", "users", rr.Body.String())
	}
}

func TestApplicationGroup_HandleMethods(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		register       func(WebApplication, string, http.Handler, ...Middleware)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "post",
			method:         http.MethodPost,
			register:       WebApplication.Post,
			expectedStatus: http.StatusCreated,
			expectedBody:   "post",
		},
		{
			name:           "get",
			method:         http.MethodGet,
			register:       WebApplication.Get,
			expectedStatus: http.StatusOK,
			expectedBody:   "get",
		},
		{
			name:           "put",
			method:         http.MethodPut,
			register:       WebApplication.Put,
			expectedStatus: http.StatusAccepted,
			expectedBody:   "put",
		},
		{
			name:           "patch",
			method:         http.MethodPatch,
			register:       WebApplication.Patch,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "patch",
		},
		{
			name:           "delete",
			method:         http.MethodDelete,
			register:       WebApplication.Delete,
			expectedStatus: http.StatusResetContent,
			expectedBody:   "delete",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mux := http.NewServeMux()
			group := NewApplication(mux).Group("/api")

			tt.register(group, "/resource", stringHandler(tt.expectedBody, tt.expectedStatus))

			req := httptest.NewRequest(tt.method, "/api/resource", nil)
			rr := httptest.NewRecorder()

			mux.ServeHTTP(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Fatalf("expected status %d, got %d", tt.expectedStatus, rr.Code)
			}
			if rr.Body.String() != tt.expectedBody {
				t.Fatalf("expected body %q, got %q", tt.expectedBody, rr.Body.String())
			}
		})
	}
}

func TestApplicationGroup_MiddlewareOrder(t *testing.T) {
	mux := http.NewServeMux()
	app := NewApplication(mux)

	var calls []string
	middleware := func(name string) Middleware {
		return func(next http.Handler) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				calls = append(calls, name)
				next.ServeHTTP(w, r)
			}
		}
	}

	group := app.Group("/api", middleware("group"))
	group.WithGlobalMiddlewares(middleware("group-global"))
	group.Get("/users", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls = append(calls, "handler")
	}), middleware("route"))

	req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)

	expected := []string{"group", "group-global", "route", "handler"}
	if !reflect.DeepEqual(calls, expected) {
		t.Fatalf("expected calls %v, got %v", expected, calls)
	}
}

func TestApplicationGroup_NestedMiddlewareOrder(t *testing.T) {
	mux := http.NewServeMux()
	app := NewApplication(mux)

	var calls []string
	middleware := func(name string) Middleware {
		return func(next http.Handler) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				calls = append(calls, name)
				next.ServeHTTP(w, r)
			}
		}
	}

	app.Group("/api", middleware("api")).
		Group("/v1", middleware("v1")).
		Get("/users", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			calls = append(calls, "handler")
		}), middleware("route"))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)

	expected := []string{"api", "v1", "route", "handler"}
	if !reflect.DeepEqual(calls, expected) {
		t.Fatalf("expected calls %v, got %v", expected, calls)
	}
}

func TestApplicationGroup_DifferentGroupsMiddlewareOrder(t *testing.T) {
	calls := make(map[string][]string)

	appendCall := func(handlerName string, call string) {
		calls[handlerName] = append(calls[handlerName], call)
	}

	middleware := func(call string) Middleware {
		return func(next http.Handler) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				handlerName := r.Header.Get("handler-name")
				appendCall(handlerName, call)
				next.ServeHTTP(w, r)
			}
		}
	}

	mux := http.NewServeMux()
	app := NewApplication(mux, middleware("app"))

	v1 := app.
		Group("/api", middleware("api")).
		Group("/v1", middleware("v1"))

	users := v1.Group("/user", middleware("user"))
	users.Get("/name", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		appendCall(r.Header.Get("handler-name"), "user-handler")
	}), middleware("route"))

	applications := v1.Group("/application", middleware("application"))
	applications.Get("/name", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		appendCall(r.Header.Get("handler-name"), "application-handler")
	}), middleware("route"))

	tests := []struct {
		name     string
		flowName string
		path     string
		expected []string
	}{
		{
			name:     "user",
			flowName: "user",
			path:     "/api/v1/user/name",
			expected: []string{"app", "api", "v1", "user", "route", "user-handler"},
		},
		{
			name:     "application",
			flowName: "application",
			path:     "/api/v1/application/name",
			expected: []string{"app", "api", "v1", "application", "route", "application-handler"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			req.Header.Set("handler-name", tt.flowName)
			rr := httptest.NewRecorder()

			mux.ServeHTTP(rr, req)

			if !reflect.DeepEqual(calls[tt.flowName], tt.expected) {
				t.Fatalf("expected calls %v, got %v", tt.expected, calls[tt.flowName])
			}
		})
	}
}

func TestApplicationGroup_WithGlobalMiddlewaresReturnsSameGroup(t *testing.T) {
	app := NewApplication(http.NewServeMux())
	group := app.Group("/api")

	got := group.WithGlobalMiddlewares(func(next http.Handler) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r)
		}
	})

	if got != group {
		t.Fatalf("expected WithGlobalMiddlewares to return the same group")
	}
}

func TestApplicationGroup_GroupCombinesMiddlewaresWithoutMutatingInputs(t *testing.T) {
	groupMiddleware := func(next http.Handler) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r)
		}
	}
	routeMiddleware := func(next http.Handler) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r)
		}
	}

	group := &ApplicationGroup{
		mws:  []Middleware{groupMiddleware},
		path: "/api",
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	path, gotHandler, middlewares := group.group("/users", handler, routeMiddleware)

	if path != "/api/users" {
		t.Fatalf("expected path %q, got %q", "/api/users", path)
	}
	if reflect.DeepEqual(gotHandler, handler) {
		t.Fatalf("expected handler to be returned unchanged")
	}
	if len(middlewares) != 2 {
		t.Fatalf("expected 2 middlewares, got %d", len(middlewares))
	}
	if reflect.ValueOf(middlewares[0]).Pointer() != reflect.ValueOf(groupMiddleware).Pointer() {
		t.Fatalf("expected first middleware to be the group middleware")
	}
	if reflect.ValueOf(middlewares[1]).Pointer() != reflect.ValueOf(routeMiddleware).Pointer() {
		t.Fatalf("expected second middleware to be the route middleware")
	}
	if len(group.mws) != 1 {
		t.Fatalf("expected group middlewares to remain unchanged, got %d", len(group.mws))
	}
}
