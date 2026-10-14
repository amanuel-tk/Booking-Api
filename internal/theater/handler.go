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

func (h *Handler) CreateScreen(w http.ResponseWriter, r *http.Request) {
	theaterID := chi.URLParam(r, "id")

	var req createScreenRequest

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

func (h *Handler) GetTheaters(w http.ResponseWriter, r *http.Request) {
	t, err := h.svc.GetTheaters(r.Context())

	if err != nil {
		log.Printf("getting theater: %v", err)
		httpx.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, t)
}

func (h *Handler) GetScreen(w http.ResponseWriter, r *http.Request) {
	layout, err := h.svc.GetScreen(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		log.Printf("getting screen: %v", err)
		httpx.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, layout)
}
