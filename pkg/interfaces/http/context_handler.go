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

func parseContextPagination(ctx *app.RequestContext) (limit, offset int, err error) {
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

func parseContextSortBy(v string) (domain.SortBy, bool) {
	switch strings.ToLower(v) {
	case "createdat", "created_at":
		return domain.ContextSortByCreatedAt, true
	case "updatedat", "updated_at":
		return domain.ContextSortByUpdatedAt, true
	case "name", "title":
		return domain.ContextSortByTitle, true
	default:
		return "", false
	}
}

func parseContextSort(ctx *app.RequestContext) ([]domain.Sort, error) {
	sortByStr := ctx.Query("sortBy")
	if sortByStr == "" {
		return nil, nil
	}
	by, ok := parseContextSortBy(sortByStr)
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

type ContextHandler struct{ svc *application.ContextService }

func NewContextHandler(svc *application.ContextService) *ContextHandler { return &ContextHandler{svc: svc} }

type createContextReq struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Color       string `json:"color"`
}

type updateContextReq struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Color       string `json:"color"`
}

type contextResp struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Color       string    `json:"color"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

func toContextResp(contextObj *domain.Context) contextResp {
	return contextResp{ID: contextObj.ID, Title: contextObj.Title, Description: contextObj.Description, Color: contextObj.Color, CreatedAt: contextObj.CreatedAt, UpdatedAt: contextObj.UpdatedAt}
}

func writeContextErr(ctx *app.RequestContext, err error) {
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

func (h *ContextHandler) Create(c context.Context, ctx *app.RequestContext) {
	var req createContextReq
	if err := ctx.BindJSON(&req); err != nil {
		ctx.JSON(stdhttp.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}
	contextObj, err := h.svc.Create(c, application.CreateContextCmd{Title: req.Title, Description: req.Description, Color: req.Color})
	if err != nil {
		writeContextErr(ctx, err)
		return
	}
	ctx.JSON(stdhttp.StatusCreated, toContextResp(contextObj))
}

func (h *ContextHandler) Get(c context.Context, ctx *app.RequestContext) {
	contextObj, err := h.svc.Get(c, ctx.Param("id"))
	if err != nil {
		writeContextErr(ctx, err)
		return
	}
	ctx.JSON(stdhttp.StatusOK, toContextResp(contextObj))
}

func (h *ContextHandler) List(c context.Context, ctx *app.RequestContext) {
	q := domain.ContextListQuery{Search: strings.TrimSpace(ctx.Query("search"))}
	sorts, err := parseContextSort(ctx)
	if err != nil {
		writeContextErr(ctx, err)
		return
	}
	q.Sorts = sorts
	limit, offset, err := parseContextPagination(ctx)
	if err != nil {
		writeContextErr(ctx, err)
		return
	}
	q.Limit = limit
	q.Offset = offset
	contexts, err := h.svc.List(c, q)
	if err != nil {
		writeContextErr(ctx, err)
		return
	}
	resp := make([]contextResp, 0, len(contexts))
	for _, contextObj := range contexts {
		resp = append(resp, toContextResp(contextObj))
	}
	ctx.JSON(stdhttp.StatusOK, resp)
}

func (h *ContextHandler) Update(c context.Context, ctx *app.RequestContext) {
	var req updateContextReq
	if err := ctx.BindJSON(&req); err != nil {
		ctx.JSON(stdhttp.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}
	contextObj, err := h.svc.Update(c, ctx.Param("id"), application.UpdateContextCmd{Title: req.Title, Description: req.Description, Color: req.Color})
	if err != nil {
		writeContextErr(ctx, err)
		return
	}
	ctx.JSON(stdhttp.StatusOK, toContextResp(contextObj))
}

func (h *ContextHandler) Delete(c context.Context, ctx *app.RequestContext) {
	if err := h.svc.Delete(c, ctx.Param("id")); err != nil {
		writeContextErr(ctx, err)
		return
	}
	ctx.Status(stdhttp.StatusNoContent)
}
