package erro

import (
	"errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ErrCardTypeInvalid	= errors.New("card type invalid")
	ErrNotFound 		= errors.New("item not found")
	ErrUpdate			= errors.New("update unsuccessful")
	InvalidToken   	= status.Errorf(codes.Unauthenticated, "invalid beard token")
	MissingData 	= status.Errorf(codes.InvalidArgument, "header missing metadata")
)