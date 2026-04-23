// Package main starts the multivac HTTP server.
package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/jmoiron/sqlx"

	"github.com/anyvoxel/multivac/internal/webui"
	contextapp "github.com/anyvoxel/multivac/pkg/context/application"
	contextsqlite "github.com/anyvoxel/multivac/pkg/context/infra/sqlite"
	contexthttp "github.com/anyvoxel/multivac/pkg/context/interfaces/http"
	inboxapp "github.com/anyvoxel/multivac/pkg/inbox/application"
	inboxsqlite "github.com/anyvoxel/multivac/pkg/inbox/infra/sqlite"
	inboxhttp "github.com/anyvoxel/multivac/pkg/inbox/interfaces/http"
	"github.com/anyvoxel/multivac/pkg/project/application"
	"github.com/anyvoxel/multivac/pkg/project/infra/sqlite"
	projecthttp "github.com/anyvoxel/multivac/pkg/project/interfaces/http"
	referenceapp "github.com/anyvoxel/multivac/pkg/reference/application"
	referencesqlite "github.com/anyvoxel/multivac/pkg/reference/infra/sqlite"
	referencehttp "github.com/anyvoxel/multivac/pkg/reference/interfaces/http"
	somedayapp "github.com/anyvoxel/multivac/pkg/someday/application"
	somedaysqlite "github.com/anyvoxel/multivac/pkg/someday/infra/sqlite"
	somedayhttp "github.com/anyvoxel/multivac/pkg/someday/interfaces/http"
	"github.com/anyvoxel/multivac/pkg/utils/version"
)

func main() {
	addr := os.Getenv("HTTP_ADDR")
	if addr == "" {
		addr = ":8080"
	}
	dbPath := os.Getenv("SQLITE_PATH")
	if dbPath == "" {
		dbPath = "./multivac.sqlite"
	}

	if err := run(addr, dbPath); err != nil {
		log.Fatalf("run server: %v", err)
	}
}

