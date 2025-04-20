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

var (
	ErrInvalidToken     = status.Error(codes.Unauthenticated, "invalid token")
	ErrWrongTokenFormat = status.Error(codes.Unauthenticated, "wrong token format")
	ErrMissingToken     = status.Error(codes.Unauthenticated, "missing token")
)
