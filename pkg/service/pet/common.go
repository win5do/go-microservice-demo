package pet

import (
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/jinzhu/gorm"
	errors2 "github.com/pkg/errors"
	"github.com/prometheus/common/log"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/win5do/golang-microservice-demo/pkg/api/errcode"
)

func pberr(err error) error {
	switch {
	case errors2.Is(err, errcode.Err_forbidden):
		return status.Error(codes.PermissionDenied, err.Error())
	case errors2.Is(err, errcode.Err_not_found),
		errors2.Is(err, gorm.ErrRecordNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors2.Is(err, errcode.Err_invalid_params):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors2.Is(err, errcode.Err_conflict):
		return status.Error(codes.FailedPrecondition, err.Error())
	}

	return status.Error(codes.Internal, err.Error())
}

func time2Pb(in time.Time) *timestamp.Timestamp {
	r, err := ptypes.TimestampProto(in)
	if err != nil {
		log.Errorf("err: %+v", err)
		return nil
	}

	return r
}

func pb2Time(in *timestamp.Timestamp) time.Time {
	if in == nil {
		return time.Time{}
	}

	r, err := ptypes.Timestamp(in)
	if err != nil {
		log.Errorf("err: %+v", err)
		return time.Time{}
	}

	return r
}
