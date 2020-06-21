package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"log"
	"mime"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/Percona-Lab/percona-version-service/server"
	pbVersion "github.com/Percona-Lab/percona-version-service/versionpb"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/rakyll/statik/fs"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/grpclog"

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

	log := grpclog.NewLoggerV2(os.Stdout, ioutil.Discard, ioutil.Discard)
	grpclog.SetLoggerV2(log)
	grpcport := os.Getenv("GRPC_PORT")
	if grpcport == "" {
		grpcport = "10000"
	}
	addr := "0.0.0.0:" + grpcport
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	servOpts := []grpc.ServerOption{}
	var tlsConfig *tls.Config
	schema := "http"
	if useTLS {
		cert, err := tls.LoadX509KeyPair("certs/cert.pem", "certs/key.pem")
		if err != nil {
			log.Fatalf("failed to load key pair: %v", err)
		}
		servOpts = append(servOpts, grpc.Creds(credentials.NewServerTLSFromCert(&cert)))

		tlsConfig = &tls.Config{
			Certificates: []tls.Certificate{cert},
		}
		schema = "https"
	}
	s := grpc.NewServer(
		servOpts...,
	)

	pbVersion.RegisterVersionServiceServer(s, server.New())
	log.Infof("serving gRPC on %s://%s", schema, addr)
	go func() {
		log.Fatal(s.Serve(lis))
	}()

	certPool, err := x509.SystemCertPool()
	if err != nil {
		log.Fatalf("failed to load system cert pool: %v", err)
	}
	// See https://github.com/grpc/grpc/blob/master/doc/naming.md
	// for gRPC naming standard information.
	dialAddr := fmt.Sprintf("dns:///%s", addr)
	dialCreds := grpc.WithInsecure()
	if useTLS {
		dialCreds = grpc.WithTransportCredentials(credentials.NewClientTLSFromCert(certPool, ""))
	}
	conn, err := grpc.DialContext(
		context.Background(),
		dialAddr,
		dialCreds,
		grpc.WithBlock(),
	)
	if err != nil {
		log.Fatalln("failed to dial server:", err)
	}

	gwmux := runtime.NewServeMux()

	err = pbVersion.RegisterVersionServiceHandler(context.Background(), gwmux, conn)
	if err != nil {
		log.Fatalln("failed to register gateway:", err)
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
			if strings.HasPrefix(r.URL.Path, "/api") {
				gwmux.ServeHTTP(w, r)
				return
			}
			oa.ServeHTTP(w, r)
		}),
	}

	if !useTLS {
		log.Info("serving gRPC-Gateway and OpenAPI Documentation on http://", gatewayAddr)
		log.Fatalln(gwServer.ListenAndServe())
	}

	gwServer.TLSConfig = tlsConfig
	log.Info("serving gRPC-Gateway and OpenAPI Documentation on https://", gatewayAddr)
	log.Fatalln(gwServer.ListenAndServeTLS("", ""))
}
