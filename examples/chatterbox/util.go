package chatterbox

import (
	"context"
	"io"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// filterErr cleans up error logging by filtering out errors related to (probably user initiated) cancel.
func filterErr(err error) error {
	if err == io.EOF || err == context.Canceled {
		return nil
	}
	if code := status.Code(err); code == codes.Canceled {
		return nil
	}
	return err
}
