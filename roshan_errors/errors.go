package roshan_errors

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// https://connectrpc.com/docs/protocol/#http-to-error-code
var (
	// ErrInvalidInput        = errors.New("invalid input")
	// ErrInternalServerError = errors.New("internal server error")

	ErrInvalidRequestMsg = "invalid request"
	ErrInvalidRequest    = status.Error(codes.InvalidArgument, ErrInvalidRequestMsg)

	ErrInternalServerErrorMsg = "internal server error"
	ErrInternalServerError    = status.Error(codes.Internal, ErrInternalServerErrorMsg)

	ErrInvalidTokenMsg = "invalid token"
	ErrInvalidToken    = status.Error(codes.Unauthenticated, ErrInvalidRequestMsg)

	ErrWrongTokenFormatMsg = "wrong token format"
	ErrWrongTokenFormat    = status.Error(codes.Unauthenticated, ErrWrongTokenFormatMsg)

	ErrMissingTokenMsg = "missing token"
	ErrMissingToken    = status.Error(codes.Unauthenticated, ErrMissingTokenMsg)

	ErrAuthFailedMsg = "authentication failed"
	ErrAuthFailed    = status.Error(codes.Unauthenticated, ErrAuthFailedMsg)
)
