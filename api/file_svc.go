package api

import (
	"bytes"
	"io"
	"log"

	file_svc_v1 "github.com/vishenosik/file-svc-sdk/gen/grpc/v1/file_svc"
	"github.com/vishenosik/gocherry/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type FileService interface {
	Upload(filename string, file []byte) (id string, err error)
}

type FileServiceApi struct {
	file_svc_v1.UnimplementedFileServiceServer
	svc FileService
}

func NewFileServiceApi(svc FileService) *FileServiceApi {
	return &FileServiceApi{
		svc: svc,
	}
}

func (fsa *FileServiceApi) RegisterService(server *grpc.Server) {
	file_svc_v1.RegisterFileServiceServer(server, fsa)
}

func (fsa *FileServiceApi) UploadStream(stream file_svc_v1.FileService_UploadStreamServer) error {

	imageData := bytes.Buffer{}
	var (
		filename string
		fileSize uint32
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

		log.Printf("received a chunk with size: %d", size)

		_, err = imageData.Write(chunk)
		if err != nil {
			return status.Errorf(codes.Internal, "cannot write chunk data: %v", err)
		}

		if filename == "" {
			filename = req.GetFileName()
		}
	}

	id, err := fsa.svc.Upload(filename, imageData.Bytes())
	if err != nil {
		return status.Errorf(codes.Internal, "cannot upload file: %v", err)
	}

	return stream.SendAndClose(&file_svc_v1.FileUploadStreamResponse{
		Id:   id,
		Size: fileSize,
	})
}
