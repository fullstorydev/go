all: $(GOPATH)/bin/chatterbox

%.pb.go %_grpc.pb.go: %.proto
	go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.26
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.1
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative $<

$(GOPATH)/bin/chatterbox: chatterbox.pb.go chatterbox_grpc.pb.go $(shell find . -name "*.go" -type f)
	go install ./cmd/chatterbox
