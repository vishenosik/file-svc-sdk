package client

import (
	"context"
	"io"

	file_svc_v1 "github.com/vishenosik/file-svc-sdk/gen/grpc/v1/file_svc"
)

type UploadResponse struct {
	ID   string
	Size uint32
}

func (cli *FileServiceClient) Upload(
	ctx context.Context,
	file io.Reader,
	filename string,
) (*UploadResponse, error) {

	stream, err := cli.client.UploadStream(ctx)
	if err != nil {
		return nil, err
	}

	buf := make([]byte, cli.batchSize)
	batchNumber := 1
	for {
		num, err := file.Read(buf)
		if err == io.EOF {
			break
		}

		if err != nil {
			return nil, err
		}

		chunk := buf[:num]

		if err := stream.Send(&file_svc_v1.FileUploadStreamRequest{
			FileName: filename,
			Chunk:    chunk,
		}); err != nil {
			return nil, err
		}
		batchNumber += 1
	}

	res, err := stream.CloseAndRecv()
	if err != nil {
		return nil, err
	}

	return &UploadResponse{
		ID:   res.GetId(),
		Size: res.GetSize(),
	}, nil
}
