// Copyright 2026 Duc-Hung Ho.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package analyticsvc

import (
	"context"
	"time"

	"github.com/sentinez/controlplane/apps/collector/v1/repos/topvalues"
	pb "github.com/sentinez/sentinez/api/proto/sentinez/apps/analytic/v1"
)

// The values per field of a top, by default and at most.
const (
	_topLimit    = 5
	_maxTopLimit = 20
)

// _topRange the range of a top without from: the hour before its end.
const _topRange = time.Hour

func (srv *AnalyticService) GetEventTopValues(ctx context.Context,
	req *pb.GetEventTopValuesRequest,
) (*pb.GetEventTopValuesResponse, error) {
	filter, err := rangedFilter(req, time.Now(), _topRange)
	if err != nil {
		return nil, err
	}

	result, err := srv.events.TopValues(ctx, req.GetKind(), filter,
		req.GetFields(), topLimit(req.GetLimit()))
	if err != nil {
		return nil, filterError(err)
	}

	return topValuesResponse(result), nil
}

// topLimit limit bounded to _maxTopLimit, _topLimit when unset.
func topLimit(limit int32) uint64 {
	if limit <= 0 {
		return _topLimit
	}

	return uint64(min(limit, _maxTopLimit))
}

// topValuesResponse the proto of result.
func topValuesResponse(
	result *topvalues.Result) *pb.GetEventTopValuesResponse {
	resp := &pb.GetEventTopValuesResponse{
		Total:  result.Total,
		Fields: make([]*pb.FieldTopValues, 0, len(result.Fields)),
	}
	for _, field := range result.Fields {
		top := &pb.FieldTopValues{
			Field:  field.Field,
			Values: make([]*pb.FieldValue, 0, len(field.Values)),
		}
		for _, value := range field.Values {
			top.Values = append(top.Values, &pb.FieldValue{
				Value: value.Value,
				Total: value.Total,
				Share: shareOf(value.Total, result.Total),
			})
		}
		resp.Fields = append(resp.Fields, top)
	}

	return resp
}

// shareOf part over total, 0 without any.
func shareOf(part, total int64) float64 {
	if total <= 0 {
		return 0
	}

	return float64(part) / float64(total)
}
