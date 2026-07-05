package mwgroup

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
)

func trailMW(name string) fiber.Handler {
	return func(c fiber.Ctx) error {
		prev, _ := c.Locals("trail").(string)
		if prev != "" {
			prev += ","
		}
		c.Locals("trail", prev+name)
		return c.Next()
	}
}

func trailHandler(c fiber.Ctx) error {
	trail, _ := c.Locals("trail").(string)
	return c.SendString(trail)
}

func doRequest(t *testing.T, app *fiber.App, method, path string) (string, int) {
	t.Helper()
	req := httptest.NewRequest(method, path, nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test(%s %s): %v", method, path, err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("reading response body: %v", err)
	}
	return string(body), resp.StatusCode
}

func TestNewMiddlewareGroup(t *testing.T) {
	app := fiber.New()
	g := New(app, "/api", trailMW("a"))

	if g.Router() != fiber.Router(app) {
		t.Errorf("Router() = %v, want app", g.Router())
	}
	if g.Prefix() != "/api" {
		t.Errorf("Prefix() = %q, want %q", g.Prefix(), "/api")
	}
	if len(g.mw) != 1 {
		t.Fatalf("len(mw) = %d, want 1", len(g.mw))
	}

	g.Get("/x", trailHandler)
	body, status := doRequest(t, app, http.MethodGet, "/api/x")
	if status != http.StatusOK {
		t.Fatalf("status = %d, want %d", status, http.StatusOK)
	}
	if body != "a" {
		t.Errorf("body = %q, want %q", body, "a")
	}
}

func TestMiddlewareGroup_Use(t *testing.T) {
	app := fiber.New()
	g := New(app, "/api", trailMW("a"))
	ret := g.Use(trailMW("b"))

	if ret != g {
		t.Errorf("Use() should return the same group for chaining")
	}
	if len(g.mw) != 2 {
		t.Fatalf("len(mw) = %d, want 2", len(g.mw))
	}

	g.Get("/x", trailHandler)
	body, _ := doRequest(t, app, http.MethodGet, "/api/x")
	if body != "a,b" {
		t.Errorf("body = %q, want %q", body, "a,b")
	}
}

func TestMiddlewareGroup_Group(t *testing.T) {
	app := fiber.New()
	parent := New(app, "/api", trailMW("parent"))
	parent.Use(trailMW("parent2"))
	child := parent.Group("/v1", trailMW("child"))

	if child.Prefix() != "" {
		t.Errorf("child Prefix() = %q, want empty (prefix folded into router.Group)", child.Prefix())
	}
	if len(child.mw) != 3 {
		t.Fatalf("len(child.mw) = %d, want 3", len(child.mw))
	}

	child.Get("/x", trailHandler)
	body, status := doRequest(t, app, http.MethodGet, "/api/v1/x")
	if status != http.StatusOK {
		t.Fatalf("status = %d, want %d", status, http.StatusOK)
	}
	if body != "parent,parent2,child" {
		t.Errorf("body = %q, want %q", body, "parent,parent2,child")
	}

	// Middleware added to the child afterwards must not leak to the parent.
	child.Use(trailMW("child2"))
	parent.Get("/y", trailHandler)
	body, _ = doRequest(t, app, http.MethodGet, "/api/y")
	if body != "parent,parent2" {
		t.Errorf("parent route leaked child middleware: body = %q", body)
	}
}

func TestMiddlewareGroup_GroupIsolation(t *testing.T) {
	app := fiber.New()
	base := New(app, "/shared")
	g1 := base.Group("/one", trailMW("g1"))
	g2 := base.Group("/two", trailMW("g2"))

	g1.Get("/x", trailHandler)
	g2.Get("/x", trailHandler)

	body1, _ := doRequest(t, app, http.MethodGet, "/shared/one/x")
	if body1 != "g1" {
		t.Errorf("g1 body = %q, want %q", body1, "g1")
	}

	body2, _ := doRequest(t, app, http.MethodGet, "/shared/two/x")
	if body2 != "g2" {
		t.Errorf("g2 body = %q, want %q", body2, "g2")
	}
}

func TestMiddlewareGroup_chain(t *testing.T) {
	g := &MiddlewareGroup{mw: []any{"mw1", "mw2"}}

	first, rest := g.chain([]any{"extra", "handler"})
	if first != "mw1" {
		t.Errorf("first = %v, want %q", first, "mw1")
	}
	wantRest := []any{"mw2", "extra", "handler"}
	if len(rest) != len(wantRest) {
		t.Fatalf("len(rest) = %d, want %d", len(rest), len(wantRest))
	}
	for i, v := range wantRest {
		if rest[i] != v {
			t.Errorf("rest[%d] = %v, want %v", i, rest[i], v)
		}
	}
}

