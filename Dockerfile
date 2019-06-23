FROM golang:1.12.6-alpine3.10

RUN apk add --no-cache git

ARG GOOS
ARG GOARCH
ENV GOOS linux
ENV GOARCH amd64
ENV GO111MODULE on
ENV CGO_ENABLED 0

WORKDIR /go/src
ADD . .
RUN go get -u
RUN echo $(env)
RUN go build -a -v -o dynflare *.go

FROM alpine:3.10

RUN apk add --no-cache ca-certificates
WORKDIR /root
COPY --from=0 /go/src/dynflare .

ENTRYPOINT ["./dynflare"]