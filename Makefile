.PHONY: generate
generate:
	protoc \
		-I proto \
		-I third_party/grpc-gateway/ \
		-I third_party/googleapis \
		--go_out=plugins=grpc,paths=source_relative:./proto \
		--grpc-gateway_out=./proto \
		--openapiv2_out=third_party/OpenAPI/ \
		proto/example.proto

	mv ./proto/github.com/Percona-Lab/percona-version-service/proto/* ./proto/
	rm -r ./proto/github.com

	statik -m -f -src third_party/OpenAPI/

.PHONY: install
install:
	go get \
		github.com/golang/protobuf/protoc-gen-go \
		github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway \
		github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2 \
		github.com/rakyll/statik
