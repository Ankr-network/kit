.PHONY: proto
proto:
	protoc -I. -I${GOPATH}/src -I${GOPATH}/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
		--go_out=. --go_opt=paths=source_relative *.proto;
