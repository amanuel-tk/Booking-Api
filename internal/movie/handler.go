package movie

import (
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/amanuel-tk/Booking-Api/internal/httpx"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{
		svc: svc,
	}
}

// Create godoc
// @Summary      Create a movie
// @Description  Create a new movie. Title is required (max 200 characters), duration must be between 1 and 600 minutes, and release_date must use YYYY-MM-DD format. Requires admin access.
// @Tags         Movies
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body CreateRequest true "Movie details"
// @Success      201 {object} Movie
// @Failure      400 {object} httpx.ErrorResponse
// @Failure      401 {object} httpx.ErrorResponse
// @Failure      403 {object} httpx.ErrorResponse
// @Failure      500 {object} httpx.ErrorResponse
// @Router       /movies [post]
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateRequest

	if !httpx.DecodeJson(w, r, &req) {
		return
	}

	m, err := h.svc.Create(r.Context(), req)

	if err != nil {
		var ve *ValidationError

		if errors.As(err, &ve) {
			httpx.WriteError(w, http.StatusBadRequest, ve.Msg)
			return
		}
		log.Printf("create movie: %v", err)
		httpx.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	httpx.WriteJSON(w, http.StatusCreated, m)

}

// ListMovies godoc
// @Summary      List movies
// @Description  Retrieve active movies with optional title search and pagination. Results are ordered by release date descending. Default page is 1, default limit is 20, and maximum limit is 100.
// @Tags         Movies
// @Produce      json
// @Param        q query string false "Search movies by title"
// @Param        page query int false "Page number (default 1)" default(1) minimum(1)
// @Param        limit query int false "Movies per page (default 20, max 100)" default(20) minimum(1) maximum(100)
// @Success      200 {object} ListMoviesResponse
// @Failure      500 {object} httpx.ErrorResponse
// @Router       /movies [get]
func (h *Handler) ListMovies(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	limit, _ := strconv.Atoi(q.Get("limit"))
	page, _ := strconv.Atoi(q.Get("page"))

	if limit < 1 || limit > 100 {
		limit = 20
	}

	if page < 1 {
		page = 1
	}

	movies, total, err := h.svc.ListMovies(
		r.Context(),
		q.Get("q"),
		limit,
		page,
	)

	if err != nil {
		log.Printf("list movies: %v", err)
		httpx.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	httpx.WriteJSON(w, http.StatusOK, ListMoviesResponse{
		Movies: movies,
		Page:   page,
		Limit:  limit,
		Total:  total,
	})
}

// GetMovie godoc
// @Summary      Get a movie
// @Description  Retrieve an active movie by its UUID. Returns 404 if the movie does not exist or is inactive.
// @Tags         Movies
// @Produce      json
// @Param        id path string true "Movie UUID" format(uuid)
// @Success      200 {object} Movie
// @Failure      404 {object} httpx.ErrorResponse
// @Failure      500 {object} httpx.ErrorResponse
// @Router       /movies/{id} [get]
func (h *Handler) GetMovie(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	m, err := h.svc.GetMovie(r.Context(), id)

	if err != nil {
		if errors.Is(err, ErrNotFound) {
			httpx.WriteError(w, http.StatusNotFound, ErrNotFound.Error())
			return
		}
		httpx.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	httpx.WriteJSON(w, http.StatusOK, m)
}
