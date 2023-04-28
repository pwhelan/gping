FROM golang:alpine as build-base

RUN apk add --no-cache git

WORKDIR /src

COPY go.* ./
RUN go mod download

COPY . ./
RUN CGO_ENABLED=0 go install \
	-ldflags '-extldflags "-static"'\
	-tags timetzdata

FROM scratch
COPY --from=build-base /go/bin/gping /gping
ENTRYPOINT ["/gping"]
