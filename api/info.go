package api

import (
	"context"

	file_svc_v1 "github.com/vishenosik/file-svc-sdk/gen/grpc/v1/file_svc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type FileInfo struct {
	ID       string
	Size     uint32
	Filename string
}

type FileInfoList struct {
	Total uint32
	Files []*FileInfo
}

func (fsa *FileServiceApi) GetFileInfo(ctx context.Context, req *file_svc_v1.FileReq) (*file_svc_v1.FileInfoResp, error) {
	info, err := fsa.info.GetFileInfo(req.GetId())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot get file info: %v", err)
	}

	return convertToFileInfo(info), nil
}

func (fsa *FileServiceApi) ListFiles(ctx context.Context, req *file_svc_v1.ListFilesReq) (*file_svc_v1.ListFilesResp, error) {
	list, err := fsa.info.ListFiles()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot list files: %v", err)
	}
	return convertToFileInfoList(list), nil
}

func convertToFileInfo(info *FileInfo) *file_svc_v1.FileInfoResp {
	return &file_svc_v1.FileInfoResp{
		Id:       info.ID,
		Size:     info.Size,
		Filename: info.Filename,
	}
}

func convertToFileInfoList(list *FileInfoList) *file_svc_v1.ListFilesResp {

	files := make([]*file_svc_v1.FileInfoResp, 0, len(list.Files))
	for _, info := range list.Files {
		files = append(files, convertToFileInfo(info))
	}

	return &file_svc_v1.ListFilesResp{
		Total: list.Total,
		Files: files,
	}
}
