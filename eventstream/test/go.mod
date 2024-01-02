module github.com/fullstorydev/go/eventstream/test

go 1.18

require (
	github.com/fullstorydev/go/eventstream v0.0.0
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c
	gotest.tools/v3 v3.0.3
)

require (
	github.com/google/go-cmp v0.4.0 // indirect
	github.com/pkg/errors v0.8.1 // indirect
)

replace github.com/fullstorydev/go/eventstream => ./..
