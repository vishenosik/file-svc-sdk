package api

import (
	"bytes"
	"io"
	"log/slog"

	file_svc_v1 "github.com/vishenosik/file-svc-sdk/gen/grpc/v1/file_svc"
	"github.com/vishenosik/gocherry/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (fsa *FileServiceApi) UploadStream(stream file_svc_v1.FileService_UploadStreamServer) error {

	imageData := bytes.Buffer{}
	var (
		filename    string
		fileSize    uint32
		chunksCount int
	)

	for {

		req, err := stream.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return status.Errorf(codes.Internal, "cannot read chunk: %v", err)
		}
		chunk := req.GetChunk()
		size := len(chunk)
		fileSize += uint32(size)

		_, err = imageData.Write(chunk)
		if err != nil {
			return status.Errorf(codes.Internal, "cannot write chunk data: %v", err)
		}

		if filename == "" {
			filename = req.GetFileName()
		}
		chunksCount++
	}

	id, err := fsa.svc.Upload(filename, imageData.Bytes())
	if err != nil {
		return status.Errorf(codes.Internal, "cannot upload file: %v", err)
	}

	fsa.log.Info("new file uploaded",
		slog.Int("file_size", int(fileSize)),
		slog.Int("chunks_count", chunksCount),
		slog.String("id", id),
	)

	return stream.SendAndClose(&file_svc_v1.FileUploadStreamResponse{
		Id:   id,
		Size: fileSize,
	})
}
