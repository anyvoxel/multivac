// Package http provides HTTP handlers for Waiting List management.
package http

import (
	"context"
	"errors"
	stdhttp "net/http"
	"strconv"
	"strings"
	"time"

	"github.com/cloudwego/hertz/pkg/app"

	"github.com/anyvoxel/multivac/pkg/waitinglist/application"
	"github.com/anyvoxel/multivac/pkg/waitinglist/domain"
)

func parseWaitingListSortBy(v string) (domain.SortBy, bool) {
	switch strings.ToLower(v) {
	case "createdat", "created_at":
		return domain.WaitingListSortByCreatedAt, true
	case "updatedat", "updated_at":
		return domain.WaitingListSortByUpdatedAt, true
	case "name":
		return domain.WaitingListSortByName, true
	case "expectedat", "expected_at":
		return domain.WaitingListSortByExpectedAt, true
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

func parseWaitingListSort(ctx *app.RequestContext) ([]domain.Sort, error) {
	sortByStr := ctx.Query("sortBy")
	if sortByStr == "" {
		return nil, nil
	}
	by, ok := parseWaitingListSortBy(sortByStr)
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

// Handler exposes Waiting List application service over HTTP.
type Handler struct {
	svc *application.Service
}

// NewHandler creates a Waiting List HTTP handler.
func NewHandler(svc *application.Service) *Handler {
	return &Handler{svc: svc}
}

type createWaitingListReq struct {
	Name       string     `json:"name"`
	Details    string     `json:"details"`
	Owner      string     `json:"owner"`
	ExpectedAt *time.Time `json:"expectedAt"`
}

type updateWaitingListReq struct {
	Name       string     `json:"name"`
	Details    string     `json:"details"`
	Owner      string     `json:"owner"`
	ExpectedAt *time.Time `json:"expectedAt"`
}

type waitingListResp struct {
	ID         string     `json:"id"`
	Name       string     `json:"name"`
	Details    string     `json:"details"`
	Owner      string     `json:"owner"`
	ExpectedAt *time.Time `json:"expectedAt,omitempty"`
	CreatedAt  time.Time  `json:"createdAt"`
	UpdatedAt  time.Time  `json:"updatedAt"`
}

func toResp(item *domain.WaitingList) waitingListResp {
	return waitingListResp{
		ID:         item.ID,
		Name:       item.Name,
		Details:    item.Details,
		Owner:      item.Owner,
		ExpectedAt: item.ExpectedAt,
		CreatedAt:  item.CreatedAt,
		UpdatedAt:  item.UpdatedAt,
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

// Create handles POST /waiting-lists.
func (h *Handler) Create(c context.Context, ctx *app.RequestContext) {
	var req createWaitingListReq
	if err := ctx.BindJSON(&req); err != nil {
		ctx.JSON(stdhttp.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}
	item, err := h.svc.Create(c, application.CreateWaitingListCmd{
		Name:       req.Name,
		Details:    req.Details,
		Owner:      req.Owner,
		ExpectedAt: req.ExpectedAt,
	})
	if err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(stdhttp.StatusCreated, toResp(item))
}

// Get handles GET /waiting-lists/:id.
func (h *Handler) Get(c context.Context, ctx *app.RequestContext) {
	id := ctx.Param("id")
	item, err := h.svc.Get(c, id)
	if err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(stdhttp.StatusOK, toResp(item))
}

// List handles GET /waiting-lists.
func (h *Handler) List(c context.Context, ctx *app.RequestContext) {
	q := domain.ListQuery{Search: strings.TrimSpace(ctx.Query("search"))}
	sorts, err := parseWaitingListSort(ctx)
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
	out := make([]waitingListResp, 0, len(items))
	for _, item := range items {
		out = append(out, toResp(item))
	}
	ctx.JSON(stdhttp.StatusOK, out)
}

// Update handles PUT /waiting-lists/:id.
func (h *Handler) Update(c context.Context, ctx *app.RequestContext) {
	id := ctx.Param("id")
	var req updateWaitingListReq
	if err := ctx.BindJSON(&req); err != nil {
		ctx.JSON(stdhttp.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}
	item, err := h.svc.Update(c, id, application.UpdateWaitingListCmd{
		Name:       req.Name,
		Details:    req.Details,
		Owner:      req.Owner,
		ExpectedAt: req.ExpectedAt,
	})
	if err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(stdhttp.StatusOK, toResp(item))
}

// Delete handles DELETE /waiting-lists/:id.
func (h *Handler) Delete(c context.Context, ctx *app.RequestContext) {
	id := ctx.Param("id")
	if err := h.svc.Delete(c, id); err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.Status(stdhttp.StatusNoContent)
}
