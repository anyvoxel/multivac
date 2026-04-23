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

func parseProjectSortBy(v string) (domain.SortBy, bool) {
	switch strings.ToLower(v) {
	case "createdat", "created_at":
		return domain.ProjectSortByCreatedAt, true
	case "updatedat", "updated_at":
		return domain.ProjectSortByUpdatedAt, true
	case "title":
		return domain.ProjectSortByTitle, true
	default:
		return "", false
	}
}

func parseProjectPagination(ctx *app.RequestContext) (limit, offset int, err error) {
	if limitStr := ctx.Query("limit"); limitStr != "" {
		n, convErr := strconv.Atoi(limitStr)
		if convErr != nil || n < 0 {
			return 0, 0, domain.InvalidPaginationValue("limit", limitStr)
		}
		limit = n
	}
	if offsetStr := ctx.Query("offset"); offsetStr != "" {
		n, convErr := strconv.Atoi(offsetStr)
		if convErr != nil || n < 0 {
			return 0, 0, domain.InvalidPaginationValue("offset", offsetStr)
		}
		offset = n
	}
	return limit, offset, nil
}

func parseProjectSort(ctx *app.RequestContext) ([]domain.Sort, error) {
	sortByStr := ctx.Query("sortBy")
	if sortByStr == "" {
		return nil, nil
	}
	by, ok := parseProjectSortBy(sortByStr)
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

type ProjectHandler struct{ svc *application.ProjectService }

func NewProjectHandler(svc *application.ProjectService) *ProjectHandler {
	return &ProjectHandler{svc: svc}
}

type goalReq struct {
	Title       string     `json:"title"`
	CreatedAt   *time.Time `json:"created_at,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

type projectReferenceReq struct {
	Title string `json:"title"`
	URL   string `json:"URL"`
}

type createProjectReq struct {
	Title       string                `json:"title"`
	Goals       []goalReq             `json:"goals"`
	Description string                `json:"description"`
	References  []projectReferenceReq `json:"references"`
}

type updateProjectReq struct {
	Title       string                `json:"title"`
	Goals       []goalReq             `json:"goals"`
	Description string                `json:"description"`
	References  []projectReferenceReq `json:"references"`
}

type setStatusReq struct {
	Status string `json:"status"`
}

type goalResp struct {
	Title       string     `json:"title"`
	CreatedAt   time.Time  `json:"created_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

type projectReferenceResp struct {
	Title string `json:"title"`
	URL   string `json:"URL"`
}

type projectResp struct {
	ID          string                 `json:"id"`
	Title       string                 `json:"title"`
	Goals       []goalResp             `json:"goals"`
	Description string                 `json:"description"`
	References  []projectReferenceResp `json:"references"`
	Status      string                 `json:"status"`
	StartAt     *time.Time             `json:"startAt,omitempty"`
	CompletedAt *time.Time             `json:"completedAt,omitempty"`
	CreatedAt   time.Time              `json:"createdAt"`
	UpdatedAt   time.Time              `json:"updatedAt"`
}

func toDomainProjectGoals(goals []goalReq) []domain.ProjectGoal {
	if len(goals) == 0 {
		return nil
	}
	out := make([]domain.ProjectGoal, 0, len(goals))
	for _, goal := range goals {
		value := strings.TrimSpace(goal.Title)
		if value == "" {
			continue
		}
		var createdAt time.Time
		if goal.CreatedAt != nil {
			createdAt = *goal.CreatedAt
		}
		out = append(out, domain.ProjectGoal{Title: value, CreatedAt: createdAt, CompletedAt: goal.CompletedAt})
	}
	return out
}

func toProjectGoalResp(goals []domain.ProjectGoal) []goalResp {
	if len(goals) == 0 {
		return []goalResp{}
	}
	out := make([]goalResp, 0, len(goals))
	for _, goal := range goals {
		out = append(out, goalResp{Title: goal.Title, CreatedAt: goal.CreatedAt, CompletedAt: goal.CompletedAt})
	}
	return out
}

func toDomainProjectReferences(references []projectReferenceReq) []domain.ProjectReference {
	if len(references) == 0 {
		return nil
	}
	out := make([]domain.ProjectReference, 0, len(references))
	for _, reference := range references {
		out = append(out, domain.ProjectReference{Title: strings.TrimSpace(reference.Title), URL: strings.TrimSpace(reference.URL)})
	}
	return out
}

func toProjectReferenceResp(references []domain.ProjectReference) []projectReferenceResp {
	if len(references) == 0 {
		return []projectReferenceResp{}
	}
	out := make([]projectReferenceResp, 0, len(references))
	for _, reference := range references {
		out = append(out, projectReferenceResp{Title: reference.Title, URL: reference.URL})
	}
	return out
}

func toProjectResp(p *domain.Project) projectResp {
	return projectResp{
		ID:          p.ID,
		Title:       p.Title,
		Goals:       toProjectGoalResp(p.Goals),
		Description: p.Description,
		References:  toProjectReferenceResp(p.References),
		Status:      string(p.Status),
		StartAt:     p.StartAt,
		CompletedAt: p.CompletedAt,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
}

func writeProjectErr(ctx *app.RequestContext, err error) {
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

func (h *ProjectHandler) Create(c context.Context, ctx *app.RequestContext) {
	var req createProjectReq
	if err := ctx.BindJSON(&req); err != nil {
		ctx.JSON(stdhttp.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}
	p, err := h.svc.Create(c, application.CreateProjectCmd{
		Title:       req.Title,
		Goals:       toDomainProjectGoals(req.Goals),
		Description: req.Description,
		References:  toDomainProjectReferences(req.References),
	})
	if err != nil {
		writeProjectErr(ctx, err)
		return
	}
	ctx.JSON(stdhttp.StatusCreated, toProjectResp(p))
}

func (h *ProjectHandler) Get(c context.Context, ctx *app.RequestContext) {
	p, err := h.svc.Get(c, ctx.Param("id"))
	if err != nil {
		writeProjectErr(ctx, err)
		return
	}
	ctx.JSON(stdhttp.StatusOK, toProjectResp(p))
}

func (h *ProjectHandler) List(c context.Context, ctx *app.RequestContext) {
	statusStr := ctx.Query("status")
	var status *domain.Status
	if statusStr != "" {
		s, ok := domain.ParseStatus(statusStr)
		if !ok {
			writeProjectErr(ctx, domain.InvalidStatus(statusStr))
			return
		}
		status = &s
	}

	q := domain.ProjectListQuery{Status: status, Search: strings.TrimSpace(ctx.Query("search"))}
	sorts, err := parseProjectSort(ctx)
	if err != nil {
		writeProjectErr(ctx, err)
		return
	}
	q.Sorts = sorts
	limit, offset, err := parseProjectPagination(ctx)
	if err != nil {
		writeProjectErr(ctx, err)
		return
	}
	q.Limit, q.Offset = limit, offset

	items, err := h.svc.List(c, q)
	if err != nil {
		writeProjectErr(ctx, err)
		return
	}
	out := make([]projectResp, 0, len(items))
	for _, p := range items {
		out = append(out, toProjectResp(p))
	}
	ctx.JSON(stdhttp.StatusOK, out)
}

func (h *ProjectHandler) Update(c context.Context, ctx *app.RequestContext) {
	id := ctx.Param("id")
	var req updateProjectReq
	if err := ctx.BindJSON(&req); err != nil {
		ctx.JSON(stdhttp.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}
	p, err := h.svc.Update(c, id, application.UpdateProjectCmd{
		Title:       req.Title,
		Goals:       toDomainProjectGoals(req.Goals),
		Description: req.Description,
		References:  toDomainProjectReferences(req.References),
	})
	if err != nil {
		writeProjectErr(ctx, err)
		return
	}
	ctx.JSON(stdhttp.StatusOK, toProjectResp(p))
}

func (h *ProjectHandler) SetStatus(c context.Context, ctx *app.RequestContext) {
	id := ctx.Param("id")
	var req setStatusReq
	if err := ctx.BindJSON(&req); err != nil {
		ctx.JSON(stdhttp.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}
	st, ok := domain.ParseStatus(req.Status)
	if !ok {
		writeProjectErr(ctx, domain.InvalidStatus(req.Status))
		return
	}
	p, err := h.svc.SetStatus(c, id, st)
	if err != nil {
		writeProjectErr(ctx, err)
		return
	}
	ctx.JSON(stdhttp.StatusOK, toProjectResp(p))
}

func (h *ProjectHandler) Delete(c context.Context, ctx *app.RequestContext) {
	if err := h.svc.Delete(c, ctx.Param("id")); err != nil {
		writeProjectErr(ctx, err)
		return
	}
	ctx.Status(stdhttp.StatusNoContent)
}
