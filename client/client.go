package client

import (
	"context"
	"io"
	"time"

	file_svc_v1 "github.com/vishenosik/file-svc-sdk/gen/grpc/v1/file_svc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/backoff"
	"google.golang.org/grpc/credentials/insecure"
)

type FileServiceV1 interface {
	Download(ctx context.Context, id string) (*DownloadResponse, error)
	Upload(ctx context.Context, file io.Reader, filename string) (*UploadResponse, error)
	DeleteFile(id string) error
	FileInfo(id string) (*FileInfo, error)
	ListFiles() (*FilesList, error)
}

type FileServiceClient struct {
	addr    string
	timeout time.Duration
	conn    *grpc.ClientConn
	v1      FileServiceV1
}

func NewFileServiceClient(config FileServiceConfig) (*FileServiceClient, error) {

	if err := config.validate(); err != nil {
		return nil, err
	}

	cli := &FileServiceClient{
		addr:    config.Addr,
		timeout: config.Timeout,
	}

	if err := cli.connect(); err != nil {
		return nil, err
	}

	return cli, nil
}

func (cli *FileServiceClient) V1() FileServiceV1 {
	return cli.v1
}

func (cli *FileServiceClient) Close(_ context.Context) error {
	return cli.conn.Close()
}

func (cli *FileServiceClient) connect() (err error) {

	cli.conn, err = grpc.NewClient(cli.addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithConnectParams(grpc.ConnectParams{
			Backoff: backoff.Config{
				BaseDelay:  time.Second,
				Multiplier: 1.5,
				Jitter:     0.2,
				MaxDelay:   time.Minute,
			},
			MinConnectTimeout: cli.timeout,
		}),
	)
	if err != nil {
		return err
	}

	cli.v1 = &fileServiceV1{
		client: file_svc_v1.NewFileServiceClient(cli.conn),
	}

	return nil
}
