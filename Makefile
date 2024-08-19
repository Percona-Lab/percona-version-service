SHELL = /bin/bash

GIT_BRANCH:=$(shell git rev-parse --abbrev-ref HEAD | sed -e 's^/^-^g; s^[.]^-^g;' | tr '[:upper:]' '[:lower:]')
GIT_COMMIT:=$(shell git rev-parse --short HEAD)
IMG ?= perconalab/version-service:$(GIT_BRANCH)-$(GIT_COMMIT)

init:
	go build -modfile=tools/go.mod -o bin/yq github.com/mikefarah/yq/v3
	go build -modfile=tools/go.mod -o bin/protoc-gen-go google.golang.org/protobuf/cmd/protoc-gen-go
	go build -modfile=tools/go.mod -o bin/protoc-gen-go-grpc google.golang.org/grpc/cmd/protoc-gen-go-grpc
	go build -modfile=tools/go.mod -o bin/protoc-gen-grpc-gateway github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway
	go build -modfile=tools/go.mod -o bin/protoc-gen-openapiv2 github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2

	curl -L "https://github.com/bufbuild/buf/releases/download/v1.34.0/buf-$(shell uname -s)-$(shell uname -m)" -o "./bin/buf"
	chmod +x ./bin/buf

	curl -L  https://github.com/go-swagger/go-swagger/releases/download/v0.31.0/swagger_$(shell uname | tr '[:upper:]' '[:lower:]')_amd64 -o ./bin/swagger
	chmod +x ./bin/swagger

	go build -race -o bin/format-release-notes ./cmd/format-release-notes

gen:
	bin/buf dep update

	bin/buf generate

	bin/yq r --prettyPrint third_party/OpenAPI/api/version.swagger.json > third_party/OpenAPI/api/version.swagger.yaml
	rm third_party/OpenAPI/api/version.swagger.json
	cp third_party/OpenAPI/api/version.swagger.yaml api/


	rm -rf ./client
	./bin/swagger generate client -m client/models -f ./api/version.swagger.yaml -t ./

format-release-notes:
	./bin/format-release-notes --dir=docs/release-notes

cert:
	mkcert -cert-file=certs/cert.pem -key-file=certs/key.pem 0.0.0.0

# Build docker image
docker-build:
	docker build --platform=linux/amd64 . -t ${IMG}

# Run docker image
docker-run-it:
	docker run -it --rm -p 10000:10000 -p 11000:11000 -e SERVE_HTTP=true ${IMG}

build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o ./bin/app

run: build
	SERVE_HTTP=true ./bin/app

# Build and push docker image
docker-push: docker-build
	docker push ${IMG}

test:
	docker-compose -f docker-compose.test.yml up --build --abort-on-container-exit
	docker-compose -f docker-compose.test.yml down
