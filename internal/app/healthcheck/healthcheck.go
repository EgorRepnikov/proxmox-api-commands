package healthcheck

import (
	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"
)

func InitHealthcheck(route *router.Router) {
	healthHandler := func(ctx *fasthttp.RequestCtx) {
		ctx.SetStatusCode(200)
	}
	route.GET("/health", healthHandler)
}
