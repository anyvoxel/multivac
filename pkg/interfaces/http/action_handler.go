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

func parseActionPagination(ctx *app.RequestContext) (limit, offset int, err error) {
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

func parseActionSortBy(v string) (domain.SortBy, bool) {
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

func parseActionSort(ctx *app.RequestContext) ([]domain.Sort, error) {
	sortByStr := ctx.Query("sortBy")
	if sortByStr == "" {
		return nil, nil
	}
	by, ok := parseActionSortBy(sortByStr)
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

func parseActionKind(v string) (domain.Kind, error) {
	if strings.TrimSpace(v) == "" {
		return "", nil
	}
	kind, ok := domain.ParseKind(v)
	if !ok {
		return "", domain.InvalidKind(v)
	}
	return kind, nil
}

type ActionHandler struct{ svc *application.ActionService }

func NewActionHandler(svc *application.ActionService) *ActionHandler { return &ActionHandler{svc: svc} }

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

type convertActionInboxReq struct {
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

func toActionResp(action *domain.Action) actionResp {
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

func writeActionErr(ctx *app.RequestContext, err error) {
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

func (h *ActionHandler) Create(c context.Context, ctx *app.RequestContext) {
	var req createActionReq
	if err := ctx.BindJSON(&req); err != nil {
		ctx.JSON(stdhttp.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}
	kind, err := parseActionKind(req.Kind)
	if err != nil {
		writeActionErr(ctx, err)
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
		writeActionErr(ctx, err)
		return
	}
	ctx.JSON(stdhttp.StatusCreated, toActionResp(action))
}

func (h *ActionHandler) ConvertFromInbox(c context.Context, ctx *app.RequestContext) {
	var req convertActionInboxReq
	if len(ctx.Request.Body()) > 0 {
		if err := ctx.BindJSON(&req); err != nil {
			ctx.JSON(stdhttp.StatusBadRequest, map[string]any{"error": err.Error()})
			return
		}
	}
	var kind *domain.Kind
	if req.Kind != nil {
		k, err := parseActionKind(*req.Kind)
		if err != nil {
			writeActionErr(ctx, err)
			return
		}
		kind = &k
	}
	action, err := h.svc.ConvertFromInbox(c, ctx.Param("id"), application.ConvertActionFromInboxCmd{
		Title:       req.Title,
		Description: req.Description,
		ProjectID:   req.ProjectID,
		Kind:        kind,
		ContextIDs:  req.Context,
		Labels:      req.Labels,
		Attributes:  req.Attributes,
	})
	if err != nil {
		writeActionErr(ctx, err)
		return
	}
	ctx.JSON(stdhttp.StatusCreated, toActionResp(action))
}

func (h *ActionHandler) Get(c context.Context, ctx *app.RequestContext) {
	action, err := h.svc.Get(c, ctx.Param("id"))
	if err != nil {
		writeActionErr(ctx, err)
		return
	}
	ctx.JSON(stdhttp.StatusOK, toActionResp(action))
}

func (h *ActionHandler) List(c context.Context, ctx *app.RequestContext) {
	q := domain.ActionListQuery{Search: strings.TrimSpace(ctx.Query("search"))}
	if kindStr := strings.TrimSpace(ctx.Query("kind")); kindStr != "" {
		kind, ok := domain.ParseKind(kindStr)
		if !ok {
			writeActionErr(ctx, domain.InvalidKind(kindStr))
			return
		}
		q.Kind = &kind
	}
	if projectID := strings.TrimSpace(ctx.Query("projectId")); projectID != "" {
		q.ProjectID = &projectID
	}
	sorts, err := parseActionSort(ctx)
	if err != nil {
		writeActionErr(ctx, err)
		return
	}
	q.Sorts = sorts
	limit, offset, err := parseActionPagination(ctx)
	if err != nil {
		writeActionErr(ctx, err)
		return
	}
	q.Limit = limit
	q.Offset = offset
	actions, err := h.svc.List(c, q)
	if err != nil {
		writeActionErr(ctx, err)
		return
	}
	resp := make([]actionResp, 0, len(actions))
	for _, action := range actions {
		resp = append(resp, toActionResp(action))
	}
	ctx.JSON(stdhttp.StatusOK, resp)
}

func (h *ActionHandler) Update(c context.Context, ctx *app.RequestContext) {
	var req updateActionReq
	if err := ctx.BindJSON(&req); err != nil {
		ctx.JSON(stdhttp.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}
	kind, err := parseActionKind(req.Kind)
	if err != nil {
		writeActionErr(ctx, err)
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
		writeActionErr(ctx, err)
		return
	}
	ctx.JSON(stdhttp.StatusOK, toActionResp(action))
}

func (h *ActionHandler) Delete(c context.Context, ctx *app.RequestContext) {
	if err := h.svc.Delete(c, ctx.Param("id")); err != nil {
		writeActionErr(ctx, err)
		return
	}
	ctx.Status(stdhttp.StatusNoContent)
}
