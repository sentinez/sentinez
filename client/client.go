package client

import (
	"context"

	discoverypb "github.com/sentinez/sentinez/api/gen/go/sentinez/common/discovery/v1"
	"github.com/sentinez/sentinez/api/gen/go/sentinez/core/iam/v1"
	"github.com/sentinez/sentinez/client/discovery"
	"github.com/sentinez/sentinez/pkg/std/names"
)

func NewIAMService(ctx context.Context,
	consulURL string) (iam.IdentityAccessManagementServiceClient, error) {

	srv, err := discovery.GetDiscovery(consulURL).
		Discover(ctx, &discoverypb.DiscoverRequest{Name: names.IAMV1.String()})
	if err != nil {
		return nil, err
	}

	conn, err := newConnection(srv.GetAddress())
	if err != nil {
		return nil, err
	}

	return iam.NewIdentityAccessManagementServiceClient(conn), nil
}
