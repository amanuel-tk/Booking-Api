package booking

import (
	"errors"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/amanuel-tk/Booking-Api/internal/auth"
	"github.com/amanuel-tk/Booking-Api/internal/httpx"
)

type Handler struct {
	svc   *Service
	avail *AvailabilityService
}

func NewHandler(svc *Service, avail *AvailabilityService) *Handler {
	return &Handler{svc: svc, avail: avail}
}

func currentUserID(r *http.Request) (string, bool) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		return "", false
	}
	return claims.Subject, true
}

// SeatMap godoc
// @Summary      Get showtime seat availability
// @Description  Retrieve all seats for a showtime grouped by row. Each seat has a status of free, held, or booked. Booked seats take precedence over temporary Redis holds.
// @Tags         Bookings
// @Produce      json
// @Param        id path string true "Showtime UUID" format(uuid)
// @Success      200 {object} SeatMap
// @Failure      404 {object} httpx.ErrorResponse
// @Failure      500 {object} httpx.ErrorResponse
// @Router       /showtimes/{id}/seats [get]
func (h *Handler) SeatMap(w http.ResponseWriter, r *http.Request) {
	m, err := h.avail.SeatMap(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, m)
}

// @Summary      Confirm a seat booking
// @Description  Create a confirmed booking for 1 to 6 seats. The user must own valid Redis holds for all requested seats. The showtime must be in the future and all seats must belong to its screen. The total price is calculated by the server.
// @Tags         Bookings
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Showtime UUID" format(uuid)
// @Param        request body CreateRequest true "Seat IDs to book"
// @Success      201 {object} BookingView
// @Failure      400 {object} httpx.ErrorResponse
// @Failure      401 {object} httpx.ErrorResponse
// @Failure      404 {object} httpx.ErrorResponse
// @Failure      409 {object} httpx.ErrorResponse
// @Failure      500 {object} httpx.ErrorResponse
// @Router       /showtimes/{id}/bookings [post]
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(r)
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req CreateRequest
	if !httpx.DecodeJson(w, r, &req) {
		return
	}

	b, err := h.svc.Create(r.Context(), userID, chi.URLParam(r, "id"), req.SeatIDs)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, b)
}

// List godoc
// @Summary      List my bookings
// @Description  Retrieve up to 100 bookings belonging to the authenticated user, ordered by creation date descending. Includes both confirmed and cancelled bookings.
// @Tags         Bookings
// @Produce      json
// @Security     BearerAuth
// @Success      200 {array} BookingView
// @Failure      401 {object} httpx.ErrorResponse
// @Failure      500 {object} httpx.ErrorResponse
// @Router       /me/bookings [get]
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(r)
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	items, err := h.svc.ListMine(r.Context(), userID)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, items)
}

// Get godoc
// @Summary      Get my booking
// @Description  Retrieve a booking by UUID. Only bookings belonging to the authenticated user can be accessed.
// @Tags         Bookings
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Booking UUID" format(uuid)
// @Success      200 {object} BookingView
// @Failure      401 {object} httpx.ErrorResponse
// @Failure      404 {object} httpx.ErrorResponse
// @Failure      500 {object} httpx.ErrorResponse
// @Router       /me/bookings/{id} [get]
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(r)
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	b, err := h.svc.GetMine(r.Context(), userID, chi.URLParam(r, "id"))
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, b)
}

// Cancel godoc
// @Summary      Cancel my booking
// @Description  Cancel a confirmed booking belonging to the authenticated user. Cancellation is allowed only when the showtime starts at least 2 hours from the current time. Successfully cancelled bookings release their seats for future bookings.
// @Tags         Bookings
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Booking UUID" format(uuid)
// @Success      200 {object} BookingView
// @Failure      401 {object} httpx.ErrorResponse
// @Failure      404 {object} httpx.ErrorResponse
// @Failure      409 {object} httpx.ErrorResponse
// @Failure      500 {object} httpx.ErrorResponse
// @Router       /me/bookings/{id}/cancel [post]
func (h *Handler) Cancel(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(r)
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	b, err := h.svc.Cancel(r.Context(), userID, chi.URLParam(r, "id"))
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, b)
}

func (h *Handler) writeServiceError(w http.ResponseWriter, err error) {
	var ve ValidationError

	switch {
	case errors.As(err, &ve):
		httpx.WriteError(w, http.StatusBadRequest, ve.Msg)
	case errors.Is(err, ErrNotFound), errors.Is(err, ErrShowtimeNotFound):
		httpx.WriteError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, ErrHoldNotOwned),
		errors.Is(err, ErrSeatsTaken),
		errors.Is(err, ErrAlreadyCancelled),
		errors.Is(err, ErrTooLateToCancel):
		httpx.WriteError(w, http.StatusConflict, err.Error())
	default:
		log.Printf("booking handler: %v", err)
		httpx.WriteError(w, http.StatusInternalServerError, "internal server error")
	}
}
