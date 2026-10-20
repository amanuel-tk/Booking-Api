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

func currentUserID(r *http.Request) (string, bool) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		return "", false
	}
	return claims.Subject, true
}

// Create godoc
// @Summary      Hold seats for a showtime
// @Description  Temporarily reserve 1 to 6 seats for the authenticated user. The showtime must exist and must not have started. All seats must belong to the showtime's screen and must not already be booked or held by another user. Holds expire automatically after the configured Redis TTL.
// @Tags         Seat Holds
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Showtime UUID" format(uuid)
// @Param        request body Request true "Seat IDs to hold"
// @Success      201 {object} Result
// @Failure      400 {object} httpx.ErrorResponse
// @Failure      401 {object} httpx.ErrorResponse
// @Failure      404 {object} httpx.ErrorResponse
// @Failure      409 {object} SeatsUnavailableResponse
// @Failure      500 {object} httpx.ErrorResponse
// @Router       /showtimes/{id}/holds [post]
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
