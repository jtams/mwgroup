// mwgroup is a wrapper around fiber.Router that allows middleware to be
// added to groups. https://github.com/gofiber/fiber/issues/2276
package mwgroup

import (
	"github.com/gofiber/fiber/v3"
)

type MiddlewareGroup struct {
	router fiber.Router
	prefix string
	mw     []any
}

// Creates a MiddlewareGroup rooted at prefix on app.
// It behaves like Fiber's app.Group(prefix) for routing purposes,
func New(app *fiber.App, prefix string, handlers ...any) *MiddlewareGroup {
	return &MiddlewareGroup{
		router: app,
		prefix: prefix,
		mw:     append([]any{}, handlers...),
	}
}

// Builds a MiddlewareGroup rooted at
// an arbitrary fiber.Router
func newChild(router fiber.Router, prefix string, inherited []any, handlers ...any) *MiddlewareGroup {
	mw := make([]any, 0, len(inherited)+len(handlers))
	mw = append(mw, inherited...)
	mw = append(mw, handlers...)
	return &MiddlewareGroup{
		router: router,
		prefix: prefix,
		mw:     mw,
	}
}

// Appends middleware to this group's own middleware stack. It does NOT
// touch the underlying fiber router's prefix-based middleware table, so it
// cannot leak onto other groups or apps sharing the same path prefix.
func (g *MiddlewareGroup) Use(handlers ...any) *MiddlewareGroup {
	g.mw = append(g.mw, handlers...)
	return g
}

// Group creates a nested MiddlewareGroup. Child groups inherit
// the parent's middleware (applied first) plus any additional
// handlers
func (g *MiddlewareGroup) Group(prefix string, handlers ...any) *MiddlewareGroup {
	sub := g.router.Group(g.prefix + prefix)
	return newChild(sub, "", g.mw, handlers...)
}

// Builds the final ordered handler list
func (g *MiddlewareGroup) chain(extraAndHandler []any) (any, []any) {
	all := make([]any, 0, len(g.mw)+len(extraAndHandler))
	all = append(all, g.mw...)
	all = append(all, extraAndHandler...)
	if len(all) == 0 {
		return func(c fiber.Ctx) error { return c.Next() }, nil
	}
	return all[0], all[1:]
}

// Registers path for the given HTTP methods, placing group's
// middleware before the per-route middleware and the handler.
// handlersAndFinal must have the actual route handler as its LAST element,
// with any extra route-specific middleware before it.
func (g *MiddlewareGroup) route(methods []string, path string, handlersAndFinal []any) fiber.Router {
	first, rest := g.chain(handlersAndFinal)
	return g.router.Add(methods, g.prefix+path, first, rest...)
}

// ======== Method Wrappers ========
func (g *MiddlewareGroup) Get(path string, handler any, extra ...any) fiber.Router {
	return g.route([]string{fiber.MethodGet}, path, append(extra, handler))
}

func (g *MiddlewareGroup) Head(path string, handler any, extra ...any) fiber.Router {
	return g.route([]string{fiber.MethodHead}, path, append(extra, handler))
}

func (g *MiddlewareGroup) Post(path string, handler any, extra ...any) fiber.Router {
	return g.route([]string{fiber.MethodPost}, path, append(extra, handler))
}

func (g *MiddlewareGroup) Put(path string, handler any, extra ...any) fiber.Router {
	return g.route([]string{fiber.MethodPut}, path, append(extra, handler))
}

func (g *MiddlewareGroup) Delete(path string, handler any, extra ...any) fiber.Router {
	return g.route([]string{fiber.MethodDelete}, path, append(extra, handler))
}

func (g *MiddlewareGroup) Connect(path string, handler any, extra ...any) fiber.Router {
	return g.route([]string{fiber.MethodConnect}, path, append(extra, handler))
}

func (g *MiddlewareGroup) Options(path string, handler any, extra ...any) fiber.Router {
	return g.route([]string{fiber.MethodOptions}, path, append(extra, handler))
}

func (g *MiddlewareGroup) Trace(path string, handler any, extra ...any) fiber.Router {
	return g.route([]string{fiber.MethodTrace}, path, append(extra, handler))
}

func (g *MiddlewareGroup) Patch(path string, handler any, extra ...any) fiber.Router {
	return g.route([]string{fiber.MethodPatch}, path, append(extra, handler))
}

func (g *MiddlewareGroup) All(path string, handler any, extra ...any) fiber.Router {
	methods := []string{
		fiber.MethodGet, fiber.MethodHead, fiber.MethodPost, fiber.MethodPut,
		fiber.MethodDelete, fiber.MethodConnect, fiber.MethodOptions,
		fiber.MethodTrace, fiber.MethodPatch,
	}
	return g.route(methods, path, append(extra, handler))
}

func (g *MiddlewareGroup) Add(methods []string, path string, handler any, extra ...any) fiber.Router {
	return g.route(methods, path, append(extra, handler))
}

// Returns the wrapped fiber.Router
func (g *MiddlewareGroup) Router() fiber.Router {
	return g.router
}

// Returns the path prefix for this group
func (g *MiddlewareGroup) Prefix() string {
	return g.prefix
}
