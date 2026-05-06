package annotation

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/K1tten2005/lyryx-backend/internal/rest_api/auth"
	"github.com/K1tten2005/lyryx-backend/internal/rest_api/middlewares"
	usecase "github.com/K1tten2005/lyryx-backend/internal/usecases/annotation"
	"github.com/K1tten2005/lyryx-backend/internal/usecases/annotation/dto"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

var (
	ErrAnnotationsNotFound = errors.New("annotations not found")
	ErrAnnotationNotFound  = errors.New("annotation not found")
	ErrForbidden           = errors.New("forbidden")
	ErrSongNotFound        = errors.New("song not found")
	ErrUserNotFound        = errors.New("user not found")
	ErrInvalidIndex        = errors.New("start_index or end_index is out of range")
	ErrAnnotationOverlap   = errors.New("annotation overlaps with an existing one")
)

//go:generate mockgen -source=handlers.go -destination=mocks/mock_handlers.go -package=mocks

type annotationUsecase interface {
	GetAnnotations(ctx context.Context, opts dto.GetAnnotationOpts) ([]dto.AnnotationInfo, error)
	GetAnnotationByID(ctx context.Context, opts dto.GetAnnotationByIDOpts) (dto.AnnotationInfo, error)
	CreateAnnotation(ctx context.Context, opts dto.PostAnnotationOpts) (dto.AnnotationInfo, error)
	UpdateAnnotation(ctx context.Context, opts dto.PatchUpdateAnnotationOpts) (dto.AnnotationInfo, error)
	DeleteAnnotation(ctx context.Context, opts dto.DeleteAnnotationOpts) error
	VoteAnnotation(ctx context.Context, opts dto.PostVoteAnnotationOpts) (int, error)
	DeleteVote(ctx context.Context, opts dto.DeleteVoteOpts) error
	GetUserAnnotations(ctx context.Context, opts dto.GetUserAnnotationsOpts) ([]dto.AnnotationInfo, int, error)
}

type claimsGetter interface {
	GetClaims(c echo.Context) (*auth.JwtCustomClaims, error)
}

type Handlers struct {
	annotationUsecase annotationUsecase
	claimsGetter      claimsGetter
	logger            *logrus.Logger
}

func NewHandlers(
	annotationUsecase annotationUsecase,
	claimsGetter claimsGetter,
	logger *logrus.Logger,
) *Handlers {
	return &Handlers{
		annotationUsecase: annotationUsecase,
		claimsGetter:      claimsGetter,
		logger:            logger,
	}
}

func (h *Handlers) RegisterHandlers(
	e *echo.Echo,
	authMiddleware echo.MiddlewareFunc,
	optionalAuthMiddleware echo.MiddlewareFunc,
	roleChecker *middlewares.RolesCheckerMiddleware,
) {
	public := e.Group("")
	public.Use(optionalAuthMiddleware)
	public.GET("/v1/song/:id/annotations", h.GetAnnotations)
	public.GET("/v1/annotation/:id", h.GetAnnotationByID)
	public.GET("/v1/user/:id/annotations", h.GetUserAnnotations)
	public.GET("/v1/user/:id/annotations", h.GetUserAnnotations)

	private := e.Group("")
	private.Use(authMiddleware)
	private.POST("/v1/song/:id/annotation", h.PostAnnotation)
	private.PATCH("/v1/annotation/:id", h.PatchUpdateAnnotation)
	private.DELETE("/v1/annotation/:id", h.DeleteAnnotation)
	private.POST("/v1/annotation/:id/vote", h.PostVoteAnnotation)
	private.DELETE("/v1/annotation/:id/vote", h.DeleteVote)
}

// ==================== LIST ANNOTATIONS ====================

