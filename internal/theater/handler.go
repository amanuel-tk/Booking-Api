package theater

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
	return &Handler{
		svc: svc,
	}
}

// CreateTheater godoc
// @Summary      Create a theater
// @Description  Create a new theater. The theater name and city must be provided and cannot exceed 100 characters. Requires admin access.
// @Tags         Theaters
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body CreateTheater true "Theater details"
// @Success      201 {object} Theater
// @Failure      400 {object} httpx.ErrorResponse
// @Failure      401 {object} httpx.ErrorResponse
// @Failure      403 {object} httpx.ErrorResponse
// @Failure      409 {object} httpx.ErrorResponse
// @Failure      500 {object} httpx.ErrorResponse
// @Router       /theaters [post]
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

// CreateScreen godoc
// @Summary      Create a screen
// @Description  Create a screen in a theater and automatically generate its seats. Rows must be between 1 and 26, and seats per row must be between 1 and 50. Requires admin access.
// @Tags         Screens
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Theater UUID" format(uuid)
// @Param        request body CreateScreenRequest true "Screen details"
// @Success      201 {object} Screen
// @Failure      400 {object} httpx.ErrorResponse
// @Failure      401 {object} httpx.ErrorResponse
// @Failure      403 {object} httpx.ErrorResponse
// @Failure      404 {object} httpx.ErrorResponse
// @Failure      409 {object} httpx.ErrorResponse
// @Failure      500 {object} httpx.ErrorResponse
// @Router       /theaters/{id}/screens [post]
func (h *Handler) CreateScreen(w http.ResponseWriter, r *http.Request) {
	theaterID := chi.URLParam(r, "id")

	var req CreateScreenRequest

	if !httpx.DecodeJson(w, r, &req) {
		return
	}

	screen, err := h.svc.CreateScreen(r.Context(), theaterID, req)

	if err != nil {
		var ve ValidationError
		switch {
		case errors.Is(err, ErrNotFound):
			httpx.WriteError(w, http.StatusNotFound, "not found")
			return
		case errors.As(err, &ve):
			httpx.WriteError(w, http.StatusBadRequest, ve.Msg)
			return
		case errors.Is(err, ErrScreenNameTaken):
			httpx.WriteError(w, http.StatusConflict, ErrScreenNameTaken.Error())
			return
		default:
			log.Printf("creating screen: %v", err)
			httpx.WriteError(w, http.StatusInternalServerError, "internal server error")
			return
		}
	}

	httpx.WriteJSON(w, http.StatusCreated, screen)
}

// GetTheaters godoc
// @Summary      List theaters
// @Description  Retrieve all active theaters ordered by creation date. Returns an empty array if no theaters exist.
// @Tags         Theaters
// @Produce      json
// @Success      200 {array} Theater
// @Failure      500 {object} httpx.ErrorResponse
// @Router       /theaters [get]
func (h *Handler) GetTheaters(w http.ResponseWriter, r *http.Request) {
	t, err := h.svc.GetTheaters(r.Context())

	if err != nil {
		log.Printf("getting theater: %v", err)
		httpx.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, t)
}

// GetScreen godoc
// @Summary      Get screen layout
// @Description  Retrieve a screen and its seats grouped by row. Each row contains its label and corresponding seats.
// @Tags         Screens
// @Produce      json
// @Param        id path string true "Screen UUID" format(uuid)
// @Success      200 {object} ScreenLayout
// @Failure      404 {object} httpx.ErrorResponse
// @Failure      500 {object} httpx.ErrorResponse
// @Router       /screens/{id} [get]
func (h *Handler) GetScreen(w http.ResponseWriter, r *http.Request) {
	layout, err := h.svc.GetScreen(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			httpx.WriteError(w, http.StatusNotFound, "not found")
			return
		}
		log.Printf("getting screen: %v", err)
		httpx.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, layout)
}

// ListScreens godoc
// @Summary      List theater screens
// @Description  Retrieve all active screens belonging to a theater. Returns an empty array if no active screens exist.
// @Tags         Screens
// @Produce      json
// @Param        id path string true "Theater UUID" format(uuid)
// @Success      200 {array} Screen
// @Failure      404 {object} httpx.ErrorResponse
// @Failure      500 {object} httpx.ErrorResponse
// @Router       /theaters/{id}/screens [get]
func (h *Handler) ListScreens(w http.ResponseWriter, r *http.Request) {
	screens, err := h.svc.ListScreens(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			httpx.WriteError(w, http.StatusNotFound, "not found")
			return
		}
		log.Printf("listing screens: %v", err)
		httpx.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, screens)
}
