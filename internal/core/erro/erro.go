package erro

import (
	"errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ErrNotFound 		= errors.New("item not found")
	MissingData 	= status.Errorf(codes.InvalidArgument, "header missing metadata")
)