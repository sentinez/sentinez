package greeter

import (
	"context"
	"net"
	"strconv"
	"time"

	discoverypb "github.com/sentinez/sentinez/api/gen/go/sentinez/common/discovery/v1"
	"github.com/sentinez/sentinez/api/gen/go/sentinez/common/v1"
	"github.com/sentinez/sentinez/client/discovery"
	"github.com/sentinez/sentinez/pkg/common/cron"
	"github.com/sentinez/sentinez/pkg/std/flags"
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

	dcvr := discovery.GetDiscovery(flags.Parse().GetConsulUrl())
	portInt, _ := strconv.Atoi(port)
	for {
		if resp, err := dcvr.Register(ctx, &discoverypb.RegisterRequest{
			Name: names.GreeterV1.String(), Address: addr, Port: int32(portInt),
			Ttl: durationpb.New(time.Second * 15),
		}); err != nil {
			zlog.Errorf("failed to register greeter service: %v", err)
			zlog.Info("retrying to register greeter service in 10 seconds...")

		} else {
			cron.Start(context.Background(), 10*time.Second, func() {
				_, _ = dcvr.Heartbeat(ctx,
					&discoverypb.HeartbeatRequest{Id: resp.GetId()})
			})
			break
		}
		time.Sleep(10 * time.Second)
	}
	zlog.Infof("greeter service registered successfully")
}
