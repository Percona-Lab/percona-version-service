generate:
	protoc \
		-I versionpb \
		-I third_party/grpc-gateway/ \
		-I third_party/googleapis \
		--go_out=plugins=grpc,paths=source_relative:./versionpb \
		--grpc-gateway_out=./versionpb \
		--openapiv2_out=third_party/OpenAPI/ \
		versionpb/version.proto

	mv ./versionpb/github.com/Percona-Lab/percona-version-service/proto/* ./versionpb/
	rm -r ./versionpb/github.com

	statik -m -f -src third_party/OpenAPI/

.PHONY: install
install:
	go get \
		github.com/golang/protobuf/protoc-gen-go \
		github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway \
		github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2 \
		github.com/rakyll/statik

cert:
	mkcert -cert-file=certs/cert.pem -key-file=certs/key.pem 0.0.0.0