func run(addr, dbPath string) error {
	db, projHandler, inboxHandler, contextHandler, referenceHandler, somedayHandler, err := setupServices(dbPath)
	if err != nil {
		return err
	}
	defer func() { _ = db.Close() }()

	h := server.New(server.WithHostPorts(addr))

	// Simple CORS for local web development.
	h.Use(func(c context.Context, ctx *app.RequestContext) {
		ctx.Header("Access-Control-Allow-Origin", "*")
		ctx.Header("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
		ctx.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		ctx.Header("Access-Control-Max-Age", "86400")
		if string(ctx.Method()) == http.MethodOptions {
			ctx.Status(http.StatusNoContent)
			ctx.Abort()
			return
		}
		ctx.Next(c)
	})

	h.GET("/", func(_ context.Context, ctx *app.RequestContext) {
		ctx.Redirect(http.StatusFound, []byte("/index.html"))
		ctx.Abort()
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

	api := h.Group("/api/v1")
	api.POST("/projects", projHandler.Create)
	api.GET("/projects", projHandler.List)
	api.GET("/projects/:id", projHandler.Get)
	api.PUT("/projects/:id", projHandler.Update)
	api.PATCH("/projects/:id/status", projHandler.SetStatus)
	api.DELETE("/projects/:id", projHandler.Delete)

	api.POST("/inboxes", inboxHandler.Create)
	api.GET("/inboxes", inboxHandler.List)
	api.GET("/inboxes/:id", inboxHandler.Get)
	api.PUT("/inboxes/:id", inboxHandler.Update)
	api.DELETE("/inboxes/:id", inboxHandler.Delete)

	api.POST("/contexts", contextHandler.Create)
	api.GET("/contexts", contextHandler.List)
	api.GET("/contexts/:id", contextHandler.Get)
	api.PUT("/contexts/:id", contextHandler.Update)
	api.DELETE("/contexts/:id", contextHandler.Delete)

	api.POST("/references", referenceHandler.Create)
	api.GET("/references", referenceHandler.List)
	api.GET("/references/:id", referenceHandler.Get)
	api.PUT("/references/:id", referenceHandler.Update)
	api.DELETE("/references/:id", referenceHandler.Delete)

	api.POST("/somedays", somedayHandler.Create)
	api.GET("/somedays", somedayHandler.List)
	api.GET("/somedays/:id", somedayHandler.Get)
	api.PUT("/somedays/:id", somedayHandler.Update)
	api.DELETE("/somedays/:id", somedayHandler.Delete)

	api.POST("/inboxes/:id/convert-to-someday", somedayHandler.ConvertFromInbox)

	// Web UI (embedded). It serves SPA under /.
	h.GET("/*filepath", func(_ context.Context, ctx *app.RequestContext) {
		p := ctx.Param("filepath")
		if p == "" || p == "/" {
			ctx.Redirect(http.StatusFound, []byte("/index.html"))
			ctx.Abort()
			return
		}
		if !strings.HasPrefix(p, "/") {
			p = "/" + p
		}
		if strings.HasPrefix(p, "/api/") {
			ctx.Status(http.StatusNotFound)
			return
		}
		asset, err := webui.ReadAsset(p)
		if err != nil {
			ctx.Status(http.StatusNotFound)
			return
		}
		ctx.Header("Content-Type", asset.ContentType)
		if asset.Immutable {
			ctx.Header("Cache-Control", "public, max-age=31536000, immutable")
		}
		ctx.Data(http.StatusOK, asset.ContentType, asset.Body)
	})

	log.Printf("http server listening on %s", addr)
	h.Spin()
	return nil
}

func setupServices(dbPath string) (*sqlx.DB, *projecthttp.Handler, *inboxhttp.Handler, *contexthttp.Handler, *referencehttp.Handler, *somedayhttp.Handler, error) {
	db, err := sqlx.Open("sqlite3", dbPath)
	if err != nil {
		return nil, nil, nil, nil, nil, nil, err
	}
	db.SetMaxOpenConns(1)
	db.SetConnMaxLifetime(time.Minute)
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, nil, nil, nil, nil, nil, err
	}
	if _, err := db.Exec("PRAGMA foreign_keys = ON;"); err != nil {
		_ = db.Close()
		return nil, nil, nil, nil, nil, nil, err
	}

	projRepo := sqlite.NewRepository(db)
	projSvc := application.NewService(projRepo)
	if err := projSvc.Migrate(context.Background()); err != nil {
		_ = db.Close()
		return nil, nil, nil, nil, nil, nil, err
	}
	projHandler := projecthttp.NewHandler(projSvc)

	inboxRepo := inboxsqlite.NewRepository(db)
	inboxSvc := inboxapp.NewService(inboxRepo)
	if err := inboxSvc.Migrate(context.Background()); err != nil {
		_ = db.Close()
		return nil, nil, nil, nil, nil, nil, err
	}
	inboxHandler := inboxhttp.NewHandler(inboxSvc)

	contextRepo := contextsqlite.NewRepository(db)
	contextSvc := contextapp.NewService(contextRepo)
	if err := contextSvc.Migrate(context.Background()); err != nil {
		_ = db.Close()
		return nil, nil, nil, nil, nil, nil, err
	}
	contextHandler := contexthttp.NewHandler(contextSvc)

	referenceRepo := referencesqlite.NewRepository(db)
	referenceSvc := referenceapp.NewService(referenceRepo)
	if err := referenceSvc.Migrate(context.Background()); err != nil {
		_ = db.Close()
		return nil, nil, nil, nil, nil, nil, err
	}
	referenceHandler := referencehttp.NewHandler(referenceSvc)

	somedayRepo := somedaysqlite.NewRepository(db)
	somedaySvc := somedayapp.NewService(somedayRepo)
	if err := somedaySvc.Migrate(context.Background()); err != nil {
		_ = db.Close()
		return nil, nil, nil, nil, nil, nil, err
	}
	somedayHandler := somedayhttp.NewHandler(somedaySvc)

	return db, projHandler, inboxHandler, contextHandler, referenceHandler, somedayHandler, nil
}
