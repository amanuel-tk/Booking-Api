package hold

import (
	"errors"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/amanuel-tk/Booking-Api/internal/auth"
	"github.com/amanuel-tk/Booking-Api/internal/httpx"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// ADJUST THIS to match how your auth package exposes the user id.
// The id must come from the JWT, never from the request body.
func currentUserID(r *http.Request) (string, bool) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		return "", false
	}
	return claims.Subject, true
}

// POST /showtimes/{id}/holds (logged in)
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(r)
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req Request
	if !httpx.DecodeJson(w, r, &req) {
		return
	}

	res, err := h.svc.HoldSeats(r.Context(), chi.URLParam(r, "id"), userID, req.SeatIDs)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, res)
}

func (h *Handler) writeServiceError(w http.ResponseWriter, err error) {
	var ve ValidationError
	var ue SeatsUnavailableError

	switch {
	case errors.As(err, &ve):
		httpx.WriteError(w, http.StatusBadRequest, ve.Msg)
	case errors.As(err, &ue):
		httpx.WriteJSON(w, http.StatusConflict, map[string]any{
			"error":    ue.Error(),
			"seat_ids": ue.SeatIDs,
		})
	case errors.Is(err, ErrNotFound):
		httpx.WriteError(w, http.StatusNotFound, err.Error())
	default:
		log.Printf("hold handler: %v", err)
		httpx.WriteError(w, http.StatusInternalServerError, "internal server error")
	}
}
