package orders

import (
	"encoding/json"
	"net/http"
)

func CreateOrder(s Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		order, err := s.NewOrder()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		} else {
			json.NewEncoder(w).Encode(order)
		}
	}
}
