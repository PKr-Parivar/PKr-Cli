grpc-out:
	protoc ./proto/*.proto --go_out=. --go-grpc_out=.

get-new-kcp:
	go get github.com/ButterHost69/kcp-go@latest

upgrade-base:
	@echo Copy Paste this in Terminal -- Don't Run using Make
	$$env:GOPRIVATE="github.com/ButterHost69"
	go get github.com/PKr-Parivar/PKr-Base@latest
