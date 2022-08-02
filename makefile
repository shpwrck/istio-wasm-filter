.PHONY: all
all: build docker deploy

.PHONY: build
build:
	tinygo build -o main.wasm -scheduler=none -target=wasi ./main.go

.PHONY: docker
docker:
	tag=latest;\
        reg=ghcr.io/shpwrck/istio-wasm-filter:$$tag;\
	docker build --tag $$reg .;\
	docker push $$reg;

.PHONY: deploy
deploy:
	kind create cluster --name istio-wasm-filter;\
	istioctl install -y;\
	kubectl create namespace int-svc;\
	kubectl create namespace ext-svc;\
	kubectl label namespace int-svc istio-injection=enabled;\
	kubectl label namespace ext-svc istio-injection=enabled;\
	kubectl apply -n int-svc -f ./wasmplugin.yaml;\
	kubectl apply -n int-svc -f https://raw.githubusercontent.com/istio/istio/master/samples/sleep/sleep.yaml;\
	kubectl apply -n int-svc -f https://raw.githubusercontent.com/istio/istio/master/samples/httpbin/httpbin.yaml;\
	kubectl apply -n ext-svc -f https://raw.githubusercontent.com/istio/istio/master/samples/httpbin/httpbin.yaml;

.PHONY: test
test:
	kubectl exec -n int-svc deployment/sleep -- curl httpbin:8000/get -H "Org: Field" -H "Product: umbrella"

.PHONY: destroy
destroy:
	kind delete cluster --name istio-wasm-filter

