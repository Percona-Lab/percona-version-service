FROM golang:1.17 AS build-env
ADD . /src
ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64
WORKDIR /src
RUN make init
RUN make gen
RUN go mod download
RUN go build -o /app

FROM scratch
COPY --from=build-env /app /
COPY sources /sources
ENTRYPOINT ["/app"]
