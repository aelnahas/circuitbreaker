# circuitbreaker
This is a go implementation of the  Circuit Breaker pattern.

![Order and Payment Service REST](images/request_diagram.svg)

Imagine that you have two services, the orders and payments services as shown on the diagram above. Now, suppose the Orders service needs to make a request to the Payments service, for example, to validate customer payment. There are potentially two kinds of failure that can occur in this situation:

- Short Intermittent Failures: These happen from time to time and are somewhate expected. For instance, if your service SLA maintains a 99% availablility, then there is a 1% chance any request could fail. In this case, retrying the request can remedy the situation.

- Long Unanticipated Failures: this could be because the payment service is down due to deployment issue, or some production issue. In this situation retrying request probably wont help because of the nature and duration of the failure. 

In fact, suppose that the orders service blocks on these calls. Now, The orders server can only do so many requests at a time. In the worst case, every one of your request to the orders service needs to make a failing call to the payment service. Eventually, you might run into a sitation where all the resources in the Order service have been exhausted and it cant take any more requests. As a result, the orders service is effectively down, resulting in this cascading failure. 

To avoid this situation, it would have been better for the orders service to not make anymore request to the payment service, and allow the payment service some timeout to recover. The orders service could in the meantime simply fail the request immediately, or in a different situation return a cached value. This in a nutshell is the Circuit Breaker Pattern.

The Circuit breaker pattern, helps alleviate failing requests by acting as a request interceptor. The Interceptor manages the requests by using a state machine that looks like the following diagram:

![State Machine](images/state_machine.svg)

As we can see there are 3 states, and they are described as follows:
- `Closed` State: 
    - This is the normal state where requests against the payment service are being executed.
- `Open` State: 
    - the Interceptor enters this state when the calls to the payment service have failed too many times. 
    - Once it is in this state, no request to the Payment service would be executed for some timeout duration, also known in my implementation as `CooldownDuration`.
    - The criteria for this failure can vary from one implementation to another. The simplest way is to count successive failures and `open` the circuit if it is over a certain amount. 
    - The implementation in this repo, attempts to calculate the failure rate over the duration of the last N calls and check if it is higher than a configured threshold `FailureRate`.
- `Half-Open` State:
    - Once the `CooldownDuration` has expired on the `open` state the Interceptor transition into this state. The idea here is to test the waters and `verify` that the payment service has actually recovered.
    - Only a subset of request goes through ( say 10 % of the normal request window ) - if the failure continues or the success is not yet on par with what we hope, we send the Interceptor back to the `Open` state
    - Otherwise, if the request are passing and they have recovered over some `RecoveryRate` threshold, then we can assume that the payment service has recovered. In this case, transition the interceptor back to the `closed` state.


To read more about this pattern, visit this [Blog by Microsoft](https://docs.microsoft.com/en-us/azure/architecture/patterns/circuit-breaker).

## Install
```sh
go get -u github.com/aelnahas/circuitbreaker
```

## Examples
A basic usage can be seen in the [Basic Example](examples/basic) implementation


## Settings

### Thresholds

There are thresholds being used in the implementation to guide the interceptor throught the states. The following table describes them:


| Name | Description | Type | Default | SettingsOption |
| --- | --- | --- | --- | --- |
| FailureRate | Acceptable rate of failures , if exceeded interceptor will switch to `Open` state | float64 | 10.0 | WithFailureRate |
| RecoveryRate | in `Half-Open` state, this threshold is used to determine if the service being requested has "recovered" and it is okay to go back to `Closed` state | float64 | 10.0 | WithRecoveryRate |
| CooldownDuration | the `duration` where the interceptor will remain in `open` and not forward any requests | time.Duration | 30 seconds | WithCooldownDuration |
| MaxRequestOnHalfOpen | Number of requests that are allowed to be executed in Half-open state | int | 10 | WithMaxRequestOnHalfOpen |


### WindowSize
 The implementation currently uses a fixed sliding window to collect metrics on the requests,and it builds an aggregate based on that. The size of this window is controlled by the `WindowSize` settings. By default, it is set to 100. 

 This means the last `100` calls will be used to drive the aggregates, and older metrics are simply dropped.

 can be modified by passing `circuitbreaker.WithWindowSize` `SettingsOption` to the  `circuitbreaker.NewSettings` constructor

### IsSuccessful
This is a callback handler that is used when the interceptor receives a response. This allows the caller to determine set their conditions for `success`. For instance, perhaps the service responds with a json that has more info on the error.

By default if nothing is passed, the settings will use the following method to check if there is an error returned from the request.

```go
func DefaultIsSuccessful(resp *http.Response, err error) bool {
	return err == nil
}
```

 can be modified by passing `circuitbreaker.WithIsSuccessfulHandler` `SettingsOption` to the  `circuitbreaker.NewSettings` constructor



### OnStateChange
another callback that will be called whenever the intercepter switches states.

You can use it to add your logs, or switch to cached values

e.g.

```go
func (s *service) OnStateChange(name string, from circuitbreaker.State, to circuitbreaker.State, metrics circuitbreaker.Metrics) {
	s.logger.Printf("intercepter %s metrics %+v\n", name, metrics)
	s.logger.Printf("intercepter %s transitioning from %s to %s\n", name, from, to)

}
```

 can be modified by passing `circuitbreaker.WithOnStateChangeHandler` `SettingsOption` to the  `circuitbreaker.NewSettings` constructor
