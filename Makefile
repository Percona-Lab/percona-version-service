SHELL = /bin/bash

GIT_BRANCH:=$(shell git rev-parse --abbrev-ref HEAD | sed -e 's^/^-^g; s^[.]^-^g;' | tr '[:upper:]' '[:lower:]')
GIT_COMMIT:=$(shell git rev-parse --short HEAD)
IMG ?= perconalab/version-service:$(GIT_BRANCH)-$(GIT_COMMIT)
PLATFORMS ?= linux/amd64,linux/arm64

init:
	mkdir -p bin
	go build -modfile=tools/go.mod -o bin/yq github.com/mikefarah/yq/v3
	go build -modfile=tools/go.mod -o bin/protoc-gen-go google.golang.org/protobuf/cmd/protoc-gen-go
	go build -modfile=tools/go.mod -o bin/protoc-gen-go-grpc google.golang.org/grpc/cmd/protoc-gen-go-grpc
	go build -modfile=tools/go.mod -o bin/protoc-gen-grpc-gateway github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway
	go build -modfile=tools/go.mod -o bin/protoc-gen-openapiv2 github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2

	curl -L "https://github.com/bufbuild/buf/releases/download/v1.34.0/buf-$(shell uname -s)-$(shell uname -m)" -o "./bin/buf"
	chmod +x ./bin/buf

	SWAGGER_ARCH=$$(uname -m); \
	case "$${SWAGGER_ARCH}" in \
		x86_64) SWAGGER_ARCH=amd64 ;; \
		aarch64|arm64) SWAGGER_ARCH=arm64 ;; \
		*) echo "unsupported swagger architecture: $${SWAGGER_ARCH}"; exit 1 ;; \
	esac; \
	curl -L "https://github.com/go-swagger/go-swagger/releases/download/v0.31.0/swagger_$(shell uname | tr '[:upper:]' '[:lower:]')_$${SWAGGER_ARCH}" -o ./bin/swagger
	chmod +x ./bin/swagger

gen:
	bin/buf dep update

	bin/buf generate

	bin/yq r --prettyPrint third_party/OpenAPI/api/version.swagger.json > third_party/OpenAPI/api/version.swagger.yaml
	rm third_party/OpenAPI/api/version.swagger.json
	cp third_party/OpenAPI/api/version.swagger.yaml api/


	rm -rf ./client
	./bin/swagger generate client -m client/models -f ./api/version.swagger.yaml -t ./

build-format-release-notes:
	go build -race -o bin/format-release-notes ./cmd/format-release-notes

format-release-notes:
	./bin/format-release-notes --dir=sources/release-notes/pmm

cert:
	mkcert -cert-file=certs/cert.pem -key-file=certs/key.pem 0.0.0.0

# Build docker image
docker-build:
	docker build --platform=linux/amd64 . -t ${IMG}

# Run docker image
docker-run-it:
	docker run -it --rm -p 10000:10000 -p 11000:11000 -e SERVE_HTTP=true ${IMG}

build:
	CGO_ENABLED=0 go build -a -o ./bin/app

run: build
	SERVE_HTTP=true ./bin/app

# Build and push multi-arch docker image
docker-push:
	docker buildx build --platform=$(PLATFORMS) -t ${IMG} $(if $(IMG_LATEST),-t $(IMG_LATEST)) --push .

test:
	docker-compose -f docker-compose.test.yml up --build --abort-on-container-exit
	docker-compose -f docker-compose.test.yml down
