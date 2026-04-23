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

func parseReferencePagination(ctx *app.RequestContext) (limit, offset int, err error) {
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

func parseReferenceSortBy(v string) (domain.SortBy, bool) {
	switch strings.ToLower(v) {
	case "createdat", "created_at":
		return domain.ReferenceSortByCreatedAt, true
	case "updatedat", "updated_at":
		return domain.ReferenceSortByUpdatedAt, true
	case "name", "title":
		return domain.ReferenceSortByTitle, true
	default:
		return "", false
	}
}

func parseReferenceSort(ctx *app.RequestContext) ([]domain.Sort, error) {
	sortByStr := ctx.Query("sortBy")
	if sortByStr == "" {
		return nil, nil
	}
	by, ok := parseReferenceSortBy(sortByStr)
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

type ReferenceHandler struct{ svc *application.ReferenceService }

func NewReferenceHandler(svc *application.ReferenceService) *ReferenceHandler { return &ReferenceHandler{svc: svc} }

type referenceLinkReq struct {
	Title string `json:"title"`
	URL   string `json:"url"`
}

type createReferenceReq struct {
	Title       string             `json:"title"`
	Description string             `json:"description"`
	References  []referenceLinkReq `json:"references"`
}

type updateReferenceReq struct {
	Title       string             `json:"title"`
	Description string             `json:"description"`
	References  []referenceLinkReq `json:"references"`
}

type referenceLinkResp struct {
	Title string `json:"title"`
	URL   string `json:"url"`
}

type referenceResp struct {
	ID          string              `json:"id"`
	Title       string              `json:"title"`
	Description string              `json:"description"`
	References  []referenceLinkResp `json:"references"`
	CreatedAt   time.Time           `json:"createdAt"`
	UpdatedAt   time.Time           `json:"updatedAt"`
}

func toDomainReferenceLinks(references []referenceLinkReq) []domain.ReferenceLink {
	if len(references) == 0 {
		return []domain.ReferenceLink{}
	}
	out := make([]domain.ReferenceLink, 0, len(references))
	for _, reference := range references {
		out = append(out, domain.ReferenceLink{Title: reference.Title, URL: reference.URL})
	}
	return out
}

func toRespReferenceLinks(references []domain.ReferenceLink) []referenceLinkResp {
	if len(references) == 0 {
		return []referenceLinkResp{}
	}
	out := make([]referenceLinkResp, 0, len(references))
	for _, reference := range references {
		out = append(out, referenceLinkResp{Title: reference.Title, URL: reference.URL})
	}
	return out
}

func toReferenceResp(reference *domain.Reference) referenceResp {
	return referenceResp{
		ID:          reference.ID,
		Title:       reference.Title,
		Description: reference.Description,
		References:  toRespReferenceLinks(reference.References),
		CreatedAt:   reference.CreatedAt,
		UpdatedAt:   reference.UpdatedAt,
	}
}

func writeReferenceErr(ctx *app.RequestContext, err error) {
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

func (h *ReferenceHandler) Create(c context.Context, ctx *app.RequestContext) {
	var req createReferenceReq
	if err := ctx.BindJSON(&req); err != nil {
		ctx.JSON(stdhttp.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}
	reference, err := h.svc.Create(c, application.CreateReferenceCmd{
		Title:       req.Title,
		Description: req.Description,
		References:  toDomainReferenceLinks(req.References),
	})
	if err != nil {
		writeReferenceErr(ctx, err)
		return
	}
	ctx.JSON(stdhttp.StatusCreated, toReferenceResp(reference))
}

func (h *ReferenceHandler) Get(c context.Context, ctx *app.RequestContext) {
	reference, err := h.svc.Get(c, ctx.Param("id"))
	if err != nil {
		writeReferenceErr(ctx, err)
		return
	}
	ctx.JSON(stdhttp.StatusOK, toReferenceResp(reference))
}

func (h *ReferenceHandler) List(c context.Context, ctx *app.RequestContext) {
	q := domain.ReferenceListQuery{Search: strings.TrimSpace(ctx.Query("search"))}
	sorts, err := parseReferenceSort(ctx)
	if err != nil {
		writeReferenceErr(ctx, err)
		return
	}
	q.Sorts = sorts
	limit, offset, err := parseReferencePagination(ctx)
	if err != nil {
		writeReferenceErr(ctx, err)
		return
	}
	q.Limit = limit
	q.Offset = offset
	references, err := h.svc.List(c, q)
	if err != nil {
		writeReferenceErr(ctx, err)
		return
	}
	resp := make([]referenceResp, 0, len(references))
	for _, reference := range references {
		resp = append(resp, toReferenceResp(reference))
	}
	ctx.JSON(stdhttp.StatusOK, resp)
}

func (h *ReferenceHandler) Update(c context.Context, ctx *app.RequestContext) {
	var req updateReferenceReq
	if err := ctx.BindJSON(&req); err != nil {
		ctx.JSON(stdhttp.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}
	reference, err := h.svc.Update(c, ctx.Param("id"), application.UpdateReferenceCmd{
		Title:       req.Title,
		Description: req.Description,
		References:  toDomainReferenceLinks(req.References),
	})
	if err != nil {
		writeReferenceErr(ctx, err)
		return
	}
	ctx.JSON(stdhttp.StatusOK, toReferenceResp(reference))
}

func (h *ReferenceHandler) Delete(c context.Context, ctx *app.RequestContext) {
	if err := h.svc.Delete(c, ctx.Param("id")); err != nil {
		writeReferenceErr(ctx, err)
		return
	}
	ctx.Status(stdhttp.StatusNoContent)
}
