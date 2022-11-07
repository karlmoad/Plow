.PHONY: all ref build dist
.DEFAULT_GOAL := all

DIST_PATH := build

all: ref build

ref:
	go get

build: clean
	go build -o $(DIST_PATH)/plow ./main.go

linux:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -a -installsuffix cgo -o $(DIST_PATH)/linux/plow ./main.go

mac:
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -a -installsuffix cgo -o $(DIST_PATH)/mac/plow ./main.go

windows:
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -a -installsuffix cgo -o $(DIST_PATH)/windows/plow ./main.go

clean:
	rm -rf $(DIST_PATH)/*