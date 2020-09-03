package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"mime"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"

	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/Percona-Lab/percona-version-service/server"
	pbVersion "github.com/Percona-Lab/percona-version-service/versionpb"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/rakyll/statik/fs"
	"google.golang.org/grpc"

	_ "github.com/Percona-Lab/percona-version-service/statik"
)

func getOpenAPIHandler() http.Handler {
	err := mime.AddExtensionType(".svg", "image/svg+xml")
	if err != nil {
		log.Fatalf("creating OpenAPI filesystem: %v", err)
	}

	statikFS, err := fs.New()
	if err != nil {
		log.Fatalf("creating OpenAPI filesystem: %v", err)
	}

	return http.FileServer(statikFS)
}

func main() {
	useTLS := strings.ToLower(os.Getenv("SERVE_HTTP")) != "true"

	logger := initLogger()

	grpcport := os.Getenv("GRPC_PORT")
	if grpcport == "" {
		grpcport = "10000"
	}
	addr := "127.0.0.1:" + grpcport
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		logger.Fatal("failed to listen interface", zap.Error(err), zap.String("addr", addr))
	}

	var tlsConfig *tls.Config
	if useTLS {
		cert, err := tls.LoadX509KeyPair("certs/cert.pem", "certs/key.pem")
		if err != nil {
			logger.Fatal("failed to load key pair", zap.Error(err))
		}

		tlsConfig = &tls.Config{
			Certificates: []tls.Certificate{cert},
		}
	}

	s := grpc.NewServer(grpcServerLogOpt(logger))
	pbVersion.RegisterVersionServiceServer(s, server.New())

	logger.Info("serving gRPC", zap.String("Addr", "http://"+addr))

	go func() {
		if err := s.Serve(lis); err != nil {
			logger.Fatal("Failed to serve grpc server", zap.Error(err))
		}
	}()

	// See https://github.com/grpc/grpc/blob/master/doc/naming.md
	// for gRPC naming standard information.
	dialAddr := fmt.Sprintf("dns:///%s", addr)
	dialCreds := grpc.WithInsecure()
	conn, err := grpc.DialContext(
		context.Background(),
		dialAddr,
		dialCreds,
		grpc.WithBlock(),
	)
	if err != nil {
		logger.Fatal("failed to dial server", zap.Error(err), zap.String("dialAddr", dialAddr))
	}

	gwmux := runtime.NewServeMux()

	err = pbVersion.RegisterVersionServiceHandler(context.Background(), gwmux, conn)
	if err != nil {
		logger.Fatal("failed to register gateway", zap.Error(err))
	}
	oa := getOpenAPIHandler()

	port := os.Getenv("GW_PORT")
	if port == "" {
		port = "11000"
	}
	gatewayAddr := "0.0.0.0:" + port
	gwServer := &http.Server{
		Addr: gatewayAddr,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.URL.Path, "/versions") {
				gwmux.ServeHTTP(w, r)
				return
			}

			oa.ServeHTTP(w, r)
		}),
	}

	if !useTLS {
		logger.Info("serving gRPC-Gateway and OpenAPI Documentation", zap.String("gatewayAddr", "http://"+gatewayAddr))
		if err := gwServer.ListenAndServe(); err != nil {
			logger.Fatal("failed to serve gRPC-Gateway", zap.Error(err), zap.Bool("tls", false))
		}
	}

	gwServer.TLSConfig = tlsConfig
	logger.Info("serving gRPC-Gateway and OpenAPI Documentation", zap.String("gatewayAddr", "https://"+gatewayAddr))
	if err := gwServer.ListenAndServeTLS("", ""); err != nil {
		logger.Fatal("failed to serve gRPC-Gateway", zap.Error(err), zap.Bool("tls", true))
	}
}

func initLogger() *zap.Logger {
	logConf := zap.NewProductionEncoderConfig()
	logConf.EncodeTime = func(time time.Time, encoder zapcore.PrimitiveArrayEncoder) {
		encoder.AppendInt64(time.Unix())
	}

	levelEnablerFunc := zap.LevelEnablerFunc(func(_ zapcore.Level) bool {
		return true
	})

	logEncoder := zapcore.NewConsoleEncoder(logConf)

	if os.Getenv("LOGGER_MODE") == "PRODUCTION" {
		logEncoder = zapcore.NewJSONEncoder(logConf)
	}

	logger := zap.New(zapcore.NewCore(logEncoder, zapcore.Lock(os.Stderr), levelEnablerFunc))

	grpc_zap.ReplaceGrpcLoggerV2(logger)

	return logger
}

func grpcServerLogOpt(logger *zap.Logger) grpc.ServerOption {
	return grpc_middleware.WithUnaryServerChain(
		grpc_zap.PayloadUnaryServerInterceptor(logger, func(_ context.Context, _ string, _ interface{}) bool {
			return true
		}),
		grpc_zap.UnaryServerInterceptor(logger),
	)
}
