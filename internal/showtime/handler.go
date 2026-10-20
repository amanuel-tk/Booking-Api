package showtime

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
	return &Handler{svc: svc}
}

// CreateShowtime godoc
// @Summary      Create a showtime
// @Description  Schedule a movie on a screen. The start time must be in the future. The end time is calculated from the movie duration plus a 15-minute cleaning buffer. Overlapping showtimes on the same screen are rejected. Requires admin access.
// @Tags         Showtimes
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body CreateShowtimeRequest true "Showtime details"
// @Success      201 {object} Showtime
// @Failure      400 {object} httpx.ErrorResponse
// @Failure      401 {object} httpx.ErrorResponse
// @Failure      403 {object} httpx.ErrorResponse
// @Failure      404 {object} httpx.ErrorResponse
// @Failure      409 {object} httpx.ErrorResponse
// @Failure      500 {object} httpx.ErrorResponse
// @Router       /showtimes [post]
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

// GetShowtime godoc
// @Summary      Get a showtime
// @Description  Retrieve an active showtime by its UUID, including its movie, screen, start time, end time, and ticket price.
// @Tags         Showtimes
// @Produce      json
// @Param        id path string true "Showtime UUID" format(uuid)
// @Success      200 {object} Showtime
// @Failure      404 {object} httpx.ErrorResponse
// @Failure      500 {object} httpx.ErrorResponse
// @Router       /showtimes/{id} [get]
func (h *Handler) GetShowtime(w http.ResponseWriter, r *http.Request) {
	st, err := h.svc.GetShowtime(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, st)
}

// ListByMovie godoc
// @Summary      List movie showtimes
// @Description  Retrieve upcoming showtimes for a movie within a date range. Dates use YYYY-MM-DD. The "date" parameter is an alternative to "from". If no dates are provided, today's date is used. The range is limited to 31 days.
// @Tags         Showtimes
// @Produce      json
// @Param        id path string true "Movie UUID" format(uuid)
// @Param        from query string false "Start date (YYYY-MM-DD)" example(2026-12-20)
// @Param        to query string false "End date, inclusive (YYYY-MM-DD)" example(2026-12-25)
// @Param        date query string false "Alternative start date if from is omitted (YYYY-MM-DD)" example(2026-12-20)
// @Success      200 {array} ListItem
// @Failure      400 {object} httpx.ErrorResponse
// @Failure      500 {object} httpx.ErrorResponse
// @Router       /movies/{id}/showtimes [get]
func (h *Handler) ListByMovie(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	from := q.Get("from")
	if from == "" {
		from = q.Get("date")
	}
	items, err := h.svc.ListByMovie(r.Context(), chi.URLParam(r, "id"), from, q.Get("to"))

	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, items)
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
