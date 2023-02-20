package util

import (
	"common/graceful"
	"common/logs"
	"context"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"net/http"
	"strings"
)

type H2Server struct {
	httpHandler http.Handler
	grpcHandler http.Handler
}

func H2CHandler(httpHandler http.Handler, grpcHandler http.Handler) http.Handler {
	return h2c.NewHandler(&H2Server{httpHandler: httpHandler, grpcHandler: grpcHandler}, &http2.Server{})
}

func (h *H2Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.ProtoMajor == 2 &&
		strings.HasPrefix(r.Header.Get("Content-Type"), "application/grpc") {
		h.grpcHandler.ServeHTTP(w, r)
		return
	}
	h.httpHandler.ServeHTTP(w, r)
}

func CommonUnaryInterceptors() grpc.ServerOption {
	return grpc.ChainUnaryInterceptor(UnaryServerRecoveryInterceptor(), UnaryLoggerInterceptor())
}

func CommonStreamInterceptors() grpc.ServerOption {
	return grpc.ChainStreamInterceptor(StreamServerRecoveryInterceptor(), StreamLoggerInterceptor())
}

func UnaryLoggerInterceptor() grpc.UnaryServerInterceptor {
	logger := logs.New("GRPC-Unary")
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (r interface{}, err error) {
		r, err = handler(ctx, req)
		if err != nil {
			logger.Infof("handler method %s err: %s", info.FullMethod, err)
		} else {
			logger.Infof("handler method %s OK", info.FullMethod)
		}
		return
	}
}

func StreamLoggerInterceptor() grpc.StreamServerInterceptor {
	logger := logs.New("GRPC-Stream")
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
		err = handler(srv, stream)
		if err != nil {
			logger.Infof("handler method %s err: %s", info.FullMethod, err)
		} else {
			logger.Infof("handler method %s OK", info.FullMethod)
		}
		return
	}
}

// UnaryServerRecoveryInterceptor returns a new unary server interceptor for panic recovery.
func UnaryServerRecoveryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (_ interface{}, err error) {
		defer graceful.Recover(func(msg string) {
			err = status.Error(codes.Internal, "panic")
		})
		return handler(ctx, req)
	}
}

// StreamServerRecoveryInterceptor returns a new streaming server interceptor for panic recovery.
func StreamServerRecoveryInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
		defer graceful.Recover(func(msg string) {
			err = status.Error(codes.Internal, "panic")
		})
		return handler(srv, stream)
	}
}