// GetAnnotations godoc
// @Summary      Получение списка аннотаций песни
// @Description  Возвращает все аннотации для указанной песни без пагинации
// @Tags         annotation
// @Produce      json
// @Param        id       path     int     true  "Song ID"
// @Success      200      {object} GetAnnotationsOut
// @Failure      400      {object} echo.HTTPError  "Некорректный запрос"
// @Failure      404      {object} echo.HTTPError  "Песня не найдена"
// @Failure      500      {object} echo.HTTPError  "Внутренняя ошибка сервера"
// @Router       /v1/song/{id}/annotations [get]
func (h *Handlers) GetAnnotations(c echo.Context) error {
	ctx := c.Request().Context()

	req := new(GetAnnotationsIn)
	if err := c.Bind(req); err != nil {
		h.logger.WithError(err).Warning("bind failed")
		return echo.NewHTTPError(http.StatusBadRequest, echo.Map{"error": "invalid input"})
	}
	if err := c.Validate(req); err != nil {
		h.logger.WithError(err).Warning("validate failed")
		return echo.NewHTTPError(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}

	var userID *int
	claims, err := h.claimsGetter.GetClaims(c)
	if err == nil && claims != nil {
		userID = &claims.UserID
	}

	annotations, err := h.annotationUsecase.GetAnnotations(ctx, dto.GetAnnotationOpts{
		SongID: req.SongID,
		UserID: userID,
	})
	if err != nil {
		h.logger.WithError(err).WithField("song_id", req.SongID).Warning("get annotations failed")
		if errors.Is(err, usecase.ErrSongNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, echo.Map{"error": ErrSongNotFound.Error()})
		}
		if errors.Is(err, usecase.ErrAnnotationsNotFound) {
			return c.JSON(http.StatusOK, GetAnnotationsOut{
				SongID:      req.SongID,
				Annotations: []Annotation{},
			})
		}
		return echo.NewHTTPError(http.StatusInternalServerError, echo.Map{"error": "internal error"})
	}

	out := getAnnotationsToOut(annotations)

	return c.JSON(http.StatusOK, out)
}

func getAnnotationsToOut(annotations []dto.AnnotationInfo) GetAnnotationsOut {
	annotationsOut := make([]Annotation, 0, len(annotations))
	for _, a := range annotations {
		annotationsOut = append(annotationsOut, Annotation{
			ID:         a.ID,
			User:       userInfoToOut(a.User),
			Content:    a.Content,
			StartIndex: a.StartIndex,
			EndIndex:   a.EndIndex,
			Rating:     a.Rating,
			CreatedAt:  a.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			MyVote:     a.MyVote,
		})
	}
	return GetAnnotationsOut{
		SongID:      annotations[0].Song.ID,
		Annotations: annotationsOut,
	}
}

// ==================== GET BY ID ====================

// GetAnnotationByID godoc
// @Summary      Получение аннотации по ID
// @Description  Возвращает полную информацию об аннотации с контекстом песни
// @Tags         annotation
// @Produce      json
// @Param        id   path  int  true  "Annotation ID"
// @Success      200  {object} GetAnnotationByIDOut
// @Failure      400  {object} echo.HTTPError
// @Failure      404  {object} echo.HTTPError  "Аннотация не найдена"
// @Failure      500  {object} echo.HTTPError
// @Router       /v1/annotation/{id} [get]
func (h *Handlers) GetAnnotationByID(c echo.Context) error {
	ctx := c.Request().Context()

	req := new(GetAnnotationByIDIn)
	if err := c.Bind(req); err != nil {
		h.logger.WithError(err).Warning("bind failed")
		return echo.NewHTTPError(http.StatusBadRequest, echo.Map{"error": "invalid input"})
	}
	if err := c.Validate(req); err != nil {
		h.logger.WithError(err).Warning("validate failed")
		return echo.NewHTTPError(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}

	var userID *int
	claims, err := h.claimsGetter.GetClaims(c)
	if err == nil && claims != nil {
		userID = &claims.UserID
	}

	ann, err := h.annotationUsecase.GetAnnotationByID(ctx, dto.GetAnnotationByIDOpts{
		AnnotationID: req.AnnotationID,
		UserID:       userID,
	})
	if err != nil {
		h.logger.WithError(err).WithField("annotation_id", req.AnnotationID).Warning("get annotation failed")
		if errors.Is(err, usecase.ErrAnnotationNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, echo.Map{"error": ErrAnnotationNotFound.Error()})
		}
		return echo.NewHTTPError(http.StatusInternalServerError, echo.Map{"error": "internal error"})
	}

	out := getAnnotationByIDToOut(ann)
	return c.JSON(http.StatusOK, out)
}