func TestMiddlewareGroup_chainEmpty(t *testing.T) {
	g := &MiddlewareGroup{}
	first, rest := g.chain(nil)
	if rest != nil {
		t.Errorf("rest = %v, want nil", rest)
	}
	if _, ok := first.(func(fiber.Ctx) error); !ok {
		t.Fatalf("first is %T, want func(fiber.Ctx) error", first)
	}
}

type methodTestCase struct {
	name     string
	method   string
	register func(g *MiddlewareGroup, path string, handler any) fiber.Router
}

func TestMiddlewareGroup_HTTPMethods(t *testing.T) {
	cases := []methodTestCase{
		{"Get", http.MethodGet, func(g *MiddlewareGroup, p string, h any) fiber.Router { return g.Get(p, h) }},
		{"Head", http.MethodHead, func(g *MiddlewareGroup, p string, h any) fiber.Router { return g.Head(p, h) }},
		{"Post", http.MethodPost, func(g *MiddlewareGroup, p string, h any) fiber.Router { return g.Post(p, h) }},
		{"Put", http.MethodPut, func(g *MiddlewareGroup, p string, h any) fiber.Router { return g.Put(p, h) }},
		{"Delete", http.MethodDelete, func(g *MiddlewareGroup, p string, h any) fiber.Router { return g.Delete(p, h) }},
		{"Connect", http.MethodConnect, func(g *MiddlewareGroup, p string, h any) fiber.Router { return g.Connect(p, h) }},
		{"Options", http.MethodOptions, func(g *MiddlewareGroup, p string, h any) fiber.Router { return g.Options(p, h) }},
		{"Trace", http.MethodTrace, func(g *MiddlewareGroup, p string, h any) fiber.Router { return g.Trace(p, h) }},
		{"Patch", http.MethodPatch, func(g *MiddlewareGroup, p string, h any) fiber.Router { return g.Patch(p, h) }},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			app := fiber.New()
			g := New(app, "/api", trailMW("mw"))
			tc.register(g, "/x", trailHandler)

			body, status := doRequest(t, app, tc.method, "/api/x")
			if status != http.StatusOK {
				t.Fatalf("status = %d, want %d", status, http.StatusOK)
			}
			// HEAD/TRACE/CONNECT responses may not carry a readable body
			// depending on the client/transport, so only assert body
			// content for methods that are expected to return one.
			if tc.method != http.MethodHead && body != "mw" {
				t.Errorf("body = %q, want %q", body, "mw")
			}
		})
	}
}

func TestMiddlewareGroup_ExtraPerRouteMiddleware(t *testing.T) {
	app := fiber.New()
	g := New(app, "/api", trailMW("group"))
	g.Get("/x", trailHandler, trailMW("extra"))

	body, _ := doRequest(t, app, http.MethodGet, "/api/x")
	if body != "group,extra" {
		t.Errorf("body = %q, want %q", body, "group,extra")
	}
}

func TestMiddlewareGroup_All(t *testing.T) {
	app := fiber.New()
	g := New(app, "/api", trailMW("mw"))
	g.All("/x", trailHandler)

	for _, method := range []string{
		http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete,
		http.MethodOptions, http.MethodPatch,
	} {
		body, status := doRequest(t, app, method, "/api/x")
		if status != http.StatusOK {
			t.Errorf("%s status = %d, want %d", method, status, http.StatusOK)
		}
		if body != "mw" {
			t.Errorf("%s body = %q, want %q", method, body, "mw")
		}
	}
}

func TestMiddlewareGroup_Add(t *testing.T) {
	app := fiber.New()
	g := New(app, "/api", trailMW("mw"))
	g.Add([]string{http.MethodGet, http.MethodPost}, "/x", trailHandler)

	body, status := doRequest(t, app, http.MethodGet, "/api/x")
	if status != http.StatusOK || body != "mw" {
		t.Errorf("GET: body=%q status=%d, want body=%q status=%d", body, status, "mw", http.StatusOK)
	}

	_, status = doRequest(t, app, http.MethodPost, "/api/x")
	if status != http.StatusOK {
		t.Errorf("POST status = %d, want %d", status, http.StatusOK)
	}

	// A method not in the explicit list should not be registered.
	_, status = doRequest(t, app, http.MethodDelete, "/api/x")
	if status == http.StatusOK {
		t.Errorf("DELETE status = %d, want not-OK (method should be unregistered)", status)
	}
}

func TestMiddlewareGroup_Router(t *testing.T) {
	app := fiber.New()
	g := New(app, "/api")
	if g.Router() != fiber.Router(app) {
		t.Errorf("Router() did not return the underlying app")
	}

	sub := app.Group("/nested")
	child := newChild(sub, "", nil)
	if child.Router() != sub {
		t.Errorf("Router() did not return the underlying nested router")
	}
}

func TestMiddlewareGroup_Prefix(t *testing.T) {
	app := fiber.New()
	g := New(app, "/api/v2")
	if g.Prefix() != "/api/v2" {
		t.Errorf("Prefix() = %q, want %q", g.Prefix(), "/api/v2")
	}
}
