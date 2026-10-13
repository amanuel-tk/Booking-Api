package movie

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
