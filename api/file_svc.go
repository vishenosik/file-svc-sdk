package api

import (
	"bytes"
	"io"
	"log/slog"

	file_svc_v1 "github.com/vishenosik/file-svc-sdk/gen/grpc/v1/file_svc"
	"github.com/vishenosik/gocherry/pkg/errors"
	"github.com/vishenosik/gocherry/pkg/logs"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func (fsa *FileServiceApi) UploadStream(stream file_svc_v1.FileService_UploadStreamServer) error {

	log := fsa.log.With(logs.Operation("UploadStream"))

	imageData := bytes.Buffer{}

	var (
		fileSize    uint32
		chunksCount int
	)

	md, ok := metadata.FromIncomingContext(stream.Context())
	if !ok {
		return status.Errorf(codes.Internal, "cannot get metadata from context")
	}

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
		chunksCount++
	}

	filename := md.Get(FilenameHeader)[0]

	id, err := fsa.svc.Upload(filename, imageData.Bytes())
	if err != nil {
		return status.Errorf(codes.Internal, "cannot upload file: %v", err)
	}

	log.Info("file uploaded",
		slog.Int("file_size", int(fileSize)),
		slog.Int("chunks_count", chunksCount),
		slog.String("filename", filename),
		slog.String("id", id),
	)

	return stream.SendAndClose(&file_svc_v1.UploadStreamResp{
		Id:   id,
		Size: fileSize,
	})
}

func (fsa *FileServiceApi) DownloadStream(
	req *file_svc_v1.DownloadStreamReq,
	stream file_svc_v1.FileService_DownloadStreamServer,
) error {

	var (
		chunksCount int
	)

	id := req.GetId()

	file, err := fsa.svc.Download(id)
	if err != nil {
		return status.Errorf(codes.Internal, "cannot download file: %v", err)
	}

	fileReader := bytes.NewBuffer(file)

	buf := make([]byte, fsa.svc.GetBatchSize())

	for {
		num, err := fileReader.Read(buf)
		if err == io.EOF {
			break
		}

		if err != nil {
			return err
		}

		chunk := buf[:num]

		if err := stream.Send(&file_svc_v1.DownloadStreamMsg{
			Chunk: chunk,
		}); err != nil {
			return err
		}
		chunksCount++
	}

	fsa.log.Info("file downloaded",
		// slog.Int("file_size", int(fileSize)),
		slog.Int("chunks_count", chunksCount),
		slog.String("id", id),
	)
	return nil
}
