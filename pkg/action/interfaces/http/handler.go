package http

import (
	"context"
	"errors"
	stdhttp "net/http"
	"strconv"
	"strings"
	"time"

	"github.com/cloudwego/hertz/pkg/app"

	"github.com/anyvoxel/multivac/pkg/action/application"
	"github.com/anyvoxel/multivac/pkg/action/domain"
)

func parsePagination(ctx *app.RequestContext) (limit, offset int, err error) {
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

func parseSortBy(v string) (domain.SortBy, bool) {
	switch strings.ToLower(v) {
	case "createdat", "created_at":
		return domain.ActionSortByCreatedAt, true
	case "updatedat", "updated_at":
		return domain.ActionSortByUpdatedAt, true
	case "name", "title":
		return domain.ActionSortByTitle, true
	default:
		return "", false
	}
}

func parseSort(ctx *app.RequestContext) ([]domain.Sort, error) {
	sortByStr := ctx.Query("sortBy")
	if sortByStr == "" {
		return nil, nil
	}
	by, ok := parseSortBy(sortByStr)
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

type Handler struct{ svc *application.Service }

func NewHandler(svc *application.Service) *Handler { return &Handler{svc: svc} }

type createActionReq struct {
	Title       string            `json:"title"`
	Description string            `json:"description"`
	ProjectID   *string           `json:"project_id"`
	Kind        string            `json:"kind"`
	Context     []string          `json:"context"`
	Labels      []domain.Label    `json:"labels"`
	Attributes  domain.Attributes `json:"attributes"`
}

type updateActionReq = createActionReq

type convertInboxReq struct {
	Title       *string            `json:"title"`
	Description *string            `json:"description"`
	ProjectID   *string            `json:"project_id"`
	Kind        *string            `json:"kind"`
	Context     []string           `json:"context"`
	Labels      []domain.Label     `json:"labels"`
	Attributes  *domain.Attributes `json:"attributes"`
}

type actionResp struct {
	ID          string            `json:"id"`
	Title       string            `json:"title"`
	Description string            `json:"description"`
	ProjectID   *string           `json:"project_id,omitempty"`
	Kind        string            `json:"kind"`
	Context     []string          `json:"context"`
	Labels      []domain.Label    `json:"labels"`
	Attributes  domain.Attributes `json:"attributes"`
	CreatedAt   time.Time         `json:"createdAt"`
	UpdatedAt   time.Time         `json:"updatedAt"`
}

func toResp(action *domain.Action) actionResp {
	return actionResp{
		ID:          action.ID,
		Title:       action.Title,
		Description: action.Description,
		ProjectID:   action.ProjectID,
		Kind:        string(action.Kind),
		Context:     action.ContextIDs,
		Labels:      action.Labels,
		Attributes:  action.Attributes,
		CreatedAt:   action.CreatedAt,
		UpdatedAt:   action.UpdatedAt,
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

func parseKind(v string) (domain.Kind, error) {
	if strings.TrimSpace(v) == "" {
		return "", nil
	}
	kind, ok := domain.ParseKind(v)
	if !ok {
		return "", domain.InvalidKind(v)
	}
	return kind, nil
}

func (h *Handler) Create(c context.Context, ctx *app.RequestContext) {
	var req createActionReq
	if err := ctx.BindJSON(&req); err != nil {
		ctx.JSON(stdhttp.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}
	kind, err := parseKind(req.Kind)
	if err != nil {
		writeErr(ctx, err)
		return
	}
	action, err := h.svc.Create(c, application.CreateActionCmd{
		Title:       req.Title,
		Description: req.Description,
		ProjectID:   req.ProjectID,
		Kind:        kind,
		ContextIDs:  req.Context,
		Labels:      req.Labels,
		Attributes:  req.Attributes,
	})
	if err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(stdhttp.StatusCreated, toResp(action))
}

func (h *Handler) ConvertFromInbox(c context.Context, ctx *app.RequestContext) {
	var req convertInboxReq
	if len(ctx.Request.Body()) > 0 {
		if err := ctx.BindJSON(&req); err != nil {
			ctx.JSON(stdhttp.StatusBadRequest, map[string]any{"error": err.Error()})
			return
		}
	}
	var kind *domain.Kind
	if req.Kind != nil {
		k, err := parseKind(*req.Kind)
		if err != nil {
			writeErr(ctx, err)
			return
		}
		kind = &k
	}
	action, err := h.svc.ConvertFromInbox(c, ctx.Param("id"), application.ConvertFromInboxCmd{
		Title:       req.Title,
		Description: req.Description,
		ProjectID:   req.ProjectID,
		Kind:        kind,
		ContextIDs:  req.Context,
		Labels:      req.Labels,
		Attributes:  req.Attributes,
	})
	if err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(stdhttp.StatusCreated, toResp(action))
}

func (h *Handler) Get(c context.Context, ctx *app.RequestContext) {
	action, err := h.svc.Get(c, ctx.Param("id"))
	if err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(stdhttp.StatusOK, toResp(action))
}

func (h *Handler) List(c context.Context, ctx *app.RequestContext) {
	q := domain.ListQuery{Search: strings.TrimSpace(ctx.Query("search"))}
	if kindStr := strings.TrimSpace(ctx.Query("kind")); kindStr != "" {
		kind, ok := domain.ParseKind(kindStr)
		if !ok {
			writeErr(ctx, domain.InvalidKind(kindStr))
			return
		}
		q.Kind = &kind
	}
	if projectID := strings.TrimSpace(ctx.Query("projectId")); projectID != "" {
		q.ProjectID = &projectID
	}
	sorts, err := parseSort(ctx)
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
	q.Limit = limit
	q.Offset = offset
	actions, err := h.svc.List(c, q)
	if err != nil {
		writeErr(ctx, err)
		return
	}
	resp := make([]actionResp, 0, len(actions))
	for _, action := range actions {
		resp = append(resp, toResp(action))
	}
	ctx.JSON(stdhttp.StatusOK, resp)
}

func (h *Handler) Update(c context.Context, ctx *app.RequestContext) {
	var req updateActionReq
	if err := ctx.BindJSON(&req); err != nil {
		ctx.JSON(stdhttp.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}
	kind, err := parseKind(req.Kind)
	if err != nil {
		writeErr(ctx, err)
		return
	}
	action, err := h.svc.Update(c, ctx.Param("id"), application.UpdateActionCmd{
		Title:       req.Title,
		Description: req.Description,
		ProjectID:   req.ProjectID,
		Kind:        kind,
		ContextIDs:  req.Context,
		Labels:      req.Labels,
		Attributes:  req.Attributes,
	})
	if err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(stdhttp.StatusOK, toResp(action))
}

func (h *Handler) Delete(c context.Context, ctx *app.RequestContext) {
	if err := h.svc.Delete(c, ctx.Param("id")); err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.Status(stdhttp.StatusNoContent)
}
