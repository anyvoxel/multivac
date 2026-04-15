// Package http provides HTTP handlers for Someday management.
package http

import (
	"context"
	"errors"
	stdhttp "net/http"
	"strconv"
	"strings"
	"time"

	"github.com/cloudwego/hertz/pkg/app"

	"github.com/anyvoxel/multivac/pkg/someday/application"
	"github.com/anyvoxel/multivac/pkg/someday/domain"
)

func parseSomedaySortBy(v string) (domain.SortBy, bool) {
	switch strings.ToLower(v) {
	case "createdat", "created_at":
		return domain.SomedaySortByCreatedAt, true
	case "updatedat", "updated_at":
		return domain.SomedaySortByUpdatedAt, true
	case "name":
		return domain.SomedaySortByName, true
	default:
		return "", false
	}
}

func parsePagination(ctx *app.RequestContext) (limit, offset int, err error) {
	limitStr := ctx.Query("limit")
	if limitStr != "" {
		n, convErr := strconv.Atoi(limitStr)
		if convErr != nil || n < 0 {
			return 0, 0, domain.InvalidPaginationValue("limit", limitStr)
		}
		limit = n
	}
	offsetStr := ctx.Query("offset")
	if offsetStr != "" {
		n, convErr := strconv.Atoi(offsetStr)
		if convErr != nil || n < 0 {
			return 0, 0, domain.InvalidPaginationValue("offset", offsetStr)
		}
		offset = n
	}
	return limit, offset, nil
}

func parseSomedaySort(ctx *app.RequestContext) ([]domain.Sort, error) {
	sortByStr := ctx.Query("sortBy")
	if sortByStr == "" {
		return nil, nil
	}
	by, ok := parseSomedaySortBy(sortByStr)
	if !ok {
		return nil, domain.InvalidSortBy(sortByStr)
	}
	dir := domain.SortDesc
	if sortDirStr := ctx.Query("sortDir"); sortDirStr != "" {
		d, ok := domain.ParseSortDir(sortDirStr)
		if !ok {
			return nil, domain.InvalidSortDir(sortDirStr)
		}
		dir = d
	}
	return []domain.Sort{{By: by, Dir: dir}}, nil
}

// Handler exposes Someday application service over HTTP.
type Handler struct {
	svc *application.Service
}

// NewHandler creates a Someday HTTP handler.
func NewHandler(svc *application.Service) *Handler {
	return &Handler{svc: svc}
}

type createSomedayReq struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type updateSomedayReq struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type somedayResp struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

func toResp(someday *domain.Someday) somedayResp {
	return somedayResp{
		ID:          someday.ID,
		Name:        someday.Name,
		Description: someday.Description,
		CreatedAt:   someday.CreatedAt,
		UpdatedAt:   someday.UpdatedAt,
	}
}

func writeErr(ctx *app.RequestContext, err error) {
	if errors.Is(err, domain.ErrNotFound) {
		ctx.JSON(stdhttp.StatusNotFound, map[string]any{"error": err.Error()})
		return
	}
	if errors.Is(err, domain.ErrInvalidArg) {
		ctx.JSON(stdhttp.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}
	ctx.JSON(stdhttp.StatusInternalServerError, map[string]any{"error": err.Error()})
}

// Create handles POST /somedays.
func (h *Handler) Create(c context.Context, ctx *app.RequestContext) {
	var req createSomedayReq
	if err := ctx.BindJSON(&req); err != nil {
		ctx.JSON(stdhttp.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}
	someday, err := h.svc.Create(c, application.CreateSomedayCmd{
		Name:        req.Name,
		Description: req.Description,
	})
	if err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(stdhttp.StatusCreated, toResp(someday))
}

// Get handles GET /somedays/:id.
func (h *Handler) Get(c context.Context, ctx *app.RequestContext) {
	id := ctx.Param("id")
	someday, err := h.svc.Get(c, id)
	if err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(stdhttp.StatusOK, toResp(someday))
}

// List handles GET /somedays.
func (h *Handler) List(c context.Context, ctx *app.RequestContext) {
	q := domain.ListQuery{Search: strings.TrimSpace(ctx.Query("search"))}
	sorts, err := parseSomedaySort(ctx)
	if err != nil {
		writeErr(ctx, err)
		return
	}
	q.Sorts = sorts
	limit, offset, err := parsePagination(ctx)
	if err != nil {
		writeErr(ctx, err)
		return
	}
	q.Limit, q.Offset = limit, offset

	items, err := h.svc.List(c, q)
	if err != nil {
		writeErr(ctx, err)
		return
	}
	out := make([]somedayResp, 0, len(items))
	for _, someday := range items {
		out = append(out, toResp(someday))
	}
	ctx.JSON(stdhttp.StatusOK, out)
}

// Update handles PUT /somedays/:id.
func (h *Handler) Update(c context.Context, ctx *app.RequestContext) {
	id := ctx.Param("id")
	var req updateSomedayReq
	if err := ctx.BindJSON(&req); err != nil {
		ctx.JSON(stdhttp.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}
	someday, err := h.svc.Update(c, id, application.UpdateSomedayCmd{
		Name:        req.Name,
		Description: req.Description,
	})
	if err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(stdhttp.StatusOK, toResp(someday))
}

// Delete handles DELETE /somedays/:id.
func (h *Handler) Delete(c context.Context, ctx *app.RequestContext) {
	id := ctx.Param("id")
	if err := h.svc.Delete(c, id); err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.Status(stdhttp.StatusNoContent)
}
