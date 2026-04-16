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

func parseProjectSort(ctx *app.RequestContext) ([]domain.Sort, error) {
	sortByStr := ctx.Query("sortBy")
	if sortByStr == "" {
		return nil, nil
	}
	by, ok := parseProjectSortBy(sortByStr)
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

// Handler exposes Project application service over HTTP.
type Handler struct {
	svc *application.Service
}

// NewHandler creates a Project HTTP handler.
func NewHandler(svc *application.Service) *Handler {
	return &Handler{svc: svc}
}

type createProjectReq struct {
	Name         string   `json:"name"`
	Goal         string   `json:"goal"`
	Principles   string   `json:"principles"`
	VisionResult string   `json:"visionResult"`
	Description  string   `json:"description"`
	Links        []string `json:"links"`
}

type updateProjectReq struct {
	Name         string   `json:"name"`
	Goal         string   `json:"goal"`
	Principles   string   `json:"principles"`
	VisionResult string   `json:"visionResult"`
	Description  string   `json:"description"`
	Links        []string `json:"links"`
}

type setStatusReq struct {
	Status string `json:"status"`
}

type linkResp struct {
	Label string `json:"label"`
	URL   string `json:"url"`
}

type projectResp struct {
	ID           string     `json:"id"`
	Name         string     `json:"name"`
	Goal         string     `json:"goal"`
	Principles   string     `json:"principles"`
	VisionResult string     `json:"visionResult"`
	Description  string     `json:"description"`
	Links        []linkResp `json:"links"`
	Status       string     `json:"status"`
	StartedAt    *time.Time `json:"startedAt,omitempty"`
	CompletedAt  *time.Time `json:"completedAt,omitempty"`
	CreatedAt    time.Time  `json:"createdAt"`
	UpdatedAt    time.Time  `json:"updatedAt"`
}

func toResp(p *domain.Project) projectResp {
	links := make([]linkResp, 0, len(p.Links))
	for _, link := range p.Links {
		links = append(links, linkResp{Label: link.Label, URL: link.URL})
	}
	return projectResp{
		ID:           p.ID,
		Name:         p.Name,
		Goal:         p.Goal,
		Principles:   p.Principles,
		VisionResult: p.VisionResult,
		Description:  p.Description,
		Links:        links,
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
		Links:        req.Links,
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
			writeErr(ctx, domain.InvalidStatus(statusStr))
			return
		}
		status = &s
	}

	q := domain.ListQuery{Status: status, Search: strings.TrimSpace(ctx.Query("search"))}
	sorts, err := parseProjectSort(ctx)
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
		Links:        req.Links,
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
		writeErr(ctx, domain.InvalidStatus(req.Status))
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
