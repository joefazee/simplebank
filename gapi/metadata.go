package gapi

import (
	"context"
	"fmt"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
)

const (
	grpcGatewayUserAgentHeader = "grpcgateway-user-agent"
	userAgent                  = "user-agent"
	xForwardedForHeader        = "x-forwarded-for"
)

type Metadata struct {
	ClientIP  string
	UserAgent string
}

func (server *Server) extractMetadata(ctx context.Context) *Metadata {

	mtdt := &Metadata{}
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		fmt.Printf("md: %v\n\n", md)

		if userAgents := md.Get(grpcGatewayUserAgentHeader); len(userAgents) > 0 {
			mtdt.UserAgent = userAgents[0]
		}

		if userAgents := md.Get(userAgent); len(userAgents) > 0 {
			mtdt.UserAgent = userAgents[0]
		}

		if clientIps := md.Get(xForwardedForHeader); len(clientIps) > 0 {
			mtdt.ClientIP = clientIps[0]
		}

	}

	if p, ok := peer.FromContext(ctx); ok {
		mtdt.ClientIP = p.Addr.String()
	}

	return mtdt
}
