grpc-out:
	protoc ./proto/*.proto --go_out=. --go-grpc_out=.

get-new-kcp:
	@echo Copy Paste this in Terminal -- Don't Run using Make
	$$env:GOPRIVATE="github.com/PKr-Parivar"
	go get github.com/PKr-Parivar/kcp-go@latest

upgrade-base:
	@echo Copy Paste this in Terminal -- Don't Run using Make
	$$env:GOPRIVATE="github.com/PKr-Parivar"
	go get github.com/PKr-Parivar/PKr-Base@latest
