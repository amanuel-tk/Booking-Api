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

func (h *Handler) SeatMap(w http.ResponseWriter, r *http.Request) {
	m, err := h.avail.SeatMap(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, m)
}
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
