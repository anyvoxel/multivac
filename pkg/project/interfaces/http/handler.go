// Package http provides HTTP handlers for Project management.
package http

import (
	"context"
	"errors"
	stdhttp "net/http"
	"strconv"
	"strings"
	"time"

	"github.com/cloudwego/hertz/pkg/app"

	"github.com/anyvoxel/multivac/pkg/project/application"
	"github.com/anyvoxel/multivac/pkg/project/domain"
)

func parseProjectSortBy(v string) (domain.SortBy, bool) {
	switch strings.ToLower(v) {
	case "createdat", "created_at":
		return domain.ProjectSortByCreatedAt, true
	case "updatedat", "updated_at":
		return domain.ProjectSortByUpdatedAt, true
	case "name":
		return domain.ProjectSortByName, true
	default:
		return "", false
	}
}

func parsePagination(ctx *app.RequestContext) (limit, offset int, ok bool) {
	limitStr := ctx.Query("limit")
	if limitStr != "" {
		n, err := strconv.Atoi(limitStr)
		if err != nil || n < 0 {
			return 0, 0, false
		}
		limit = n
	}
	offsetStr := ctx.Query("offset")
	if offsetStr != "" {
		n, err := strconv.Atoi(offsetStr)
		if err != nil || n < 0 {
			return 0, 0, false
		}
		offset = n
	}
	return limit, offset, true
}

func parseProjectSort(ctx *app.RequestContext) ([]domain.Sort, bool) {
	sortByStr := ctx.Query("sortBy")
	if sortByStr == "" {
		return nil, true
	}
	by, ok := parseProjectSortBy(sortByStr)
	if !ok {
		return nil, false
	}
	dir := domain.SortDesc
	if sortDirStr := ctx.Query("sortDir"); sortDirStr != "" {
		d, ok := domain.ParseSortDir(sortDirStr)
		if !ok {
			return nil, false
		}
		dir = d
	}
	return []domain.Sort{{By: by, Dir: dir}}, true
}

// Handler exposes Project application service over HTTP.
type Handler struct {
	svc *application.Service
}

// NewHandler creates a Project HTTP handler.
func NewHandler(svc *application.Service) *Handler {
	return &Handler{svc: svc}
}

type createProjectReq struct {
	Name         string `json:"name"`
	Goal         string `json:"goal"`
	Principles   string `json:"principles"`
	VisionResult string `json:"visionResult"`
	Description  string `json:"description"`
}

type updateProjectReq struct {
	Name         string `json:"name"`
	Goal         string `json:"goal"`
	Principles   string `json:"principles"`
	VisionResult string `json:"visionResult"`
	Description  string `json:"description"`
}

type setStatusReq struct {
	Status string `json:"status"`
}

type projectResp struct {
	ID           string     `json:"id"`
	Name         string     `json:"name"`
	Goal         string     `json:"goal"`
	Principles   string     `json:"principles"`
	VisionResult string     `json:"visionResult"`
	Description  string     `json:"description"`
	Status       string     `json:"status"`
	StartedAt    *time.Time `json:"startedAt,omitempty"`
	CompletedAt  *time.Time `json:"completedAt,omitempty"`
	CreatedAt    time.Time  `json:"createdAt"`
	UpdatedAt    time.Time  `json:"updatedAt"`
}

func toResp(p *domain.Project) projectResp {
	return projectResp{
		ID:           p.ID,
		Name:         p.Name,
		Goal:         p.Goal,
		Principles:   p.Principles,
		VisionResult: p.VisionResult,
		Description:  p.Description,
		Status:       string(p.Status),
		StartedAt:    p.StartedAt,
		CompletedAt:  p.CompletedAt,
		CreatedAt:    p.CreatedAt,
		UpdatedAt:    p.UpdatedAt,
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

// Create handles POST /projects.
func (h *Handler) Create(c context.Context, ctx *app.RequestContext) {
	var req createProjectReq
	if err := ctx.BindJSON(&req); err != nil {
		ctx.JSON(stdhttp.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}
	p, err := h.svc.Create(c, application.CreateProjectCmd{
		Name:         req.Name,
		Goal:         req.Goal,
		Principles:   req.Principles,
		VisionResult: req.VisionResult,
		Description:  req.Description,
	})
	if err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(stdhttp.StatusCreated, toResp(p))
}

// Get handles GET /projects/:id.
func (h *Handler) Get(c context.Context, ctx *app.RequestContext) {
	id := ctx.Param("id")
	p, err := h.svc.Get(c, id)
	if err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(stdhttp.StatusOK, toResp(p))
}

// List handles GET /projects.
func (h *Handler) List(c context.Context, ctx *app.RequestContext) {
	statusStr := ctx.Query("status")
	var status *domain.Status
	if statusStr != "" {
		s, ok := domain.ParseStatus(statusStr)
		if !ok {
			ctx.JSON(stdhttp.StatusBadRequest, map[string]any{"error": "invalid status"})
			return
		}
		status = &s
	}

	q := domain.ListQuery{Status: status}
	if sorts, ok := parseProjectSort(ctx); ok {
		q.Sorts = sorts
	} else {
		ctx.JSON(stdhttp.StatusBadRequest, map[string]any{"error": "invalid sort"})
		return
	}
	limit, offset, ok := parsePagination(ctx)
	if !ok {
		ctx.JSON(stdhttp.StatusBadRequest, map[string]any{"error": "invalid pagination"})
		return
	}
	q.Limit, q.Offset = limit, offset

	items, err := h.svc.List(c, q)
	if err != nil {
		writeErr(ctx, err)
		return
	}
	out := make([]projectResp, 0, len(items))
	for _, p := range items {
		out = append(out, toResp(p))
	}
	ctx.JSON(stdhttp.StatusOK, out)
}

// Update handles PUT /projects/:id.
func (h *Handler) Update(c context.Context, ctx *app.RequestContext) {
	id := ctx.Param("id")
	var req updateProjectReq
	if err := ctx.BindJSON(&req); err != nil {
		ctx.JSON(stdhttp.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}
	p, err := h.svc.Update(c, id, application.UpdateProjectCmd{
		Name:         req.Name,
		Goal:         req.Goal,
		Principles:   req.Principles,
		VisionResult: req.VisionResult,
		Description:  req.Description,
	})
	if err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(stdhttp.StatusOK, toResp(p))
}

// SetStatus handles PATCH /projects/:id/status.
func (h *Handler) SetStatus(c context.Context, ctx *app.RequestContext) {
	id := ctx.Param("id")
	var req setStatusReq
	if err := ctx.BindJSON(&req); err != nil {
		ctx.JSON(stdhttp.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}
	st, ok := domain.ParseStatus(req.Status)
	if !ok {
		ctx.JSON(stdhttp.StatusBadRequest, map[string]any{"error": "invalid status"})
		return
	}
	p, err := h.svc.SetStatus(c, id, st)
	if err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(stdhttp.StatusOK, toResp(p))
}

// Delete handles DELETE /projects/:id.
func (h *Handler) Delete(c context.Context, ctx *app.RequestContext) {
	id := ctx.Param("id")
	if err := h.svc.Delete(c, id); err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.Status(stdhttp.StatusNoContent)
}
