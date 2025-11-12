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

type fileServiceV1 struct {
	client      file_svc_v1.FileServiceClient
	batchSize   uint32
	maxFileSize uint32
}

type UploadResponse struct {
	ID   string
	Size uint32
}

func (cli *fileServiceV1) Upload(
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

func (cli *fileServiceV1) Download(
	ctx context.Context,
	id string,
) (*DownloadResponse, error) {

	stream, err := cli.client.DownloadStream(ctx, &file_svc_v1.FileReq{
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

func (cli *fileServiceV1) constraints() error {
	resp, err := cli.client.Constraints(context.TODO(), &file_svc_v1.ConstraintsReq{})
	if err != nil {
		return err
	}
	cli.batchSize = resp.GetMaxBatchSize()
	cli.maxFileSize = resp.GetMaxFileSize()
	return nil
}

type FileInfo struct {
	ID   string `json:"id"`
	Size uint32 `json:"size"`
	Name string `json:"name"`
}

type FilesList struct {
	Total uint32     `json:"total"`
	Files []FileInfo `json:"files"`
}

func (cli *fileServiceV1) ListFiles() (*FilesList, error) {
	resp, err := cli.client.ListFiles(context.TODO(), &file_svc_v1.ListFilesReq{})
	if err != nil {
		return nil, err
	}

	files := make([]FileInfo, 0, len(resp.GetFiles()))
	for _, f := range resp.GetFiles() {
		files = append(files, FileInfo{
			ID:   f.GetId(),
			Size: f.GetSize(),
			Name: f.GetFilename(),
		})
	}

	return &FilesList{
		Total: resp.GetTotal(),
		Files: files,
	}, nil
}

func (cli *fileServiceV1) FileInfo(id string) (*FileInfo, error) {
	resp, err := cli.client.GetFileInfo(context.TODO(), &file_svc_v1.FileReq{
		Id: id,
	})
	if err != nil {
		return nil, err
	}

	return &FileInfo{
		ID:   resp.GetId(),
		Size: resp.GetSize(),
		Name: resp.GetFilename(),
	}, nil
}

func (cli *fileServiceV1) DeleteFile(id string) error {
	_, err := cli.client.DeleteFile(context.TODO(), &file_svc_v1.FileReq{
		Id: id,
	})
	if err != nil {
		return err
	}

	return nil
}
