package grpcapi

import (
	"common/performance"
	"common/response"
	"common/util"
	"errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"sync"
	"time"
)

var (
	connPool = map[string]*grpc.ClientConn{}
	poolLock = sync.Mutex{}
)

func getConn(addr string) (*grpc.ClientConn, error) {
	conn, ok := connPool[addr]
	if ok {
		return conn, nil
	}
	poolLock.Lock()
	defer poolLock.Unlock()
	conn, ok = connPool[addr]
	if ok {
		return conn, nil
	}
	var err error
	conn, err = grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	connPool[addr] = conn
	return conn, nil
}

func Close() error {
	var errs []error
	for _, v := range connPool {
		if err := v.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

func ResolveErr(err error) response.IResponseErr {
	s, ok := status.FromError(err)
	if !ok {
		return response.NewError(500, err.Error())
	}
	switch s.Code() {
	case codes.OK:
		return nil
	case codes.InvalidArgument, codes.Aborted, codes.NotFound:
		return response.NewError(400, s.Message())
	default:
		return response.NewError(500, s.Message())
	}
}

var performCollector performance.Collector

func SetPerformanceCollector(c performance.Collector) {
	performCollector = c
}

func perform(written bool) func() {
	if performCollector == nil {
		return func() {}
	}
	t := time.Now()
	return func() {
		performCollector.PutAsync(
			util.IfElse(written, performance.ActionWrite, performance.ActionRead),
			performance.KindOfGRPC,
			time.Since(t),
		)
	}
}
