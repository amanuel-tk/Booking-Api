package httpx

import (
	"encoding/json"
	"log"
	"net/http"
)

func DecodeJson(w http.ResponseWriter, r *http.Request, dst any) bool {

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		log.Printf("decoding json: %v", err)
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return false
	}
	return true
}
