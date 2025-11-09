package client

import (
	"context"
	"net/url"
	"time"

	file_svc_v1 "github.com/vishenosik/file-svc-sdk/gen/grpc/v1/file_svc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/backoff"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	defaultTimeout = time.Second * 15
)

type FileServiceConfig struct {
	Addr    string
	Timeout time.Duration
}

func (config FileServiceConfig) validate() error {

	_, err := url.Parse(config.Addr)
	if err != nil {
		return ErrInvalidAddr
	}

	if config.Timeout <= 0 {
		config.Timeout = defaultTimeout
	}

	return nil
}

type FileServiceClient struct {
	addr        string
	timeout     time.Duration
	conn        *grpc.ClientConn
	client      file_svc_v1.FileServiceClient
	batchSize   uint32
	maxFileSize uint32
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

	if err := cli.constraints(); err != nil {
		return nil, err
	}

	return cli, nil
}

func (cli *FileServiceClient) Close(_ context.Context) error {
	return cli.conn.Close()
}

func (cli *FileServiceClient) connect() error {

	conn, err := grpc.NewClient(cli.addr,
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

	cli.conn = conn
	cli.client = file_svc_v1.NewFileServiceClient(conn)

	return nil
}

func (cli *FileServiceClient) constraints() error {
	resp, err := cli.client.Constraints(context.TODO(), &file_svc_v1.ConstraintsRequest{})
	if err != nil {
		return err
	}
	cli.batchSize = resp.GetMaxBatchSize()
	cli.maxFileSize = resp.GetMaxFileSize()
	return nil
}
