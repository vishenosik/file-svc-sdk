package api

import (
	"context"

	file_svc_v1 "github.com/vishenosik/file-svc-sdk/gen/grpc/v1/file_svc"
)

func (fsa *FileServiceApi) Constraints(ctx context.Context, req *file_svc_v1.ConstraintsRequest) (*file_svc_v1.ConstraintsResponse, error) {
	return &file_svc_v1.ConstraintsResponse{
		MaxBatchSize: fsa.svc.GetBatchSize(),
		MaxFileSize:  fsa.svc.GetMaxFileSize(),
	}, nil
}
