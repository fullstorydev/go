.PHONY: ci
ci: deps checkgofmt vet staticcheck ineffassign predeclared errcheck test

.PHONY: deps
deps:
	go get -d -v -t ./...

.PHONY: updatedeps
updatedeps:
	go get -d -v -t -u -f ./...

.PHONY: checkgofmt
checkgofmt:
	gofmt -s -l .
	@if [ -n "$$(gofmt -s -l .)" ]; then \
		exit 1; \
	fi

.PHONY: vet
vet:
	go vet -composites=false ./...

.PHONY: staticcheck
staticcheck:
	@go install honnef.co/go/tools/cmd/staticcheck@v0.3.3
	staticcheck ./...

.PHONY: ineffassign
ineffassign:
	@go install github.com/gordonklaus/ineffassign@v0.0.0-20220928193011-d2c82e48359b
	ineffassign .

.PHONY: predeclared
predeclared:
	@go install github.com/nishanths/predeclared@v0.2.2
	predeclared ./...

# Intentionally omitted from CI, but target here for ad-hoc reports.
.PHONY: golint
golint:
	@go install golang.org/x/lint/golint@v0.0.0-20210508222113-6edffad5e616
	golint -min_confidence 0.9 -set_exit_status ./...

.PHONY: errcheck
errcheck:
	@go install github.com/kisielk/errcheck@v1.6.2
	errcheck ./...

.PHONY: test
test:
	$(MAKE) -C test
