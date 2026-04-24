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
	actionapp "github.com/anyvoxel/multivac/pkg/application"
	actionsqlite "github.com/anyvoxel/multivac/pkg/infra/sqlite"
	actionhttp "github.com/anyvoxel/multivac/pkg/interfaces/http"
	contextapp "github.com/anyvoxel/multivac/pkg/application"
	contextsqlite "github.com/anyvoxel/multivac/pkg/infra/sqlite"
	contexthttp "github.com/anyvoxel/multivac/pkg/interfaces/http"
	inboxapp "github.com/anyvoxel/multivac/pkg/application"
	inboxsqlite "github.com/anyvoxel/multivac/pkg/infra/sqlite"
	inboxhttp "github.com/anyvoxel/multivac/pkg/interfaces/http"
	projectapp "github.com/anyvoxel/multivac/pkg/application"
	projectsqlite "github.com/anyvoxel/multivac/pkg/infra/sqlite"
	projecthttp "github.com/anyvoxel/multivac/pkg/interfaces/http"
	referenceapp "github.com/anyvoxel/multivac/pkg/application"
	referencesqlite "github.com/anyvoxel/multivac/pkg/infra/sqlite"
	referencehttp "github.com/anyvoxel/multivac/pkg/interfaces/http"
	somedayapp "github.com/anyvoxel/multivac/pkg/application"
	somedaysqlite "github.com/anyvoxel/multivac/pkg/infra/sqlite"
	somedayhttp "github.com/anyvoxel/multivac/pkg/interfaces/http"
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
	db, projHandler, inboxHandler, actionHandler, contextHandler, referenceHandler, somedayHandler, err := setupServices(dbPath)
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

	api.POST("/actions", actionHandler.Create)
	api.GET("/actions", actionHandler.List)
	api.GET("/actions/:id", actionHandler.Get)
	api.PUT("/actions/:id", actionHandler.Update)
	api.DELETE("/actions/:id", actionHandler.Delete)

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

	api.POST("/inboxes/:id/convert-to-action", actionHandler.ConvertFromInbox)
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

func setupServices(dbPath string) (*sqlx.DB, *projecthttp.ProjectHandler, *inboxhttp.InboxHandler, *actionhttp.ActionHandler, *contexthttp.ContextHandler, *referencehttp.ReferenceHandler, *somedayhttp.SomedayHandler, error) {
	db, err := sqlx.Open("sqlite3", dbPath)
	if err != nil {
		return nil, nil, nil, nil, nil, nil, nil, err
	}
	db.SetMaxOpenConns(1)
	db.SetConnMaxLifetime(time.Minute)
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, nil, nil, nil, nil, nil, nil, err
	}

	projRepo := projectsqlite.NewProjectRepository(db)
	projSvc := projectapp.NewProjectService(projRepo)
	if err := projSvc.Migrate(context.Background()); err != nil {
		_ = db.Close()
		return nil, nil, nil, nil, nil, nil, nil, err
	}
	projHandler := projecthttp.NewProjectHandler(projSvc)

	inboxRepo := inboxsqlite.NewInboxRepository(db)
	inboxSvc := inboxapp.NewInboxService(inboxRepo)
	if err := inboxSvc.Migrate(context.Background()); err != nil {
		_ = db.Close()
		return nil, nil, nil, nil, nil, nil, nil, err
	}
	inboxHandler := inboxhttp.NewInboxHandler(inboxSvc)

	actionRepo := actionsqlite.NewActionRepository(db)
	actionSvc := actionapp.NewActionService(actionRepo)
	if err := actionSvc.Migrate(context.Background()); err != nil {
		_ = db.Close()
		return nil, nil, nil, nil, nil, nil, nil, err
	}
	actionHandler := actionhttp.NewActionHandler(actionSvc)

	contextRepo := contextsqlite.NewContextRepository(db)
	contextSvc := contextapp.NewContextService(contextRepo)
	if err := contextSvc.Migrate(context.Background()); err != nil {
		_ = db.Close()
		return nil, nil, nil, nil, nil, nil, nil, err
	}
	contextHandler := contexthttp.NewContextHandler(contextSvc)

	referenceRepo := referencesqlite.NewReferenceRepository(db)
	referenceSvc := referenceapp.NewReferenceService(referenceRepo)
	if err := referenceSvc.Migrate(context.Background()); err != nil {
		_ = db.Close()
		return nil, nil, nil, nil, nil, nil, nil, err
	}
	referenceHandler := referencehttp.NewReferenceHandler(referenceSvc)

	somedayRepo := somedaysqlite.NewSomedayRepository(db)
	somedaySvc := somedayapp.NewSomedayService(somedayRepo)
	if err := somedaySvc.Migrate(context.Background()); err != nil {
		_ = db.Close()
		return nil, nil, nil, nil, nil, nil, nil, err
	}
	somedayHandler := somedayhttp.NewSomedayHandler(somedaySvc)

	return db, projHandler, inboxHandler, actionHandler, contextHandler, referenceHandler, somedayHandler, nil
}
