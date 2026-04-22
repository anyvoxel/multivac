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
	case "title":
		return domain.ProjectSortByTitle, true
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

type goalReq struct {
	Title       string     `json:"title"`
	CreatedAt   *time.Time `json:"created_at,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

type referenceReq struct {
	Title string `json:"title"`
	URL   string `json:"URL"`
}

type createProjectReq struct {
	Title       string         `json:"title"`
	Goals       []goalReq      `json:"goals"`
	Description string         `json:"description"`
	References  []referenceReq `json:"references"`
}

type updateProjectReq struct {
	Title       string         `json:"title"`
	Goals       []goalReq      `json:"goals"`
	Description string         `json:"description"`
	References  []referenceReq `json:"references"`
}

type setStatusReq struct {
	Status string `json:"status"`
}

type goalResp struct {
	Title       string     `json:"title"`
	CreatedAt   time.Time  `json:"created_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

type referenceResp struct {
	Title string `json:"title"`
	URL   string `json:"URL"`
}

type projectResp struct {
	ID          string          `json:"id"`
	Title       string          `json:"title"`
	Goals       []goalResp      `json:"goals"`
	Description string          `json:"description"`
	References  []referenceResp `json:"references"`
	Status      string          `json:"status"`
	StartAt     *time.Time      `json:"startAt,omitempty"`
	CompletedAt *time.Time      `json:"completedAt,omitempty"`
	CreatedAt   time.Time       `json:"createdAt"`
	UpdatedAt   time.Time       `json:"updatedAt"`
}

func toDomainGoals(goals []goalReq) []domain.Goal {
	if len(goals) == 0 {
		return nil
	}
	out := make([]domain.Goal, 0, len(goals))
	for _, goal := range goals {
		value := strings.TrimSpace(goal.Title)
		if value == "" {
			continue
		}
		var createdAt time.Time
		if goal.CreatedAt != nil {
			createdAt = *goal.CreatedAt
		}
		out = append(out, domain.Goal{
			Title:       value,
			CreatedAt:   createdAt,
			CompletedAt: goal.CompletedAt,
		})
	}
	return out
}

func toGoalResp(goals []domain.Goal) []goalResp {
	if len(goals) == 0 {
		return []goalResp{}
	}
	out := make([]goalResp, 0, len(goals))
	for _, goal := range goals {
		out = append(out, goalResp{
			Title:       goal.Title,
			CreatedAt:   goal.CreatedAt,
			CompletedAt: goal.CompletedAt,
		})
	}
	return out
}

func toDomainReferences(references []referenceReq) []domain.Reference {
	if len(references) == 0 {
		return nil
	}
	out := make([]domain.Reference, 0, len(references))
	for _, reference := range references {
		out = append(out, domain.Reference{
			Title: strings.TrimSpace(reference.Title),
			URL:   strings.TrimSpace(reference.URL),
		})
	}
	return out
}

func toReferenceResp(references []domain.Reference) []referenceResp {
	if len(references) == 0 {
		return []referenceResp{}
	}
	out := make([]referenceResp, 0, len(references))
	for _, reference := range references {
		out = append(out, referenceResp{Title: reference.Title, URL: reference.URL})
	}
	return out
}

func toResp(p *domain.Project) projectResp {
	return projectResp{
		ID:          p.ID,
		Title:       p.Title,
		Goals:       toGoalResp(p.Goals),
		Description: p.Description,
		References:  toReferenceResp(p.References),
		Status:      string(p.Status),
		StartAt:     p.StartAt,
		CompletedAt: p.CompletedAt,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
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
		Title:       req.Title,
		Goals:       toDomainGoals(req.Goals),
		Description: req.Description,
		References:  toDomainReferences(req.References),
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

	q := domain.ListQuery{
		Status: status,
		Search: strings.TrimSpace(ctx.Query("search")),
	}
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
		Title:       req.Title,
		Goals:       toDomainGoals(req.Goals),
		Description: req.Description,
		References:  toDomainReferences(req.References),
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
