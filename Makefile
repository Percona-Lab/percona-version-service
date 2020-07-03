init:
	go build -modfile=tools/go.mod -o bin/yq github.com/mikefarah/yq/v3
	go build -modfile=tools/go.mod -o bin/statik github.com/rakyll/statik

	curl -L https://github.com/uber/prototool/releases/download/v1.8.0/prototool-$(shell uname -s)-$(shell uname -m) -o ./bin/prototool
	chmod +x ./bin/prototool

gen:
	./bin/prototool all ./api

	bin/yq r --prettyPrint third_party/OpenAPI/api/version.swagger.json > third_party/OpenAPI/api/version.swagger.yaml
	rm third_party/OpenAPI/api/version.swagger.json

	mv ./versionpb/github.com/Percona-Lab/percona-version-service/version/* ./versionpb/
	rm -r ./versionpb/github.com

	bin/statik -m -f -src third_party/OpenAPI/

cert:
	mkcert -cert-file=certs/cert.pem -key-file=certs/key.pem 0.0.0.0

image: gen
	scripts/build.sh
