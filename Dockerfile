FROM --platform=$BUILDPLATFORM golang:1.23 AS build-env

ARG TARGETPLATFORM
ARG BUILDPLATFORM
ARG TARGETARCH
ARG TARGETOS

RUN echo "I am running on $BUILDPLATFORM, building for $TARGETPLATFORM" > /log

ENV CGO_ENABLED=0
ENV GOOS=${TARGETOS}
ENV GOARCH=${TARGETARCH}

ADD . /src
WORKDIR /src
RUN go mod download
RUN make init
RUN make gen
RUN go build -o /app

FROM scratch
COPY --from=build-env /app /
COPY sources /sources
ENTRYPOINT ["/app"]
