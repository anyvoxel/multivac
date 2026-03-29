// Package main starts the multivac HTTP server.
package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"

	"github.com/anyvoxel/multivac/pkg/utils/version"
)

func main() {
	addr := os.Getenv("HTTP_ADDR")
	if addr == "" {
		addr = ":8080"
	}

	h := server.New(server.WithHostPorts(addr))

	h.GET("/", func(_ context.Context, ctx *app.RequestContext) {
		ctx.String(http.StatusOK, "multivac\n")
	})
	h.GET("/healthz", func(_ context.Context, ctx *app.RequestContext) {
		ctx.String(http.StatusOK, "ok\n")
	})
	h.GET("/version", func(_ context.Context, ctx *app.RequestContext) {
		b, err := json.Marshal(version.Info())
		if err != nil {
			ctx.String(http.StatusInternalServerError, "marshal version: %v\n", err)
			return
		}
		ctx.Data(http.StatusOK, "application/json; charset=utf-8", b)
	})

	log.Printf("http server listening on %s", addr)
	h.Spin()
}
