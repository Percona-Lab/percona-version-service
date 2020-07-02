init:
	go build -modfile=tools/go.mod -o bin/yq github.com/mikefarah/yq/v3
	go build -modfile=tools/go.mod -o bin/statik github.com/rakyll/statik

generate:
	protoc \
		-I api \
		-I third_party/grpc-gateway/ \
		-I third_party/googleapis \
		--go_out=plugins=grpc,paths=source_relative:./versionpb \
		--grpc-gateway_out=./versionpb \
		--openapiv2_out=third_party/OpenAPI/ \
		api/version.proto

	bin/yq r --prettyPrint third_party/OpenAPI/version.swagger.json > third_party/OpenAPI/version.swagger.yaml
	rm third_party/OpenAPI/version.swagger.json

	mv ./versionpb/github.com/Percona-Lab/percona-version-service/proto/* ./versionpb/
	rm -r ./versionpb/github.com

	bin/statik -m -f -src third_party/OpenAPI/

cert:
	mkcert -cert-file=certs/cert.pem -key-file=certs/key.pem 0.0.0.0

image: generate
	scripts/build.sh
