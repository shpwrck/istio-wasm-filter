## WASM Plugin for Istio

### What it does?

This filter will inject a service call in the normal request flow to gather information and translate based on given configuration.
The example currently uses an external `httpbin` service to map from values `Org` and `Product` to `X-Org` and `X-Product`.

![architecture](./architecture.png)

### What is required?

* make
* tinygo
* docker
* kind
* kubectl
* istioctl

### How to test?

`make deploy`


