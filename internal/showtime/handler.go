package showtime

import (
	"errors"
	"log"
	"net/http"

	"github.com/amanuel-tk/Booking-Api/internal/httpx"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) CreateShowtime(w http.ResponseWriter, r *http.Request) {
	var req CreateShowtimeRequest
	if !httpx.DecodeJson(w, r, &req) {
		return
	}

	st, err := h.svc.CreateShowtime(r.Context(), req)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, st)
}

func (h *Handler) GetShowtime(w http.ResponseWriter, r *http.Request) {
	st, err := h.svc.GetShowtime(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, st)
}

func (h *Handler) ListByMovie(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	from := q.Get("from")
	if from == "" {
		from = q.Get("date")
	}
	items, err := h.svc.ListByMovie(r.Context(), chi.URLParam(r, "id"), from, q.Get("to"))

	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, items)
}

func (h *Handler) writeServiceError(w http.ResponseWriter, err error) {
	var ve ValidationError

	switch {
	case errors.As(err, &ve):
		httpx.WriteError(w, http.StatusBadRequest, ve.Msg)
	case errors.Is(err, ErrNotFound), errors.Is(err, ErrMovieNotFound), errors.Is(err, ErrScreenNotFound):
		httpx.WriteError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, ErrShowtimeOverlap):
		httpx.WriteError(w, http.StatusConflict, err.Error())
	default:
		log.Printf("showtime handler: %v", err)
		httpx.WriteError(w, http.StatusInternalServerError, "internal server error")
	}
}
