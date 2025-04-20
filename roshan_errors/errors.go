package roshan_errors

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// https://connectrpc.com/docs/protocol/#http-to-error-code
var (
	// ErrInvalidInput        = errors.New("invalid input")
	// ErrInternalServerError = errors.New("internal server error")

	ErrInvalidInput        = status.Error(codes.InvalidArgument, "invalid input")
	ErrInternalServerError = status.Error(codes.Internal, "internal server error")
)
