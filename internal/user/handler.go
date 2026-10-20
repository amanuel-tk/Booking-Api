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

// Register godoc
// @Summary      Register a new user
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        body  body      RegisterRequest  true  "New account"
// @Success      201   {object}  User
// @Failure      400   {object}  httpx.ErrorResponse  "Validation error"
// @Failure      409   {object}  httpx.ErrorResponse  "Email already registered"
// @Failure      500   {object}  httpx.ErrorResponse
// @Router       /auth/register [post]
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

// Login godoc
// @Summary      Log in
// @Description  Validates credentials and returns a 15-minute JWT access token.
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        body  body      LoginRequest  true  "Credentials"
// @Success      200   {object}  LoginResponse
// @Failure      400   {object}  httpx.ErrorResponse  "Invalid request body"
// @Failure      401   {object}  httpx.ErrorResponse  "Invalid email or password"
// @Failure      500   {object}  httpx.ErrorResponse
// @Router       /auth/login [post]
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if !httpx.DecodeJson(w, r, &req) {
		return
	}

	token, u, err := h.svc.Login(r.Context(), req)

	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			httpx.WriteError(w, http.StatusUnauthorized, "invalid email or password")
			return
		}
		log.Printf("login %v", err)
		httpx.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	httpx.WriteJSON(w, http.StatusOK, LoginResponse{
		AccessToken: token,
		TokenType:   "Bearer",
		User:        *u,
	})
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": "logged in"})
}