func getAnnotationByIDToOut(ann dto.AnnotationInfo) GetAnnotationByIDOut {
	return GetAnnotationByIDOut{
		ID:         ann.ID,
		Song:       songInfoToOut(ann.Song),
		User:       userInfoToOut(ann.User),
		Content:    ann.Content,
		StartIndex: ann.StartIndex,
		EndIndex:   ann.EndIndex,
		Rating:     ann.Rating,
		CreatedAt:  ann.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:  ann.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		MyVote:     ann.MyVote,
	}
}

// ==================== CREATE ====================

// PostAnnotation godoc
// @Summary      Создание аннотации
// @Description  Создает новую аннотацию для песни. Не допускается пересечение диапазонов с уже существующими аннотациями.
// @Tags         annotation
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        id        path  int               true  "Song ID"
// @Param        request   body  PostAnnotationIn  true  "Данные аннотации"
// @Success      201       {object} PostAnnotationOut
// @Failure      400       {object} echo.HTTPError "Невалидные данные, неверный индекс или пересечение выделений"
// @Failure      401       {object} echo.HTTPError "Не авторизован"
// @Failure      404       {object} echo.HTTPError "Песня не найдена"
// @Failure      500       {object} echo.HTTPError "Внутренняя ошибка сервера"
// @Router       /v1/song/{id}/annotation [post]
func (h *Handlers) PostAnnotation(c echo.Context) error {
	ctx := c.Request().Context()

	claims, err := h.claimsGetter.GetClaims(c)
	if err != nil {
		h.logger.WithError(err).Warning("get claims failed")
		return echo.NewHTTPError(http.StatusUnauthorized, echo.Map{"error": "unauthorized"})
	}

	req := new(PostAnnotationIn)
	if err := c.Bind(req); err != nil {
		h.logger.WithError(err).Warning("bind failed")
		return echo.NewHTTPError(http.StatusBadRequest, echo.Map{"error": "invalid input"})
	}
	if err := c.Validate(req); err != nil {
		h.logger.WithError(err).Warning("validate failed")
		return echo.NewHTTPError(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}

	opts, err := postAnnotationToOpts(req, claims.UserID)
	if err != nil {
		h.logger.WithError(err).Warning("validate failed")
		return echo.NewHTTPError(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}

	ann, err := h.annotationUsecase.CreateAnnotation(ctx, opts)
	if err != nil {
		h.logger.WithError(err).WithField("user_id", claims.UserID).Warning("create annotation failed")
		if errors.Is(err, usecase.ErrSongNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, echo.Map{"error": ErrSongNotFound.Error()})
		}
		if errors.Is(err, usecase.ErrInvalidIndex) {
			return echo.NewHTTPError(http.StatusBadRequest, echo.Map{"error": ErrInvalidIndex.Error()})
		}
		if errors.Is(err, usecase.ErrAnnotationOverlap) {
			return echo.NewHTTPError(http.StatusBadRequest, echo.Map{"error": ErrAnnotationOverlap.Error()})
		}
		return echo.NewHTTPError(http.StatusInternalServerError, echo.Map{"error": "internal error"})
	}

	out := postAnnotationToOut(ann)
	return c.JSON(http.StatusCreated, out)
}

func postAnnotationToOpts(req *PostAnnotationIn, userID int) (dto.PostAnnotationOpts, error) {
	if req == nil {
		return dto.PostAnnotationOpts{}, errors.New("invalid input")
	}

	if *req.StartIndex < 0 || *req.EndIndex <= *req.StartIndex {
		return dto.PostAnnotationOpts{}, errors.New("invalid start_index or end_index")
	}

	return dto.PostAnnotationOpts{
		SongID:     req.SongID,
		UserID:     userID,
		Content:    strings.TrimSpace(req.Content),
		StartIndex: *req.StartIndex,
		EndIndex:   *req.EndIndex,
	}, nil
}

func postAnnotationToOut(ann dto.AnnotationInfo) PostAnnotationOut {
	return PostAnnotationOut{
		ID:         ann.ID,
		SongID:     ann.Song.ID,
		User:       userInfoToOut(ann.User),
		Content:    ann.Content,
		StartIndex: ann.StartIndex,
		EndIndex:   ann.EndIndex,
		Rating:     ann.Rating,
		CreatedAt:  ann.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

// ==================== UPDATE ====================

// UpdateAnnotation godoc
// @Summary      Редактирование аннотации
// @Description  Обновляет контент аннотации. Доступно только автору или модератору.
// @Tags         annotation
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        id        path  int                      true  "Annotation ID"
// @Param        request   body  PatchUpdateAnnotationIn  true  "Данные для обновления"
// @Success      200       {object} PatchUpdateAnnotationOut
// @Failure      400       {object} echo.HTTPError
// @Failure      401       {object} echo.HTTPError
// @Failure      403       {object} echo.HTTPError  "Недостаточно прав"
// @Failure      404       {object} echo.HTTPError
// @Failure      500       {object} echo.HTTPError
// @Router       /v1/annotation/{id} [patch]
func (h *Handlers) PatchUpdateAnnotation(c echo.Context) error {
	ctx := c.Request().Context()

	claims, err := h.claimsGetter.GetClaims(c)
	if err != nil {
		h.logger.WithError(err).Warning("get claims failed")
		return echo.NewHTTPError(http.StatusUnauthorized, echo.Map{"error": "unauthorized"})
	}

	req := new(PatchUpdateAnnotationIn)
	if err := c.Bind(req); err != nil {
		h.logger.WithError(err).Warning("bind failed")
		return echo.NewHTTPError(http.StatusBadRequest, echo.Map{"error": "invalid input"})
	}
	if err := c.Validate(req); err != nil {
		h.logger.WithError(err).Warning("validate failed")
		return echo.NewHTTPError(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}

	opts, err := patchUpdateAnnotationToOpts(req, claims.UserID)
	if err != nil {
		h.logger.WithError(err).Warning("validate failed")
		return echo.NewHTTPError(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}

	ann, err := h.annotationUsecase.UpdateAnnotation(ctx, opts)
	if err != nil {
		h.logger.WithError(err).WithField("annotation_id", req.AnnotationID).Warning("update annotation failed")
		if errors.Is(err, usecase.ErrAnnotationNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, echo.Map{"error": ErrAnnotationNotFound.Error()})
		}
		if errors.Is(err, usecase.ErrForbidden) {
			return echo.NewHTTPError(http.StatusForbidden, echo.Map{"error": ErrForbidden.Error()})
		}
		return echo.NewHTTPError(http.StatusInternalServerError, echo.Map{"error": "internal error"})
	}

	out := PatchUpdateAnnotationOut{
		ID:        ann.ID,
		Content:   ann.Content,
		UpdatedAt: ann.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		Rating:    ann.Rating,
	}
	return c.JSON(http.StatusOK, out)
}

func patchUpdateAnnotationToOpts(req *PatchUpdateAnnotationIn, userID int) (dto.PatchUpdateAnnotationOpts, error) {
	if req == nil {
		return dto.PatchUpdateAnnotationOpts{}, errors.New("invalid input")
	}

	if req.Content == nil {
		return dto.PatchUpdateAnnotationOpts{}, errors.New("content is required")
	}

	return dto.PatchUpdateAnnotationOpts{
		AnnotationID: req.AnnotationID,
		Content:      strings.TrimSpace(*req.Content),
		UserID:       userID,
	}, nil
}

// ==================== DELETE ====================

// DeleteAnnotation godoc
// @Summary      Удаление аннотации
// @Description  Удаляет аннотацию. Доступно только автору или модератору.
// @Tags         annotation
// @Security     ApiKeyAuth
// @Produce      json
// @Param        id   path  int  true  "Annotation ID"
// @Success      204  "No Content"
// @Failure      401  {object} echo.HTTPError
// @Failure      403  {object} echo.HTTPError
// @Failure      404  {object} echo.HTTPError
// @Failure      500  {object} echo.HTTPError
// @Router       /v1/annotation/{id} [delete]
func (h *Handlers) DeleteAnnotation(c echo.Context) error {
	ctx := c.Request().Context()

	claims, err := h.claimsGetter.GetClaims(c)
	if err != nil {
		h.logger.WithError(err).Warning("get claims failed")
		return echo.NewHTTPError(http.StatusUnauthorized, echo.Map{"error": "unauthorized"})
	}

	req := new(DeleteAnnotationIn)
	if err := c.Bind(req); err != nil {
		h.logger.WithError(err).Warning("bind failed")
		return echo.NewHTTPError(http.StatusBadRequest, echo.Map{"error": "invalid input"})
	}
	if err := c.Validate(req); err != nil {
		h.logger.WithError(err).Warning("validate failed")
		return echo.NewHTTPError(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}

	opts, err := deleteAnnotationToOpts(req, claims.UserID, claims.Role)
	if err != nil {
		h.logger.WithError(err).Warning("validate failed")
		return echo.NewHTTPError(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}

	err = h.annotationUsecase.DeleteAnnotation(ctx, opts)
	if err != nil {
		h.logger.WithError(err).WithField("annotation_id", req.AnnotationID).Warning("delete annotation failed")
		if errors.Is(err, usecase.ErrAnnotationNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, echo.Map{"error": ErrAnnotationNotFound.Error()})
		}
		if errors.Is(err, usecase.ErrForbidden) {
			return echo.NewHTTPError(http.StatusForbidden, echo.Map{"error": ErrForbidden.Error()})
		}
		return echo.NewHTTPError(http.StatusInternalServerError, echo.Map{"error": "internal error"})
	}

	return c.NoContent(http.StatusNoContent)
}

func deleteAnnotationToOpts(req *DeleteAnnotationIn, userID int, userRole string) (dto.DeleteAnnotationOpts, error) {
	if req == nil {
		return dto.DeleteAnnotationOpts{}, errors.New("invalid input")
	}

	return dto.DeleteAnnotationOpts{
		AnnotationID: req.AnnotationID,
		UserID:       userID,
		Role:         userRole,
	}, nil
}

// ==================== VOTE ====================

// PostVoteAnnotation godoc
// @Summary      Голосование за аннотацию
// @Description  Поставить +1 или -1 за аннотацию. Доступно только аутентифицированным пользователям.
// @Tags         annotation
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        id        path  int                   true  "Annotation ID"
// @Param        request   body  PostVoteAnnotationIn  true  "Значение голоса (-1 или 1)"
// @Success      200       {object} PostVoteAnnotationOut
// @Failure      400       {object} echo.HTTPError
// @Failure      401       {object} echo.HTTPError
// @Failure      404       {object} echo.HTTPError  "Аннотация не найдена"
// @Failure      500       {object} echo.HTTPError
// @Router       /v1/annotation/{id}/vote [post]
func (h *Handlers) PostVoteAnnotation(c echo.Context) error {
	ctx := c.Request().Context()

	claims, err := h.claimsGetter.GetClaims(c)
	if err != nil {
		h.logger.WithError(err).Warning("get claims failed")
		return echo.NewHTTPError(http.StatusUnauthorized, echo.Map{"error": "unauthorized"})
	}

	req := new(PostVoteAnnotationIn)
	if err := c.Bind(req); err != nil {
		h.logger.WithError(err).Warning("bind failed")
		return echo.NewHTTPError(http.StatusBadRequest, echo.Map{"error": "invalid input"})
	}
	if err := c.Validate(req); err != nil {
		h.logger.WithError(err).Warning("validate failed")
		return echo.NewHTTPError(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}

	opts, err := voteAnnotationToOpts(req, claims.UserID)
	if err != nil {
		h.logger.WithError(err).Warning("validate failed")
		return echo.NewHTTPError(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}

	newRating, err := h.annotationUsecase.VoteAnnotation(ctx, opts)
	if err != nil {
		h.logger.WithError(err).WithField("annotation_id", req.AnnotationID).Warning("vote annotation failed")
		if errors.Is(err, usecase.ErrAnnotationNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, echo.Map{"error": ErrAnnotationNotFound.Error()})
		}
		return echo.NewHTTPError(http.StatusInternalServerError, echo.Map{"error": "internal error"})
	}

	out := PostVoteAnnotationOut{
		AnnotationID: req.AnnotationID,
		NewRating:    newRating,
		MyVote:       &req.Value,
	}
	return c.JSON(http.StatusOK, out)
}

func voteAnnotationToOpts(req *PostVoteAnnotationIn, userID int) (dto.PostVoteAnnotationOpts, error) {
	if req == nil {
		return dto.PostVoteAnnotationOpts{}, errors.New("invalid input")
	}

	if req.Value != -1 && req.Value != 1 {
		return dto.PostVoteAnnotationOpts{}, errors.New("value must be -1 or 1")
	}

	return dto.PostVoteAnnotationOpts{
		AnnotationID: req.AnnotationID,
		UserID:       userID,
		Value:        req.Value,
	}, nil
}

// DeleteVote godoc
// @Summary      Отмена голоса за аннотацию
// @Description  Удаляет голос пользователя за аннотацию. Доступно только аутентифицированным пользователям.
// @Tags         annotation
// @Security     ApiKeyAuth
// @Produce      json
// @Param        id   path  int  true  "Annotation ID"
// @Success      204  "No Content"
// @Failure      401  {object} echo.HTTPError
// @Failure      404  {object} echo.HTTPError  "Аннотация не найдена"
// @Failure      500  {object} echo.HTTPError
// @Router       /v1/annotation/{id}/vote [delete]
func (h *Handlers) DeleteVote(c echo.Context) error {
	ctx := c.Request().Context()

	claims, err := h.claimsGetter.GetClaims(c)
	if err != nil {
		h.logger.WithError(err).Warning("get claims failed")
		return echo.NewHTTPError(http.StatusUnauthorized, echo.Map{"error": "unauthorized"})
	}

	req := new(DeleteVoteIn)
	if err := c.Bind(req); err != nil {
		h.logger.WithError(err).Warning("bind failed")
		return echo.NewHTTPError(http.StatusBadRequest, echo.Map{"error": "invalid input"})
	}
	if err := c.Validate(req); err != nil {
		h.logger.WithError(err).Warning("validate failed")
		return echo.NewHTTPError(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}

	err = h.annotationUsecase.DeleteVote(ctx, dto.DeleteVoteOpts{
		AnnotationID: req.AnnotationID,
		UserID:       claims.UserID,
	})
	if err != nil {
		h.logger.WithError(err).WithField("annotation_id", req.AnnotationID).Warning("remove vote failed")
		if errors.Is(err, usecase.ErrAnnotationNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, echo.Map{"error": ErrAnnotationNotFound.Error()})
		}
		return echo.NewHTTPError(http.StatusInternalServerError, echo.Map{"error": "internal error"})
	}

	return c.NoContent(http.StatusNoContent)
}

// ==================== USER'S ANNOTATIONS ====================

// GetUserAnnotations godoc
// @Summary      Получение аннотаций пользователя
// @Description  Возвращает список аннотаций, созданных пользователем
// @Tags         annotation
// @Produce      json
// @Param        id       path  int  true  "User ID"
// @Param        limit    query int  false "Лимит"  default(20)
// @Param        offset   query int  false "Смещение"  default(0)
// @Success      200      {object} GetUserAnnotationsOut
// @Failure      400      {object} echo.HTTPError
// @Failure      404      {object} echo.HTTPError  "Пользователь не найден"
// @Failure      500      {object} echo.HTTPError
// @Router       /v1/user/{id}/annotations [get]
func (h *Handlers) GetUserAnnotations(c echo.Context) error {
	ctx := c.Request().Context()

	req := new(GetUserAnnotationsIn)
	if err := c.Bind(req); err != nil {
		h.logger.WithError(err).Warning("bind failed")
		return echo.NewHTTPError(http.StatusBadRequest, echo.Map{"error": "invalid input"})
	}
	if err := c.Validate(req); err != nil {
		h.logger.WithError(err).Warning("validate failed")
		return echo.NewHTTPError(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}

	var currentUserID *int
	claims, err := h.claimsGetter.GetClaims(c)
	if err == nil && claims != nil {
		currentUserID = &claims.UserID
	}

	opts := getUserAnnotationsInToOpts(req, currentUserID)

	annotations, total, err := h.annotationUsecase.GetUserAnnotations(ctx, opts)
	if err != nil {
		h.logger.WithError(err).WithField("user_id", req.UserID).Warning("get user annotations failed")
		if errors.Is(err, usecase.ErrUserNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, echo.Map{"error": ErrUserNotFound.Error()})
		}
		return echo.NewHTTPError(http.StatusInternalServerError, echo.Map{"error": "internal error"})
	}

	out := GetUserAnnotationsOut{
		UserID:      req.UserID,
		Annotations: annotationsToOut(annotations),
		Total:       total,
		HasMore:     req.Offset+len(annotations) < total,
	}
	return c.JSON(http.StatusOK, out)
}

func getUserAnnotationsInToOpts(req *GetUserAnnotationsIn, currentUserID *int) dto.GetUserAnnotationsOpts {
	limit := req.Limit
	if limit <= 0 {
		limit = 20
	}
	offset := max(0, req.Offset)

	return dto.GetUserAnnotationsOpts{
		UserID:        req.UserID,
		CurrentUserID: currentUserID,
		Limit:         limit,
		Offset:        offset,
	}
}

// ==================== HELPERS ====================

func annotationsToOut(anns []dto.AnnotationInfo) []Annotation {
	result := make([]Annotation, 0, len(anns))
	for _, a := range anns {
		result = append(result, Annotation{
			ID:         a.ID,
			User:       userInfoToOut(a.User),
			Content:    a.Content,
			StartIndex: a.StartIndex,
			EndIndex:   a.EndIndex,
			Rating:     a.Rating,
			CreatedAt:  a.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			MyVote:     a.MyVote,
		})
	}
	return result
}

func userInfoToOut(info dto.UserInfo) UserInfo {
	return UserInfo{
		UserID:          info.UserID,
		Username:        info.Username,
		AvatarURL:       info.AvatarURL,
		ReputationScore: info.ReputationScore,
	}
}

func songInfoToOut(song dto.SongInfo) SongInfo {
	return SongInfo{
		ID:       song.ID,
		Title:    song.Title,
		Artist:   ArtistInfo{ID: song.Artist.ID, Name: song.Artist.Name},
		CoverURL: song.CoverURL,
	}
}
