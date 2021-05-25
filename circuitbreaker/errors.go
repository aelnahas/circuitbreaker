package circuitbreaker

import (
	"fmt"
)

//ErrInvalidSettingParam gets thrown when a setting value is not valid
type ErrInvalidSettingParam struct {
	Param string
	Val   interface{}
}

func (eisp ErrInvalidSettingParam) Error() string {
	return fmt.Sprintf("invalid setting %s value %v", eisp.Param, eisp.Val)
}

//ErrRequestNotPermitted gets thrown when a caller attempts to make a request while the circuit breaker is
// open
type ErrRequestNotPermitted struct {
	State State
	Name  string
}

func (ernp ErrRequestNotPermitted) Error() string {
	return fmt.Sprintf("circuit breaker not permitting requests, name : %s, state: %s", ernp.Name, ernp.State)
}
