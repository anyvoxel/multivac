// Package http provides HTTP handlers for Task management.
package http

import (
	"context"
	"errors"
	stdhttp "net/http"
	"strconv"
	"strings"
	"time"

	"github.com/cloudwego/hertz/pkg/app"

	"github.com/anyvoxel/multivac/pkg/task/application"
	"github.com/anyvoxel/multivac/pkg/task/domain"
)

func parseTaskSortBy(v string) (domain.SortBy, bool) {
	switch strings.ToLower(v) {
	case "createdat", "created_at":
		return domain.TaskSortByCreatedAt, true
	case "updatedat", "updated_at":
		return domain.TaskSortByUpdatedAt, true
	case "dueat", "due_at":
		return domain.TaskSortByDueAt, true
	case "priority":
		return domain.TaskSortByPriority, true
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

func parseTaskSort(ctx *app.RequestContext) ([]domain.Sort, error) {
	sortByStr := ctx.Query("sortBy")
	if sortByStr == "" {
		return nil, nil
	}
	by, ok := parseTaskSortBy(sortByStr)
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

// Handler exposes Task application service over HTTP.
type Handler struct {
	svc *application.Service
}

// NewHandler creates a Task HTTP handler.
func NewHandler(svc *application.Service) *Handler {
	return &Handler{svc: svc}
}

type createTaskReq struct {
	ProjectID    *string `json:"projectId,omitempty"`
	Name         string  `json:"name"`
	Description  string  `json:"description"`
	Context      string  `json:"context"`
	Details      string  `json:"details"`
	Status       string  `json:"status,omitempty"`
	Priority     string  `json:"priority"`
	DueAtRFC3339 *string `json:"dueAt,omitempty"`
}

type updateTaskReq struct {
	ProjectID    *string `json:"projectId,omitempty"`
	Name         string  `json:"name"`
	Description  string  `json:"description"`
	Context      string  `json:"context"`
	Details      string  `json:"details"`
	Priority     string  `json:"priority"`
	DueAtRFC3339 *string `json:"dueAt,omitempty"`
}

type setStatusReq struct {
	Status string `json:"status"`
}

type taskResp struct {
	ID          string     `json:"id"`
	ProjectID   string     `json:"projectId"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Context     string     `json:"context"`
	Details     string     `json:"details"`
	Status      string     `json:"status"`
	Priority    string     `json:"priority"`
	DueAt       *time.Time `json:"dueAt,omitempty"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
}

func toResp(t *domain.Task) taskResp {
	return taskResp{
		ID:          t.ID,
		ProjectID:   t.ProjectID,
		Name:        t.Name,
		Description: t.Description,
		Context:     t.Context,
		Details:     t.Details,
		Status:      string(t.Status),
		Priority:    string(t.Priority),
		DueAt:       t.DueAt,
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,
	}
}

func parseDueAt(s *string) (*time.Time, error) {
	if s == nil {
		return nil, nil
	}
	if *s == "" {
		return nil, nil
	}
	t, err := time.Parse(time.RFC3339, *s)
	if err != nil {
		return nil, domain.ErrInvalidArg
	}
	return &t, nil
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

// List handles GET /tasks.
func (h *Handler) List(c context.Context, ctx *app.RequestContext) {
	projectID := ctx.Query("projectId")
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

	q := domain.ListQuery{ProjectID: projectID, Status: status, Search: strings.TrimSpace(ctx.Query("search"))}
	sorts, err := parseTaskSort(ctx)
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
	out := make([]taskResp, 0, len(items))
	for _, t := range items {
		out = append(out, toResp(t))
	}
	ctx.JSON(stdhttp.StatusOK, out)
}

// Create handles POST /tasks.
func (h *Handler) Create(c context.Context, ctx *app.RequestContext) {
	var req createTaskReq
	if err := ctx.BindJSON(&req); err != nil {
		ctx.JSON(stdhttp.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}
	prio, ok := domain.ParsePriority(req.Priority)
	if !ok {
		writeErr(ctx, domain.InvalidPriority(req.Priority))
		return
	}
	dueAt, err := parseDueAt(req.DueAtRFC3339)
	if err != nil {
		writeErr(ctx, err)
		return
	}

	projectID := ""
	if req.ProjectID != nil {
		projectID = strings.TrimSpace(*req.ProjectID)
	}

	t, err := h.svc.Create(c, application.CreateTaskCmd{
		ProjectID:   projectID,
		Name:        req.Name,
		Description: req.Description,
		Context:     req.Context,
		Details:     req.Details,
		Priority:    prio,
		DueAt:       dueAt,
	})
	if err != nil {
		writeErr(ctx, err)
		return
	}
	if req.Status != "" {
		st, ok := domain.ParseStatus(req.Status)
		if !ok {
			writeErr(ctx, domain.InvalidStatus(req.Status))
			return
		}
		t, err = h.svc.SetStatus(c, t.ID, st)
		if err != nil {
			writeErr(ctx, err)
			return
		}
	}
	ctx.JSON(stdhttp.StatusCreated, toResp(t))
}

// Get handles GET /tasks/:id.
func (h *Handler) Get(c context.Context, ctx *app.RequestContext) {
	id := ctx.Param("id")
	t, err := h.svc.Get(c, id)
	if err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(stdhttp.StatusOK, toResp(t))
}

// ListByProject handles GET /projects/:projectId/tasks.
func (h *Handler) ListByProject(c context.Context, ctx *app.RequestContext) {
	pid := ctx.Param("projectId")
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

	q := domain.ListQuery{ProjectID: pid, Status: status, Search: strings.TrimSpace(ctx.Query("search"))}
	sorts, err := parseTaskSort(ctx)
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
	out := make([]taskResp, 0, len(items))
	for _, t := range items {
		out = append(out, toResp(t))
	}
	ctx.JSON(stdhttp.StatusOK, out)
}

// Update handles PUT /tasks/:id.
func (h *Handler) Update(c context.Context, ctx *app.RequestContext) {
	id := ctx.Param("id")
	var req updateTaskReq
	if err := ctx.BindJSON(&req); err != nil {
		ctx.JSON(stdhttp.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}
	prio, ok := domain.ParsePriority(req.Priority)
	if !ok {
		writeErr(ctx, domain.InvalidPriority(req.Priority))
		return
	}
	dueAt, err := parseDueAt(req.DueAtRFC3339)
	if err != nil {
		writeErr(ctx, err)
		return
	}

	projectID := ""
	if req.ProjectID != nil {
		projectID = strings.TrimSpace(*req.ProjectID)
	}

	t, err := h.svc.Update(c, id, application.UpdateTaskCmd{
		ProjectID:   projectID,
		Name:        req.Name,
		Description: req.Description,
		Context:     req.Context,
		Details:     req.Details,
		Priority:    prio,
		DueAt:       dueAt,
	})
	if err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(stdhttp.StatusOK, toResp(t))
}

// SetStatus handles PATCH /tasks/:id/status.
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
	t, err := h.svc.SetStatus(c, id, st)
	if err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(stdhttp.StatusOK, toResp(t))
}

// Delete handles DELETE /tasks/:id.
func (h *Handler) Delete(c context.Context, ctx *app.RequestContext) {
	id := ctx.Param("id")
	if err := h.svc.Delete(c, id); err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.Status(stdhttp.StatusNoContent)
}
