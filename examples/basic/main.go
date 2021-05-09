package main

import (
	"log"
	"net/http"

	"github.com/aelnahas/circuitbreaker/examples/basic/orders"
	"github.com/aelnahas/circuitbreaker/examples/basic/payments"
	"github.com/go-chi/chi"
)

func main() {
	r := chi.NewRouter()
	orderSrvc := orders.NewService()
	paymentSrvc := payments.NewService()

	r.Post("/orders", orders.CreateOrder(orderSrvc))

	r.Route("/payments", func(r chi.Router) {
		r.Post("/", payments.NewTransaction(paymentSrvc))
		r.Get("/failure-rate", payments.GetFailureRate(paymentSrvc))
		r.Put("/failure-rate", payments.SetFailureRate(paymentSrvc))
	})

	log.Panicln(http.ListenAndServe(":3000", r))
}
