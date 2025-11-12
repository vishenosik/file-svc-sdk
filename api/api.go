package api

import (
	"log/slog"

	file_svc_v1 "github.com/vishenosik/file-svc-sdk/gen/grpc/v1/file_svc"
	"github.com/vishenosik/gocherry/pkg/logs"
	"google.golang.org/grpc"
)

const (
	FilenameHeader = "filename"
)

type FileService interface {
	Upload(filename string, file []byte) (id string, err error)
	Download(id string) (file []byte, err error)
	DeleteFile(id string) error
}

type Info interface {
	GetFileInfo(id string) (info *FileInfo, err error)
	ListFiles() (list *FileInfoList, err error)
}

type Settings interface {
	GetBatchSize() uint32
	GetMaxFileSize() uint32
}

type FileServiceApi struct {
	file_svc_v1.UnimplementedFileServiceServer
	svc      FileService
	settings Settings
	info     Info
	// log is a structured logger for the application.
	log *slog.Logger
}

func NewFileServiceApi(svc FileService, info Info, s Settings) *FileServiceApi {
	return &FileServiceApi{
		svc:      svc,
		settings: s,
		info:     info,
		log:      logs.SetupLogger().With(appComponent()),
	}
}

func (fsa *FileServiceApi) RegisterService(server *grpc.Server) {
	file_svc_v1.RegisterFileServiceServer(server, fsa)
}

func appComponent() slog.Attr {
	return logs.AppComponent("gRPC")
}
