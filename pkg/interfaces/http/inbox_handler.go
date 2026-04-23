package http

import (
	"context"
	"errors"
	stdhttp "net/http"
	"strconv"
	"strings"
	"time"

	"github.com/cloudwego/hertz/pkg/app"

	"github.com/anyvoxel/multivac/pkg/application"
	"github.com/anyvoxel/multivac/pkg/domain"
)

func parseInboxPagination(ctx *app.RequestContext) (limit, offset int, err error) {
	if v := ctx.Query("limit"); v != "" {
		n, convErr := strconv.Atoi(v)
		if convErr != nil || n < 0 {
			return 0, 0, domain.InvalidPaginationValue("limit", v)
		}
		limit = n
	}
	if v := ctx.Query("offset"); v != "" {
		n, convErr := strconv.Atoi(v)
		if convErr != nil || n < 0 {
			return 0, 0, domain.InvalidPaginationValue("offset", v)
		}
		offset = n
	}
	return limit, offset, nil
}

func parseInboxSortBy(v string) (domain.SortBy, bool) {
	switch strings.ToLower(v) {
	case "createdat", "created_at":
		return domain.InboxSortByCreatedAt, true
	case "updatedat", "updated_at":
		return domain.InboxSortByUpdatedAt, true
	case "name", "title":
		return domain.InboxSortByTitle, true
	default:
		return "", false
	}
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
	if v := ctx.Query("sortDir"); v != "" {
		parsed, ok := domain.ParseSortDir(v)
		if !ok {
			return nil, domain.InvalidSortDir(v)
		}
		dir = parsed
	}
	return []domain.Sort{{By: by, Dir: dir}}, nil
}

type InboxHandler struct{ svc *application.InboxService }

func NewInboxHandler(svc *application.InboxService) *InboxHandler { return &InboxHandler{svc: svc} }

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

func toInboxResp(inbox *domain.Inbox) inboxResp {
	return inboxResp{ID: inbox.ID, Name: inbox.Title, Description: inbox.Description, CreatedAt: inbox.CreatedAt, UpdatedAt: inbox.UpdatedAt}
}

func writeInboxErr(ctx *app.RequestContext, err error) {
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

func (h *InboxHandler) Create(c context.Context, ctx *app.RequestContext) {
	var req createInboxReq
	if err := ctx.BindJSON(&req); err != nil {
		ctx.JSON(stdhttp.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}
	inbox, err := h.svc.Create(c, application.CreateInboxCmd{Title: req.Name, Description: req.Description})
	if err != nil {
		writeInboxErr(ctx, err)
		return
	}
	ctx.JSON(stdhttp.StatusCreated, toInboxResp(inbox))
}

func (h *InboxHandler) Get(c context.Context, ctx *app.RequestContext) {
	inbox, err := h.svc.Get(c, ctx.Param("id"))
	if err != nil {
		writeInboxErr(ctx, err)
		return
	}
	ctx.JSON(stdhttp.StatusOK, toInboxResp(inbox))
}

func (h *InboxHandler) List(c context.Context, ctx *app.RequestContext) {
	q := domain.InboxListQuery{Search: strings.TrimSpace(ctx.Query("search"))}
	sorts, err := parseInboxSort(ctx)
	if err != nil {
		writeInboxErr(ctx, err)
		return
	}
	q.Sorts = sorts
	limit, offset, err := parseInboxPagination(ctx)
	if err != nil {
		writeInboxErr(ctx, err)
		return
	}
	q.Limit = limit
	q.Offset = offset
	inboxes, err := h.svc.List(c, q)
	if err != nil {
		writeInboxErr(ctx, err)
		return
	}
	resp := make([]inboxResp, 0, len(inboxes))
	for _, inbox := range inboxes {
		resp = append(resp, toInboxResp(inbox))
	}
	ctx.JSON(stdhttp.StatusOK, resp)
}

func (h *InboxHandler) Update(c context.Context, ctx *app.RequestContext) {
	var req updateInboxReq
	if err := ctx.BindJSON(&req); err != nil {
		ctx.JSON(stdhttp.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}
	inbox, err := h.svc.Update(c, ctx.Param("id"), application.UpdateInboxCmd{Title: req.Name, Description: req.Description})
	if err != nil {
		writeInboxErr(ctx, err)
		return
	}
	ctx.JSON(stdhttp.StatusOK, toInboxResp(inbox))
}

func (h *InboxHandler) Delete(c context.Context, ctx *app.RequestContext) {
	if err := h.svc.Delete(c, ctx.Param("id")); err != nil {
		writeInboxErr(ctx, err)
		return
	}
	ctx.Status(stdhttp.StatusNoContent)
}
