package auth

import (
	"context"
	"net/http"
	"slices"
	"strings"

	"github.com/amanuel-tk/Booking-Api/internal/httpx"
)

type ctxKey struct{}

func ClaimsFromContext(ctx context.Context) (*Claims, bool) {
	c, ok := ctx.Value(ctxKey{}).(*Claims)
	return c, ok
}

func (m *JWTManager) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")

		if !ok {
			httpx.WriteError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		claims, err := m.Parse(raw)

		if err != nil {
			httpx.WriteError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		ctx := context.WithValue(r.Context(), ctxKey{}, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func RequireRole(allowed ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := ClaimsFromContext(r.Context())
			if !ok {
				httpx.WriteError(w, http.StatusUnauthorized, "unauthorized")
				return
			}

			for _, role := range claims.Roles {
				if slices.Contains(allowed, role) {
					next.ServeHTTP(w, r)
					return
				}

				httpx.WriteError(w, http.StatusForbidden, "forbidden")
			}
		})
	}
}
