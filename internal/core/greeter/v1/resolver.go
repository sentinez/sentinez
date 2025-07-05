package greeter

import (
	"context"
	"net"
	"strconv"
	"time"

	"github.com/sentinez/sentinez/api/gen/go/sentinez/common/v1"
	discoverypb "github.com/sentinez/sentinez/api/gen/go/sentinez/core/discovery/v1"
	"github.com/sentinez/sentinez/pkg/common/cron"
	"github.com/sentinez/sentinez/pkg/std/grpc/client"
	"github.com/sentinez/sentinez/pkg/std/names"
	"github.com/sentinez/sentinez/pkg/std/zlog"
	"google.golang.org/protobuf/types/known/durationpb"
)

func Resolver(ctx context.Context, flag *common.FlagGRPCService) {
	addr, port, err := net.SplitHostPort(flag.GetAddress())
	if err != nil {
		zlog.Errorf("failed to split address: %v", err)
		return
	}

	discovery, err := client.NewDiscoveryClient()
	if err != nil {
		zlog.Errorf("failed to create discovery client: %v", err)
	}

	portInt, _ := strconv.Atoi(port)
	for {
		if resp, err := discovery.Register(ctx, &discoverypb.RegisterRequest{
			Name: names.GreeterV1.String(), Address: addr, Port: int32(portInt),
			Ttl: durationpb.New(time.Second * 15),
		}); err != nil {
			zlog.Errorf("failed to register greeter service: %v", err)
			zlog.Info("retrying to register greeter service in 10 seconds...")

		} else {
			cron.Start(context.Background(), 10*time.Second, func() {
				_, _ = discovery.Heartbeat(ctx,
					&discoverypb.HeartbeatRequest{Id: resp.GetId()})
			})
			break
		}
		time.Sleep(10 * time.Second)
	}
	zlog.Infof("greeter service registered successfully")
}
