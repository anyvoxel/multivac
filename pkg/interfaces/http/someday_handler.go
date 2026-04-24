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

func parseSomedayPagination(ctx *app.RequestContext) (limit, offset int, err error) {
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

func parseSomedaySortBy(v string) (domain.SortBy, bool) {
	switch strings.ToLower(v) {
	case "createdat", "created_at":
		return domain.SomedaySortByCreatedAt, true
	case "updatedat", "updated_at":
		return domain.SomedaySortByUpdatedAt, true
	case "name", "title":
		return domain.SomedaySortByTitle, true
	default:
		return "", false
	}
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
	if v := ctx.Query("sortDir"); v != "" {
		parsed, ok := domain.ParseSortDir(v)
		if !ok {
			return nil, domain.InvalidSortDir(v)
		}
		dir = parsed
	}
	return []domain.Sort{{By: by, Dir: dir}}, nil
}

type SomedayHandler struct{ svc *application.SomedayService }

func NewSomedayHandler(svc *application.SomedayService) *SomedayHandler { return &SomedayHandler{svc: svc} }

type createSomedayReq struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

type updateSomedayReq struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

type convertSomedayInboxReq struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
}

type convertSomedayActionReq struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
}

type somedayResp struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

func toSomedayResp(someday *domain.Someday) somedayResp {
	return somedayResp{ID: someday.ID, Title: someday.Title, Description: someday.Description, CreatedAt: someday.CreatedAt, UpdatedAt: someday.UpdatedAt}
}

func writeSomedayErr(ctx *app.RequestContext, err error) {
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

func (h *SomedayHandler) Create(c context.Context, ctx *app.RequestContext) {
	var req createSomedayReq
	if err := ctx.BindJSON(&req); err != nil {
		ctx.JSON(stdhttp.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}
	someday, err := h.svc.Create(c, application.CreateSomedayCmd{Title: req.Title, Description: req.Description})
	if err != nil {
		writeSomedayErr(ctx, err)
		return
	}
	ctx.JSON(stdhttp.StatusCreated, toSomedayResp(someday))
}

func (h *SomedayHandler) ConvertFromInbox(c context.Context, ctx *app.RequestContext) {
	var req convertSomedayInboxReq
	if len(ctx.Request.Body()) > 0 {
		if err := ctx.BindJSON(&req); err != nil {
			ctx.JSON(stdhttp.StatusBadRequest, map[string]any{"error": err.Error()})
			return
		}
	}
	someday, err := h.svc.ConvertFromInbox(c, ctx.Param("id"), application.ConvertSomedayFromInboxCmd{Title: req.Title, Description: req.Description})
	if err != nil {
		writeSomedayErr(ctx, err)
		return
	}
	ctx.JSON(stdhttp.StatusCreated, toSomedayResp(someday))
}

func (h *SomedayHandler) ConvertFromAction(c context.Context, ctx *app.RequestContext) {
	var req convertSomedayActionReq
	if len(ctx.Request.Body()) > 0 {
		if err := ctx.BindJSON(&req); err != nil {
			ctx.JSON(stdhttp.StatusBadRequest, map[string]any{"error": err.Error()})
			return
		}
	}
	someday, err := h.svc.ConvertFromAction(c, ctx.Param("id"), application.ConvertSomedayFromActionCmd{Title: req.Title, Description: req.Description})
	if err != nil {
		writeSomedayErr(ctx, err)
		return
	}
	ctx.JSON(stdhttp.StatusCreated, toSomedayResp(someday))
}

func (h *SomedayHandler) Get(c context.Context, ctx *app.RequestContext) {
	someday, err := h.svc.Get(c, ctx.Param("id"))
	if err != nil {
		writeSomedayErr(ctx, err)
		return
	}
	ctx.JSON(stdhttp.StatusOK, toSomedayResp(someday))
}

func (h *SomedayHandler) List(c context.Context, ctx *app.RequestContext) {
	q := domain.SomedayListQuery{Search: strings.TrimSpace(ctx.Query("search"))}
	sorts, err := parseSomedaySort(ctx)
	if err != nil {
		writeSomedayErr(ctx, err)
		return
	}
	q.Sorts = sorts
	limit, offset, err := parseSomedayPagination(ctx)
	if err != nil {
		writeSomedayErr(ctx, err)
		return
	}
	q.Limit = limit
	q.Offset = offset
	somedays, err := h.svc.List(c, q)
	if err != nil {
		writeSomedayErr(ctx, err)
		return
	}
	resp := make([]somedayResp, 0, len(somedays))
	for _, someday := range somedays {
		resp = append(resp, toSomedayResp(someday))
	}
	ctx.JSON(stdhttp.StatusOK, resp)
}

func (h *SomedayHandler) Update(c context.Context, ctx *app.RequestContext) {
	var req updateSomedayReq
	if err := ctx.BindJSON(&req); err != nil {
		ctx.JSON(stdhttp.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}
	someday, err := h.svc.Update(c, ctx.Param("id"), application.UpdateSomedayCmd{Title: req.Title, Description: req.Description})
	if err != nil {
		writeSomedayErr(ctx, err)
		return
	}
	ctx.JSON(stdhttp.StatusOK, toSomedayResp(someday))
}

func (h *SomedayHandler) Delete(c context.Context, ctx *app.RequestContext) {
	if err := h.svc.Delete(c, ctx.Param("id")); err != nil {
		writeSomedayErr(ctx, err)
		return
	}
	ctx.Status(stdhttp.StatusNoContent)
}
