# Fiber Middleware Group

This package is a thin wrapper around Fiber V3's Router. Fiber follows Express.js conventions which causes middleware to be applied by prefix rather than group. This means applying middleware to a group with the `/` prefix is equivalent to applying a global middleware, but that's not always desirable.

## Installation

```bash
go get github.com/jtams/mwgroup
```

## Usage

```go
package main

import (
	"log"

	"github.com/gofiber/fiber/v3"
	"github.com/jtams/mwgroup"
    "myproject/middlewares"
)


func main() {
	app := fiber.New()

	private := mwgroup.New(app, "/api/v1")
	private.Use(middlewares.RequestLogger(), middlewares.RequireAuth())
	private.Get("/me", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{"user": "jtams"})
	})

	public := mwgroup.New(app, "/api/v1")
	public.Use(middlewares.RequestLogger())
    // With Fiber's router, public would also include the RequireAuth middleware
	public.Get("/status", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})


	log.Fatal(app.Listen(":3000"))
}
```

## Documentation

`mwgroup.New` returns the router wrapper: `*mwgroup.MiddlewareGroup`

`mwgroup.MiddlewareGroup` includes HTTP methods: `Get`, `Head`, `Post`, `Put`, `Delete`, `Connect`, `Options`, `Trace`, and `Patch`. Also include `All` to match all. They all follow the same pattern as Fiber: `group.Get(path string, handler any, extra ...any)`

`mwgroup.MiddlewareGroup.Add(methods []string, path string, handler any, handlers ...any)`

`mwgroup.MiddlewareGroup.Router()` returns the underlying Fiber Router: `fiber.Router`

`mwgroup.MiddlewareGroup.Prefix()` returns the prefix: `string`

## Why?

Fiber follows Express.js in the way it handles middleware. The problem is that middleware is applied to prefixes rather than groups. This is counter intuitive because `Group` feels like it should solve this, but groups are just so you don't have to write the prefix each time. Creating a new group with prefix `/` and then applying a middleware to it is equivalent to creating a global middleware.

It can lead to situations where we have two middlewares and we want one group with one middleware, another group with the other middleware, and finally another group with both:

```go
xGroup := app.Group("/api", middleware.X())
// middleware.X() is applied to /api/exampleX
xGroup.Get("/exampleX", xHandler)

yGroup := app.Group("/api", middleware.Y())
// middleware.Y() is applied to /api/exampleY
// BUT middleware.X() is still applied to /api/exampleY
// because it was not applied to xGroup,
// rather it was applied to the /api prefix
yGroup.Get("/exampleY", yHandler)

// middleware.X() AND middleware.Y() is already applied to /api/exampleXY
// from the previous xGroup and yGroup
xyGroup := app.Group("/api")
xyGroup.Get("/exampleXY", xyHandler)
```

One solution is to put the middleware inline:

```go
groups := app.Group("/api")
groups.Get("/exampleX", middleware.X(), xHandler)
groups.Get("/exampleY", middleware.Y(), yHandler)
groups.Get("/exampleXY", middleware.X(), middleware.Y(), xyHandler)
```

This is what `mwgroup` does under the hood, but it can get ugly without the wrapper.

The Fiber/Express solution is _better_ routing. For example grouping like: /api/x, /api/y, and /api/xy. The middleware is applied to each group. This can work, but sometimes I want routes on the same prefix. Like /home, /about, and /account, where /account requires auth.

## License

MIT

Copyright (c) 2026 Joe Tams
