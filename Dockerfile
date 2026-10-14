FROM --platform=$BUILDPLATFORM golang:1.25 AS build-env
ARG TARGETOS
ARG TARGETARCH
ADD . /src
WORKDIR /src
RUN go mod download
RUN make init
RUN make gen
RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-amd64} go build -o /app

FROM scratch
COPY --from=build-env /app /
COPY sources /sources
ENTRYPOINT ["/app"]
