package evo

import (
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
	fstatic "github.com/gofiber/fiber/v3/middleware/static"
)

// newTestApp creates a bare Fiber v3 app and sets the package-level app variable
// so evo routing helpers (Get, Post, Redirect, etc.) can be called in tests.
func newTestApp() *fiber.App {
	a := fiber.New(fiber.Config{})
	app = a
	return a
}

// TestFiberCtxIsInterface verifies that fiber.Ctx is now an interface and that
// our Upgrade() wrapper correctly wraps it into a *Request.
func TestFiberCtxIsInterface(t *testing.T) {
	a := newTestApp()

	a.Get("/hello", func(ctx fiber.Ctx) error {
		r := Upgrade(ctx)
		if r == nil {
			t.Error("Upgrade returned nil")
		}
		if r.Context == nil {
			t.Error("Request.Context is nil")
		}
		return ctx.SendString("ok")
	})

	req := httptest.NewRequest("GET", "/hello", nil)
	resp, err := a.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

// TestRouteParams verifies that the AllParams replacement (Route().Params + Params(key))
// correctly returns all route parameters.
func TestRouteParams(t *testing.T) {
	a := newTestApp()

	a.Get("/user/:id/order/:orderId", func(ctx fiber.Ctx) error {
		r := Upgrade(ctx)
		params := r.Params()
		if params["id"] != "42" {
			t.Errorf("expected id=42, got %q", params["id"])
		}
		if params["orderId"] != "99" {
			t.Errorf("expected orderId=99, got %q", params["orderId"])
		}
		return ctx.SendString("ok")
	})

	req := httptest.NewRequest("GET", "/user/42/order/99", nil)
	resp, err := a.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

// TestRouteParamsEmpty verifies that Params() returns an empty map when there are no route params.
func TestRouteParamsEmpty(t *testing.T) {
	a := newTestApp()

	a.Get("/no-params", func(ctx fiber.Ctx) error {
		r := Upgrade(ctx)
		params := r.Params()
		if len(params) != 0 {
			t.Errorf("expected empty params map, got %v", params)
		}
		return ctx.SendString("ok")
	})

	req := httptest.NewRequest("GET", "/no-params", nil)
	resp, err := a.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

// TestRedirectWithStatus verifies ctx.Redirect().Status(code).To(url) behavior.
func TestRedirectWithStatus(t *testing.T) {
	a := newTestApp()

	a.Get("/old", func(ctx fiber.Ctx) error {
		return ctx.Redirect().Status(301).To("/new")
	})

	req := httptest.NewRequest("GET", "/old", nil)
	resp, err := a.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 301 {
		t.Errorf("expected 301, got %d", resp.StatusCode)
	}
	if loc := resp.Header.Get("Location"); loc != "/new" {
		t.Errorf("expected Location=/new, got %q", loc)
	}
}

// TestRedirectDefault verifies redirect without explicit status uses 303 (Fiber v3 default).
func TestRedirectDefault(t *testing.T) {
	a := newTestApp()

	a.Get("/move", func(ctx fiber.Ctx) error {
		return ctx.Redirect().To("/dest")
	})

	req := httptest.NewRequest("GET", "/move", nil)
	resp, err := a.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	// Fiber v3 changed the default redirect code from 302 to 303 See Other.
	if resp.StatusCode != 303 {
		t.Errorf("expected 303 (fiber v3 default), got %d", resp.StatusCode)
	}
}

// TestRequestRedirect verifies the Request.Redirect wrapper.
func TestRequestRedirect(t *testing.T) {
	a := newTestApp()

	a.Get("/r1", func(ctx fiber.Ctx) error {
		r := Upgrade(ctx)
		return r.Redirect("/target", 303)
	})
	a.Get("/r2", func(ctx fiber.Ctx) error {
		r := Upgrade(ctx)
		return r.Redirect("/target2")
	})

	for _, tc := range []struct {
		path string
		code int
		loc  string
	}{
		{"/r1", 303, "/target"},
		{"/r2", 303, "/target2"}, // Fiber v3 default is 303
	} {
		req := httptest.NewRequest("GET", tc.path, nil)
		resp, err := a.Test(req)
		if err != nil {
			t.Fatalf("%s: %v", tc.path, err)
		}
		if resp.StatusCode != tc.code {
			t.Errorf("%s: expected %d, got %d", tc.path, tc.code, resp.StatusCode)
		}
		if loc := resp.Header.Get("Location"); loc != tc.loc {
			t.Errorf("%s: expected Location=%q, got %q", tc.path, tc.loc, loc)
		}
	}
}

// TestLocals verifies Locals() still works with fiber.Ctx as interface.
func TestLocals(t *testing.T) {
	a := newTestApp()

	a.Use(func(ctx fiber.Ctx) error {
		ctx.Locals("user", "alice")
		return ctx.Next()
	})
	a.Get("/locals", func(ctx fiber.Ctx) error {
		r := Upgrade(ctx)
		v := r.Var("user")
		return ctx.SendString(v.String())
	})

	req := httptest.NewRequest("GET", "/locals", nil)
	resp, err := a.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(resp.Body)
	if string(body) != "alice" {
		t.Errorf("expected 'alice', got %q", string(body))
	}
}

// TestHandleHelpers verifies the package-level evo routing helpers wrap fiber.Ctx correctly.
func TestHandleHelpers(t *testing.T) {
	a := newTestApp()

	Get("/ping", func(req *Request) any {
		return "pong"
	})

	req := httptest.NewRequest("GET", "/ping", nil)
	resp, err := a.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(resp.Body)
	// The handler returns "pong" → WriteResponse encodes it as JSON {"success":true,"data":"pong"}
	if !strings.Contains(string(body), "pong") {
		t.Errorf("expected body to contain 'pong', got %q", string(body))
	}
}

// TestSendStatus verifies SendStatus still works after removing *fiber.Ctx.
func TestSendStatus(t *testing.T) {
	a := newTestApp()

	a.Get("/teapot", func(ctx fiber.Ctx) error {
		return ctx.SendStatus(418)
	})

	req := httptest.NewRequest("GET", "/teapot", nil)
	resp, err := a.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 418 {
		t.Errorf("expected 418, got %d", resp.StatusCode)
	}
}

// TestStaticMiddleware verifies static.New() is used correctly (no panic, correct middleware chain).
func TestStaticMiddleware(t *testing.T) {
	a := newTestApp()
	// Just verify it registers without panic; actual file serving needs real files.
	_ = a.Use("/static", fstatic.New(".", fstatic.Config{}))
}

// TestRequestHeaders verifies header access via the Request wrapper.
func TestRequestHeaders(t *testing.T) {
	a := newTestApp()

	a.Get("/hdr", func(ctx fiber.Ctx) error {
		r := Upgrade(ctx)
		v := r.Header("X-Custom")
		return ctx.SendString(v)
	})

	req := httptest.NewRequest("GET", "/hdr", nil)
	req.Header.Set("X-Custom", "hello")
	resp, err := a.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(resp.Body)
	if string(body) != "hello" {
		t.Errorf("expected 'hello', got %q", string(body))
	}
}

// TestRequestBody verifies Body() reads the request body correctly.
func TestRequestBody(t *testing.T) {
	a := newTestApp()

	a.Post("/body", func(ctx fiber.Ctx) error {
		r := Upgrade(ctx)
		return ctx.SendString(r.Body())
	})

	req := httptest.NewRequest("POST", "/body", strings.NewReader("hello body"))
	resp, err := a.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(resp.Body)
	if string(body) != "hello body" {
		t.Errorf("expected 'hello body', got %q", string(body))
	}
}
