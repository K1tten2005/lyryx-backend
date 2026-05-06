package search

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/K1tten2005/lyryx-backend/internal/rest_api/middlewares"
	usecase "github.com/K1tten2005/lyryx-backend/internal/usecases/search"
	"github.com/K1tten2005/lyryx-backend/internal/usecases/search/dto"
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

type searchUsecase interface {
	GetSearch(ctx context.Context, opts dto.GetSearchOpts) (dto.SearchResult, error)
}

type Handlers struct {
	searchUsecase searchUsecase
	logger        *logrus.Logger
}

func NewHandlers(
	searchUsecase searchUsecase,
	logger *logrus.Logger,
) *Handlers {
	return &Handlers{
		searchUsecase: searchUsecase,
		logger:        logger,
	}
}

func (h *Handlers) RegisterHandlers(
	e *echo.Echo,
	authMiddleware echo.MiddlewareFunc,
	roleChecker *middlewares.RolesCheckerMiddleware,
) {
	public := e.Group("")
	public.GET("/v1/search", h.GetSearch)
}

// GetSearch godoc
// @Summary      Поиск по сайту
// @Description  Полнотекстовый поиск по песням, артистам и пользователям с поддержкой русского и английского языков
// @Tags         search
// @Produce      json
// @Param        q        query  string  true  "Поисковый запрос"
// @Param        limit    query  int     false "Лимит результатов на категорию"  default(20)
// @Success      200      {object} GetSearchOut
// @Failure      400      {object} echo.HTTPError
// @Router       /v1/search [get]
func (h *Handlers) GetSearch(c echo.Context) error {
	ctx := c.Request().Context()

	req := new(GetSearchIn)
	if err := c.Bind(req); err != nil {
		h.logger.WithError(err).Warning("bind failed")
		return echo.NewHTTPError(http.StatusBadRequest, echo.Map{"error": "invalid input"})
	}
	if err := c.Validate(req); err != nil {
		h.logger.WithError(err).Warning("validate failed")
		return echo.NewHTTPError(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}

	opts := getSearchToOpts(req)

	searchRes, err := h.searchUsecase.GetSearch(ctx, opts)
	if err != nil {
		h.logger.WithError(err).Warning("get annotations failed")
		if errors.Is(err, usecase.ErrResultNotFound) {
			return c.JSON(http.StatusOK, GetSearchOut{
				Songs:              []SongInfo{},
				LyricsMatchedSongs: []SongInfo{},
				Users:              []UserInfo{},
				Artists:            []ArtistInfo{},
			})
		}
		return echo.NewHTTPError(http.StatusInternalServerError, echo.Map{"error": "internal error"})
	}

	out := getSearchToOut(searchRes)

	return c.JSON(http.StatusOK, out)
}

func getSearchToOpts(req *GetSearchIn) dto.GetSearchOpts {
	limit := req.Limit
	if limit <= 0 {
		limit = 20
	}
	if limit > 50 {
		limit = 50 // защита от перегрузки
	}
	return dto.GetSearchOpts{
		Query: strings.TrimSpace(req.Query),
		Limit: limit,
	}
}

func getSearchToOut(res dto.SearchResult) GetSearchOut {
    return GetSearchOut{
        Songs:              mapSongInfos(res.Songs),
        LyricsMatchedSongs: mapSongInfos(res.LyricsMatchedSongs),
        Artists:            mapArtistInfos(res.Artists),
        Users:              mapUserInfos(res.Users),
    }
}

func mapSongInfos(src []dto.SongInfo) []SongInfo {
    out := make([]SongInfo, 0, len(src))
    for _, s := range src {
        out = append(out, SongInfo{
            ID:            s.ID,
            Title:         s.Title,
            LyricsSnippet: s.LyricsSnippet,
			Views:         s.Views,
            Artist: ArtistInfo{
                ID:        s.Artist.ID,
                Name:      s.Artist.Name,
                AvatarURL: s.Artist.AvatarURL,
            },
            CoverURL: s.CoverURL,
        })
    }
    return out
}

func mapArtistInfos(src []dto.ArtistInfo) []ArtistInfo {
    out := make([]ArtistInfo, 0, len(src))
    for _, s := range src {
        out = append(out, ArtistInfo{
			ID:        s.ID,
			Name:      s.Name,
			AvatarURL: s.AvatarURL,
		})
    }
    return out
}

func mapUserInfos(src []dto.UserInfo) []UserInfo {
    out := make([]UserInfo, 0, len(src))
    for _, s := range src {
        out = append(out, UserInfo{
            UserID:          s.UserID,
            Username:        s.Username,
            AvatarURL:       s.AvatarURL,
            ReputationScore: s.ReputationScore,
        })
    }
    return out
}