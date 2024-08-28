.PHONY: build
build:
	rm -rf bin/ && mkdir -p bin/
	go build -ldflags "-s -w" -o "bin/hptester_m3" "github.com/Xib1uvXi/xholepunch/cmd/hptester"

.PHONY: build_server
build_server:
	rm -rf bin/ && mkdir -p bin/
	GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o "bin/hpserver" "github.com/Xib1uvXi/xholepunch/cmd/hpserver"

.PHONY: build_c2
build_c2:
	rm -rf bin/ && mkdir -p bin/
	GOOS=linux GOARCH=arm64 go build -ldflags "-s -w" -o "bin/hptester_c2" "github.com/Xib1uvXi/xholepunch/cmd/hptester"