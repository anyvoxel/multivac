package http

import (
	"context"
	"errors"
	stdhttp "net/http"
	"strconv"
	"strings"
	"time"

	"github.com/cloudwego/hertz/pkg/app"

	"github.com/anyvoxel/multivac/pkg/item/application"
	"github.com/anyvoxel/multivac/pkg/item/domain"
)

func splitCSV(v string) []string {
	if strings.TrimSpace(v) == "" {
		return nil
	}
	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))
	seen := map[string]struct{}{}
	for _, part := range parts {
		normalized := strings.ToLower(strings.TrimSpace(part))
		if normalized == "" {
			continue
		}
		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		out = append(out, normalized)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func parseSortBy(v string) (domain.SortBy, bool) {
	switch strings.ToLower(v) {
	case "createdat", "created_at":
		return domain.ItemSortByCreatedAt, true
	case "updatedat", "updated_at":
		return domain.ItemSortByUpdatedAt, true
	case "title", "name":
		return domain.ItemSortByTitle, true
	case "dueat", "due_at":
		return domain.ItemSortByDueAt, true
	case "expectedat", "expected_at":
		return domain.ItemSortByExpectedAt, true
	case "priority":
		return domain.ItemSortByPriority, true
	default:
		return "", false
	}
}

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

type labelReq struct {
	Value string `json:"value"`
}

type createItemReq struct {
	Kind        string     `json:"kind"`
	Bucket      string     `json:"bucket"`
	ProjectID   *string    `json:"projectId,omitempty"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Labels      []labelReq `json:"labels"`
	Context     string     `json:"context"`
	Details     string     `json:"details"`
	TaskStatus  string     `json:"taskStatus,omitempty"`
	Priority    string     `json:"priority,omitempty"`
	WaitingFor  string     `json:"waitingFor,omitempty"`
	ExpectedAt  *time.Time `json:"expectedAt,omitempty"`
	DueAt       *time.Time `json:"dueAt,omitempty"`
}

type updateBucketReq struct {
	Bucket string `json:"bucket"`
}

type labelResp struct {
	Value      string `json:"value"`
	Kind       string `json:"kind"`
	Filterable bool   `json:"filterable"`
}

type itemResp struct {
	ID          string      `json:"id"`
	Kind        string      `json:"kind"`
	Bucket      string      `json:"bucket"`
	ProjectID   string      `json:"projectId,omitempty"`
	Title       string      `json:"title"`
	Description string      `json:"description"`
	Labels      []labelResp `json:"labels"`
	Context     string      `json:"context"`
	Details     string      `json:"details"`
	TaskStatus  string     `json:"taskStatus,omitempty"`
	Priority    string     `json:"priority,omitempty"`
	WaitingFor  string     `json:"waitingFor,omitempty"`
	ExpectedAt  *time.Time `json:"expectedAt,omitempty"`
	DueAt       *time.Time `json:"dueAt,omitempty"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
}

func toDomainLabels(labels []labelReq) []domain.Label {
	if len(labels) == 0 {
		return nil
	}
	out := make([]domain.Label, 0, len(labels))
	for _, label := range labels {
		value := strings.TrimSpace(label.Value)
		if value == "" {
			continue
		}
		kind := domain.LabelKindTag
		filterable := false
		if strings.HasPrefix(value, "@") {
			kind = domain.LabelKindContext
			filterable = true
			value = strings.TrimSpace(strings.TrimPrefix(value, "@"))
		} else if strings.HasPrefix(value, "#") {
			kind = domain.LabelKindTag
			filterable = true
			value = strings.TrimSpace(strings.TrimPrefix(value, "#"))
		}
		if value == "" {
			continue
		}
		out = append(out, domain.Label{Value: strings.ToLower(value), Kind: kind, Filterable: filterable})
	}
	return out
}

func toLabelResp(labels []domain.Label) []labelResp {
	if len(labels) == 0 {
		return []labelResp{}
	}
	out := make([]labelResp, 0, len(labels))
	for _, label := range labels {
		out = append(out, labelResp{Value: label.Value, Kind: string(label.Kind), Filterable: label.Filterable})
	}
	return out
}

func toResp(item *domain.Item) itemResp {
	return itemResp{ID: item.ID, Kind: string(item.Kind), Bucket: string(item.Bucket), ProjectID: item.ProjectID, Title: item.Title, Description: item.Description, Labels: toLabelResp(item.Labels), Context: item.Context, Details: item.Details, TaskStatus: item.TaskStatus, Priority: item.Priority, WaitingFor: item.WaitingFor, ExpectedAt: item.ExpectedAt, DueAt: item.DueAt, CreatedAt: item.CreatedAt, UpdatedAt: item.UpdatedAt}
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
	kind, ok := domain.ParseKind(v)
	if !ok {
		return "", domain.InvalidKind(v)
	}
	return kind, nil
}

func parseBucket(v string) (domain.Bucket, error) {
	bucket, ok := domain.ParseBucket(v)
	if !ok {
		return "", domain.InvalidBucket(v)
	}
	return bucket, nil
}

func toCreateCmd(req createItemReq) (application.CreateItemCmd, error) {
	kind, err := parseKind(req.Kind)
	if err != nil {
		return application.CreateItemCmd{}, err
	}
	bucket, err := parseBucket(req.Bucket)
	if err != nil {
		return application.CreateItemCmd{}, err
	}
	projectID := ""
	if req.ProjectID != nil {
		projectID = strings.TrimSpace(*req.ProjectID)
	}
	return application.CreateItemCmd{Kind: kind, Bucket: bucket, ProjectID: projectID, Title: req.Title, Description: req.Description, Labels: toDomainLabels(req.Labels), Context: req.Context, Details: req.Details, TaskStatus: req.TaskStatus, Priority: req.Priority, WaitingFor: req.WaitingFor, ExpectedAt: req.ExpectedAt, DueAt: req.DueAt}, nil
}

func (h *Handler) Create(c context.Context, ctx *app.RequestContext) {
	var req createItemReq
	if err := ctx.BindJSON(&req); err != nil {
		ctx.JSON(stdhttp.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}
	cmd, err := toCreateCmd(req)
	if err != nil {
		writeErr(ctx, err)
		return
	}
	item, err := h.svc.Create(c, cmd)
	if err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(stdhttp.StatusCreated, toResp(item))
}

func (h *Handler) Get(c context.Context, ctx *app.RequestContext) {
	item, err := h.svc.Get(c, ctx.Param("id"))
	if err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(stdhttp.StatusOK, toResp(item))
}

func (h *Handler) List(c context.Context, ctx *app.RequestContext) {
	q := domain.ListQuery{
		ProjectID:  strings.TrimSpace(ctx.Query("projectId")),
		TaskStatus: strings.TrimSpace(ctx.Query("taskStatus")),
		Search:     strings.TrimSpace(ctx.Query("search")),
		Contexts:   splitCSV(ctx.Query("contexts")),
		Tags:       splitCSV(ctx.Query("tags")),
	}
	if v := ctx.Query("bucket"); v != "" {
		b, err := parseBucket(v)
		if err != nil {
			writeErr(ctx, err)
			return
		}
		q.Bucket = &b
	}
	if v := ctx.Query("kind"); v != "" {
		k, err := parseKind(v)
		if err != nil {
			writeErr(ctx, err)
			return
		}
		q.Kind = &k
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
	q.Limit, q.Offset = limit, offset
	items, err := h.svc.List(c, q)
	if err != nil {
		writeErr(ctx, err)
		return
	}
	out := make([]itemResp, 0, len(items))
	for _, item := range items {
		out = append(out, toResp(item))
	}
	ctx.JSON(stdhttp.StatusOK, out)
}

func (h *Handler) Update(c context.Context, ctx *app.RequestContext) {
	var req createItemReq
	if err := ctx.BindJSON(&req); err != nil {
		ctx.JSON(stdhttp.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}
	cmd, err := toCreateCmd(req)
	if err != nil {
		writeErr(ctx, err)
		return
	}
	item, err := h.svc.Update(c, ctx.Param("id"), application.UpdateItemCmd(cmd))
	if err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(stdhttp.StatusOK, toResp(item))
}

func (h *Handler) MoveBucket(c context.Context, ctx *app.RequestContext) {
	var req updateBucketReq
	if err := ctx.BindJSON(&req); err != nil {
		ctx.JSON(stdhttp.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}
	bucket, err := parseBucket(req.Bucket)
	if err != nil {
		writeErr(ctx, err)
		return
	}
	item, err := h.svc.MoveBucket(c, ctx.Param("id"), bucket)
	if err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(stdhttp.StatusOK, toResp(item))
}

func (h *Handler) Delete(c context.Context, ctx *app.RequestContext) {
	if err := h.svc.Delete(c, ctx.Param("id")); err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.Status(stdhttp.StatusNoContent)
}
