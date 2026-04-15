// Package http provides HTTP handlers for Inbox management.
package http

import (
	"context"
	"errors"
	stdhttp "net/http"
	"strconv"
	"strings"
	"time"

	"github.com/cloudwego/hertz/pkg/app"

	"github.com/anyvoxel/multivac/pkg/inbox/application"
	"github.com/anyvoxel/multivac/pkg/inbox/domain"
)

func parseInboxSortBy(v string) (domain.SortBy, bool) {
	switch strings.ToLower(v) {
	case "createdat", "created_at":
		return domain.InboxSortByCreatedAt, true
	case "updatedat", "updated_at":
		return domain.InboxSortByUpdatedAt, true
	case "name":
		return domain.InboxSortByName, true
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

func parseInboxSort(ctx *app.RequestContext) ([]domain.Sort, error) {
	sortByStr := ctx.Query("sortBy")
	if sortByStr == "" {
		return nil, nil
	}
	by, ok := parseInboxSortBy(sortByStr)
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

// Handler exposes Inbox application service over HTTP.
type Handler struct {
	svc *application.Service
}

// NewHandler creates an Inbox HTTP handler.
func NewHandler(svc *application.Service) *Handler {
	return &Handler{svc: svc}
}

type createInboxReq struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type updateInboxReq struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type inboxResp struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

func toResp(inbox *domain.Inbox) inboxResp {
	return inboxResp{
		ID:          inbox.ID,
		Name:        inbox.Name,
		Description: inbox.Description,
		CreatedAt:   inbox.CreatedAt,
		UpdatedAt:   inbox.UpdatedAt,
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

// Create handles POST /inboxes.
func (h *Handler) Create(c context.Context, ctx *app.RequestContext) {
	var req createInboxReq
	if err := ctx.BindJSON(&req); err != nil {
		ctx.JSON(stdhttp.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}
	inbox, err := h.svc.Create(c, application.CreateInboxCmd{
		Name:        req.Name,
		Description: req.Description,
	})
	if err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(stdhttp.StatusCreated, toResp(inbox))
}

// Get handles GET /inboxes/:id.
func (h *Handler) Get(c context.Context, ctx *app.RequestContext) {
	id := ctx.Param("id")
	inbox, err := h.svc.Get(c, id)
	if err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(stdhttp.StatusOK, toResp(inbox))
}

// List handles GET /inboxes.
func (h *Handler) List(c context.Context, ctx *app.RequestContext) {
	q := domain.ListQuery{Search: strings.TrimSpace(ctx.Query("search"))}
	sorts, err := parseInboxSort(ctx)
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
	out := make([]inboxResp, 0, len(items))
	for _, inbox := range items {
		out = append(out, toResp(inbox))
	}
	ctx.JSON(stdhttp.StatusOK, out)
}

// Update handles PUT /inboxes/:id.
func (h *Handler) Update(c context.Context, ctx *app.RequestContext) {
	id := ctx.Param("id")
	var req updateInboxReq
	if err := ctx.BindJSON(&req); err != nil {
		ctx.JSON(stdhttp.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}
	inbox, err := h.svc.Update(c, id, application.UpdateInboxCmd{
		Name:        req.Name,
		Description: req.Description,
	})
	if err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(stdhttp.StatusOK, toResp(inbox))
}

// Delete handles DELETE /inboxes/:id.
func (h *Handler) Delete(c context.Context, ctx *app.RequestContext) {
	id := ctx.Param("id")
	if err := h.svc.Delete(c, id); err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.Status(stdhttp.StatusNoContent)
}
