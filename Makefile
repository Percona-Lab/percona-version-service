SHELL = /bin/bash

init:
	go build -modfile=tools/go.mod -o bin/yq github.com/mikefarah/yq/v3
	go build -modfile=tools/go.mod -o bin/statik github.com/rakyll/statik
	go build -modfile=tools/go.mod -o tools/bin/modvendor github.com/goware/modvendor

	curl -L https://github.com/uber/prototool/releases/download/v1.10.0/prototool-$(shell uname -s)-$(shell uname -m) -o ./bin/prototool
	chmod +x ./bin/prototool

gen:
	pushd ${CURDIR}/tools; \
	go mod vendor; \
	./bin/modvendor -copy="**/*.proto" -v \
		-include="github.com/grpc-ecosystem/grpc-gateway/v2/third_party/googleapis/google/api,github.com/grpc-ecosystem/grpc-gateway/v2/third_party/googleapis/google/rpc"; \
	go install \
		github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway \
		github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2 \
		github.com/golang/protobuf/protoc-gen-go

	./bin/prototool all ./api

	bin/yq r --prettyPrint third_party/OpenAPI/api/version.swagger.json > third_party/OpenAPI/api/version.swagger.yaml
	rm third_party/OpenAPI/api/version.swagger.json
	cp third_party/OpenAPI/api/version.swagger.yaml api/

	mv ./versionpb/github.com/Percona-Lab/percona-version-service/version/* ./versionpb/
	rm -r ./versionpb/github.com

	bin/statik -m -f -src third_party/OpenAPI/

cert:
	mkcert -cert-file=certs/cert.pem -key-file=certs/key.pem 0.0.0.0

image: gen
	scripts/build.sh
