package payments

import (
	"encoding/json"
	"net/http"
)

type FailureRateRequest struct {
	FailureRate int `json:"failure_rate"`
}

func NewTransaction(s Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uuid, err := s.NewTransaction()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		} else {
			w.Write([]byte(uuid))
		}
	}
}

func SetFailureRate(s Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var fr FailureRateRequest
		if err := json.NewDecoder(r.Body).Decode(&fr); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		} else {
			s.SetFailureRate(fr.FailureRate)
		}
	}
}

func GetFailureRate(s Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fr := FailureRateRequest{
			FailureRate: s.FailureRate(),
		}

		json.NewEncoder(w).Encode(&fr)
	}
}
