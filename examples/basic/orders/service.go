package orders

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/aelnahas/circuitbreaker/circuitbreaker"
)

type Order struct {
	TransactionID string
}

type Service interface {
	NewOrder() (*Order, error)
}

type service struct {
	interceptor *circuitbreaker.Breaker
	logger      *log.Logger
}

func NewService() Service {
	s := &service{}
	// create the settings by declaring the name of the intercepter, and also
	// passing in the handlers to determine if the request is successful, and to alert the service of a state change
	settings, err := circuitbreaker.NewSettings("Orders.Payments",
		circuitbreaker.WithIsSuccessfulHandler(s.isSuccessful),
		circuitbreaker.WithOnStateChangeHandler(s.OnStateChange),
	)
	if err != nil {
		panic(err)
	}

	// we finally make a new intercepter
	interceptor, err := circuitbreaker.NewBreakerWithSettings(settings)
	if err != nil {
		panic(err)
	}
	s.logger = log.New(os.Stderr, "orders\t", log.LstdFlags)
	s.interceptor = interceptor
	return s
}

func (s *service) NewOrder() (*Order, error) {
	order := &Order{}

	// here we pass the request handler to Execute
	resp, err := s.interceptor.Execute(s.requestTransaction)
	if err != nil {
		s.logger.Printf("request to payments failed %s\n", err.Error())
		return nil, err
	}

	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}

	order.TransactionID = string(data)
	return order, nil

}

// Will be called when the interceptor determins that it is okay to make the request
func (s *service) requestTransaction(name string) (*http.Response, error) {
	body, err := json.Marshal((map[string]string{}))
	if err != nil {
		return nil, err
	}
	s.logger.Println("sending post request")
	return http.Post("http://localhost:3000/payments", "application/json", bytes.NewBuffer(body))
}

// the handler to check if the request is successful.
func (s *service) isSuccessful(res *http.Response, err error) bool {
	return err == nil && res.StatusCode < 500
}

func (s *service) OnStateChange(name string, from circuitbreaker.State, to circuitbreaker.State) {
	s.logger.Printf("intercepter %s transitioning from %s to %s\n", name, from, to)
}
