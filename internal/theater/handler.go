package theater

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
	return &Handler{
		svc: svc,
	}
}

func (h *Handler) CreateTheater(w http.ResponseWriter, r *http.Request) {
	var req CreateTheater

	if !httpx.DecodeJson(w, r, &req) {
		return
	}

	t, err := h.svc.CreateTheater(r.Context(), req)

	if err != nil {
		var ve ValidationError
		switch {
		case errors.Is(err, ErrTheaterTaken):
			httpx.WriteError(w, http.StatusConflict, ErrTheaterTaken.Error())
			return
		case errors.As(err, &ve):
			httpx.WriteError(w, http.StatusBadRequest, ve.Msg)
			return
		default:
			log.Printf("creating theater: %v", err)
			httpx.WriteError(w, http.StatusInternalServerError, "internal server error")
			return
		}
	}
	httpx.WriteJSON(w, http.StatusCreated, t)

}
