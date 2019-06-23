TAG ?= "trapped/dynflare:$(shell git rev-parse --short HEAD)"

all: linux-amd64 linux-arm darwin-amd64

linux-amd64:
	docker build -t $(TAG) --build-arg=GOOS=linux --build-arg=GOARCH=amd64 .

linux-arm:
	docker build -t $(TAG)-arm --build-arg=GOOS=linux --build-arg=GOARCH=arm .

darwin-amd64:
	docker build -t $(TAG)-darwin --build-arg=GOOS=darwin --build-arg=GOARCH=amd64 .
