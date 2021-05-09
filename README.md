# circuitbreaker
This is a go implementation of the  Circuit Breaker pattern.

Imagine you have two services , the order and payment services: 

![Order and Payment Service REST](images/request_diagram.svg)

Now, suppose the Order service needs to make a request to the payment service, to validate customer payment. 

There two kinds of failure that can occur in this situation:

- Short Intermittent Failure: a failure that happens from time to time, for example if your service SLA is 99% availablility then there is 1% chance a request could fail. In this case, retrying the request can remedy the situation.

- Long Unanticipated Failures: this could be because the payment service is down an deployment issue, or because it's experiencing a heavy load. 

In second case, a simple retry wont help since this failure may last for a little while. In fact, suppose the order service blocks on these call. Your server can only do so many requests at a time, if each request is slowed down with the payment request, and you are heavily dependent on it , then after a while your service wont be able to accept anymore requests because of your threads are busy waiting for the payment service. Therefore, now the order service is down / overloaded. To avoid this situation, it would have been better for the order service to not make anymore request to the payment service for a bit of time. It could simply fail the request immediately , in a different situation you might opt out to returning a cached value while you wait for the payment service to recover. 

The Circuit breaker pattern, helps alleviate by  acting as a proxy or a request interceptor. This request intercepter uses a state machine that looks like the following diagram:


![State Machine](images/state_machine.svg)


For more about this pattern, visit this [Blog by Microsoft](https://docs.microsoft.com/en-us/azure/architecture/patterns/circuit-breaker).

## Install
```sh
go get -u github.com/aelnahas/circuitbreaker
```

## Examples

## Settings Options

