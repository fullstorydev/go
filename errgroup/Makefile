.PHONY: ci
ci: deps checkgofmt vet staticcheck ineffassign predeclared golint errcheck test

.PHONY: deps
deps:
	go get -v -t ./...
	go mod tidy

.PHONY: updatedeps
updatedeps:
	go get -v -t -u ./...
	go mod tidy

.PHONY: checkgofmt
checkgofmt:
	gofmt -s -l .
	@if [ -n "$$(gofmt -s -l .)" ]; then \
		exit 1; \
	fi

.PHONY: vet
vet:
	go vet

.PHONY: staticcheck
staticcheck:
	@go install honnef.co/go/tools/cmd/staticcheck@v0.8.2
	staticcheck ./...

.PHONY: ineffassign
ineffassign:
	@go install github.com/gordonklaus/ineffassign@v0.2.0
	ineffassign .

.PHONY: predeclared
predeclared:
	@go install github.com/nishanths/predeclared@v0.3.1
	predeclared ./...

.PHONY: golint
golint:
	@go install golang.org/x/lint/golint@v0.0.0-20210508222113-6edffad5e616
	golint -min_confidence 0.9 -set_exit_status ./...

.PHONY: errcheck
errcheck:
	@go install github.com/kisielk/errcheck@v1.30.0
	errcheck ./...

.PHONY: test
test:
	# The race detector requires CGO: https://github.com/golang/go/issues/6508
	CGO_ENABLED=1 go test -race ./...
