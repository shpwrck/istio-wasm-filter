.PHONY: all
all: build docker

.PHONY: build
build:
	tinygo build -o main.wasm -scheduler=none -target=wasi ./main.go

.PHONY: docker
docker:
	tag=latest;\
        reg=ghcr.io/shpwrck/istio-wasm-filter:$$tag;\
	docker build --tag $$reg .;\
	docker push $$reg;

