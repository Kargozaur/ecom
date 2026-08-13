.PHONY: build-gateway

build-gateway:
	go build -o gateway ./svc/gateway
