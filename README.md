# Fiber Middleware Group

This package is a thin wrapper around Fiber V3's Router. Fiber matches Express.js spec which causes middleware to be applied by prefix rather than group. This means applying middleware to a group with the `/` prefix is equivalent to applying a global middleware, but that's not always desirable.

## Installation

```
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

## License

MIT

Copyright (c) 2026 Joe Tams
