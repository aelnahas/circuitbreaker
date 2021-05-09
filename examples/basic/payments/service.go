package payments

import (
	"errors"
	"math/rand"

	"github.com/google/uuid"
)

const (
	MinFailureRate = 0
	MaxFailureRate = 100
)

type Service interface {
	SetFailureRate(rate int)
	FailureRate() int
	NewTransaction() (string, error)
}

type service struct {
	failureRate int
}

func NewService() Service {
	return &service{
		failureRate: 0,
	}
}

func (s *service) SetFailureRate(rate int) {
	s.failureRate = rate
}

func (s *service) FailureRate() int {
	return s.failureRate
}

func (s *service) NewTransaction() (string, error) {
	if rand.Intn(MaxFailureRate-MinFailureRate)+MinFailureRate < s.failureRate {
		return "", errors.New("failed")
	}

	return uuid.NewString(), nil
}
