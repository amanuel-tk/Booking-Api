package user

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

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest

	if !httpx.DecodeJson(w, r, &req) {
		return
	}

	u, err := h.svc.Register(r.Context(), req)

	if err != nil {
		var ve *ValidationError
		switch {
		case errors.As(err, &ve):
			httpx.WriteError(w, http.StatusBadRequest, ve.Msg)
		case errors.Is(err, ErrEmailTaken):
			httpx.WriteError(w, http.StatusConflict, ErrEmailTaken.Error())
		default:
			log.Printf("register: %v", err)
			httpx.WriteError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, u)
}
