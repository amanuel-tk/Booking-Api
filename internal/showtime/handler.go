package showtime

import (
	"errors"
	"log"
	"net/http"

	"github.com/amanuel-tk/Booking-Api/internal/httpx"
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
