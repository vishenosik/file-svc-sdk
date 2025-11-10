package client

import (
	"bytes"
	"context"
	"io"

	"github.com/vishenosik/file-svc-sdk/api"
	file_svc_v1 "github.com/vishenosik/file-svc-sdk/gen/grpc/v1/file_svc"
	"github.com/vishenosik/gocherry/pkg/errors"
	"google.golang.org/grpc/metadata"
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

	if filename == "" {
		return nil, errors.New("filename is required")
	}

	if err := cli.constraints(); err != nil {
		return nil, err
	}

	md := metadata.New(map[string]string{
		api.FilenameHeader: filename,
	})
	ctx = metadata.NewOutgoingContext(ctx, md)

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

		if err := stream.Send(&file_svc_v1.UploadStreamMsg{
			Chunk: chunk,
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

type DownloadResponse struct {
	ID   string
	Size uint32
	File []byte
}

func (cli *FileServiceClient) Download(
	ctx context.Context,
	id string,
) (*DownloadResponse, error) {

	stream, err := cli.client.DownloadStream(ctx, &file_svc_v1.DownloadStreamReq{
		Id: id,
	})
	if err != nil {
		return nil, err
	}

	imageData := bytes.Buffer{}
	var (
		fileSize    uint32
		chunksCount int
	)

	for {

		req, err := stream.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, errors.Wrap(err, "failed to receive message")
		}

		chunk := req.GetChunk()
		fileSize += uint32(len(chunk))

		if _, err = imageData.Write(chunk); err != nil {
			return nil, errors.Wrap(err, "failed to write chunk data")
		}

		chunksCount++
	}

	return &DownloadResponse{
		ID:   id,
		Size: fileSize,
		File: imageData.Bytes(),
	}, nil
}
