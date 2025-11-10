package api

import (
	"context"

	file_svc_v1 "github.com/vishenosik/file-svc-sdk/gen/grpc/v1/file_svc"
)

func (fsa *FileServiceApi) Constraints(ctx context.Context, req *file_svc_v1.ConstraintsReq) (*file_svc_v1.ConstraintsResp, error) {
	return &file_svc_v1.ConstraintsResp{
		MaxBatchSize: fsa.svc.GetBatchSize(),
		MaxFileSize:  fsa.svc.GetMaxFileSize(),
	}, nil
}
