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
	}

	httpx.WriteJSON(w, http.StatusOK, m)

}

func (h *Handler) ListMovies(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	limit, _ := strconv.Atoi(q.Get("limit"))
	page, _ := strconv.Atoi(q.Get("page"))

	movies, total, err := h.svc.ListMovies(r.Context(), q.Get("q"), limit, page)

	if err != nil {
		log.Printf("list movies: %v", err)
		httpx.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	res := map[string]any{
		"movies": movies,
		"page":   page,
		"limit":  limit,
		"total":  total,
	}

	httpx.WriteJSON(w, http.StatusOK, res)
}

func (h *Handler) GetMovie(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	m, err := h.svc.GetMovie(r.Context(), id)

	if err != nil {
		if errors.Is(err, ErrNotFound) {
			httpx.WriteError(w, http.StatusNotFound, ErrNotFound.Error())
			return
		}
		httpx.WriteError(w, http.StatusInternalServerError, "internal server error")
	}

	httpx.WriteJSON(w, http.StatusOK, m)
}
