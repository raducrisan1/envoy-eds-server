package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"

	discoverygrpc "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	endpointservice "github.com/envoyproxy/go-control-plane/envoy/service/endpoint/v3"
	"github.com/envoyproxy/go-control-plane/pkg/server/v3"
)

const (
	grpcKeepaliveTime        = 30 * time.Second
	grpcKeepaliveTimeout     = 5 * time.Second
	grpcKeepaliveMinTime     = 30 * time.Second
	grpcMaxConcurrentStreams = 1000000
)

type CustomEdsServer struct {
	GrpcServer *grpc.Server
}

func (s *CustomEdsServer) Shutdown() {
	if s.GrpcServer != nil {
		s.GrpcServer.Stop()
	}
}

func (s *CustomEdsServer) registerServer(server server.Server) {
	// register services
	discoverygrpc.RegisterAggregatedDiscoveryServiceServer(s.GrpcServer, server)
	endpointservice.RegisterEndpointDiscoveryServiceServer(s.GrpcServer, server)
}

func (s *CustomEdsServer) Initialize() {
	// gRPC golang library sets a very small upper bound for the number gRPC/h2
	// streams over a single TCP connection. If a proxy multiplexes requests over
	// a single connection to the management server, then it might lead to
	// availability problems. Keepalive timeouts based on connection_keepalive parameter https://www.envoyproxy.io/docs/envoy/latest/configuration/overview/examples#dynamic
	var grpcOptions []grpc.ServerOption
	grpcOptions = append(grpcOptions,
		grpc.MaxConcurrentStreams(grpcMaxConcurrentStreams),
		grpc.KeepaliveParams(keepalive.ServerParameters{
			Time:    grpcKeepaliveTime,
			Timeout: grpcKeepaliveTimeout,
		}),
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime:             grpcKeepaliveMinTime,
			PermitWithoutStream: true,
		}),
	)
	s.GrpcServer = grpc.NewServer(grpcOptions...)
}

// RunGrpcServer starts an xDS server at the given port.
func (s *CustomEdsServer) RunGrpcServer(ctx context.Context, srv server.Server, port uint) {

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatal(err)
	}

	s.registerServer(srv)

	log.Printf("EDS Server is listening for incoming GRPC requests from Envoy on port %d", port)
	if err = s.GrpcServer.Serve(lis); err != nil {
		log.Println(err)
	}
}
