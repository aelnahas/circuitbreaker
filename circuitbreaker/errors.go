package circuitbreaker

import (
	"errors"
	"fmt"
)

type ErrInvalidSettingParam struct {
	Param string
	Val   interface{}
}

func (eisp ErrInvalidSettingParam) Error() string {
	return fmt.Sprintf("invalid setting %s value %v", eisp.Param, eisp.Val)
}

type ErrRequestNotPermitted struct {
	State State
	Name  string
}

func (ernp ErrRequestNotPermitted) Error() string {
	return fmt.Sprintf("circuit breaker not permitting requests, name : %s, state: %s", ernp.Name, ernp.State)
}

var ErrEmptyMeasurements = errors.New("can read past record, no measurements taken")
