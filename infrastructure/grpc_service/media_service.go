package grpc_service

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"media-service/domain/entity"
	"os"

	"media-service/domain/usecase"
	"strings"

	"github.com/anhvanhoa/sf-proto/gen/media/v1"

	"github.com/anhvanhoa/service-core/domain/goid"
	"github.com/anhvanhoa/service-core/domain/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type MediaServiceServer struct {
	media.UnsafeMediaServiceServer
	mediaUsecases usecase.MediaUsecaseInterfaces
	logger        *log.LogGRPCImpl
	uuid          goid.GoUUID
}

func NewMediaServiceServer(mediaUsecases usecase.MediaUsecaseInterfaces, logger *log.LogGRPCImpl) media.MediaServiceServer {
	uuid := goid.NewGoId().UUID()
	return &MediaServiceServer{
		mediaUsecases: mediaUsecases,
		logger:        logger,
		uuid:          uuid,
	}
}

func (s *MediaServiceServer) UploadMediaStream(stream media.MediaService_UploadMediaStreamServer) error {
	var info *media.UploadMediaStreamRequest
	tmpFile, err := os.CreateTemp("", "stream-upload-*")
	if err != nil {
		return status.Errorf(codes.Internal, "cannot create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	for {
		chunk, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return status.Errorf(codes.Internal, "failed to receive chunk: %v", err)
		}

		if chunk.GetInfo() != nil {
			info = chunk.GetInfo()
		}
		if data := chunk.GetChunk(); data != nil {
			if _, err := tmpFile.Write(data); err != nil {
				return status.Errorf(codes.Internal, "failed to write chunk: %v", err)
			}
		}
	}

	if info == nil {
		return status.Errorf(codes.InvalidArgument, "missing upload info")
	}

	_, _ = tmpFile.Seek(0, 0)

	infoFile, err := tmpFile.Stat()
	if err != nil {
		return status.Errorf(codes.Internal, "failed to get file info: %v", err)
	}

	id := s.uuid.Gen()
	uploadReq := &usecase.UploadMediaStreamRequest{
		ID:        id,
		FileName:  info.FileName,
		CreatedBy: info.CreatedBy,
		Metadata:  info.Metadata,
		FileData:  tmpFile,
		FileSize:  infoFile.Size(),
	}

	result, err := s.mediaUsecases.UploadMediaStream(stream.Context(), uploadReq)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to upload media via stream: %v", err))
		return status.Errorf(codes.Internal, "failed to upload media: %v", err)
	}

	response := &media.UploadMediaResponse{
		Media: s.entityToProto(result),
	}

	return stream.SendAndClose(response)
}

func (s *MediaServiceServer) UploadMedia(ctx context.Context, req *media.UploadMediaRequest) (*media.UploadMediaResponse, error) {
	id := s.uuid.Gen()
	uploadReq := &usecase.UploadMediaRequest{
		ID:        id,
		FileName:  req.FileName,
		CreatedBy: req.CreatedBy,
		Metadata:  req.Metadata,
		FileData:  bytes.NewReader(req.FileData),
		Type:      entity.MediaTypeImage,
		Size:      int64(len(req.FileData)),
	}

	result, err := s.mediaUsecases.UploadMedia(ctx, uploadReq)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to upload media: %v", err))
		return nil, status.Errorf(codes.Internal, "failed to upload media: %v", err)
	}

	return &media.UploadMediaResponse{
		Media: s.entityToProto(result),
	}, nil
}

func (s *MediaServiceServer) GetMedia(ctx context.Context, req *media.GetMediaRequest) (*media.GetMediaResponse, error) {
	result, err := s.mediaUsecases.GetByID(ctx, req.Id)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to get media: %v", err))
		if strings.Contains(err.Error(), "not found") {
			return nil, status.Errorf(codes.NotFound, "media not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get media: %v", err)
	}

	return &media.GetMediaResponse{
		Media: s.entityToProto(result),
	}, nil
}

func (s *MediaServiceServer) ListMedia(ctx context.Context, req *media.ListMediaRequest) (*media.ListMediaResponse, error) {
	listReq := &usecase.ListMediaRequest{
		CreatedBy: req.CreatedBy,
		Limit:     int(req.Limit),
		Offset:    int(req.Offset),
		SortBy:    req.SortBy,
		SortOrder: req.SortOrder,
	}

	if req.Type != "" {
		listReq.Type = entity.MediaType(req.Type)
	}
	if req.MimeType != "" {
		listReq.MimeType = req.MimeType
	}

	response, err := s.mediaUsecases.List(ctx, listReq)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to list media: %v", err))
		return nil, status.Errorf(codes.Internal, "failed to list media: %v", err)
	}

	protoResults := make([]*media.Media, len(response.Media))
	for i, result := range response.Media {
		protoResults[i] = s.entityToProto(result)
	}

	return &media.ListMediaResponse{
		Media: protoResults,
		Total: int32(response.Total),
	}, nil
}

func (s *MediaServiceServer) UpdateMedia(ctx context.Context, req *media.UpdateMediaRequest) (*media.UpdateMediaResponse, error) {
	updateReq := &usecase.UpdateMediaRequest{
		Metadata: req.Metadata,
	}

	if req.Name != "" {
		updateReq.Name = &req.Name
	}

	result, err := s.mediaUsecases.Update(ctx, req.Id, req.CreatedBy, updateReq)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to update media: %v", err))
		if strings.Contains(err.Error(), "not found") {
			return nil, status.Errorf(codes.NotFound, "media not found")
		}
		if strings.Contains(err.Error(), "unauthorized") {
			return nil, status.Errorf(codes.PermissionDenied, "unauthorized")
		}
		return nil, status.Errorf(codes.Internal, "failed to update media: %v", err)
	}

	return &media.UpdateMediaResponse{
		Media: s.entityToProto(result),
	}, nil
}

func (s *MediaServiceServer) DeleteMedia(ctx context.Context, req *media.DeleteMediaRequest) (*media.DeleteMediaResponse, error) {
	err := s.mediaUsecases.Delete(ctx, req.Id, req.CreatedBy)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to delete media: %v", err))
		if strings.Contains(err.Error(), "not found") {
			return nil, status.Errorf(codes.NotFound, "media not found")
		}
		if strings.Contains(err.Error(), "unauthorized") {
			return nil, status.Errorf(codes.PermissionDenied, "unauthorized")
		}
		return nil, status.Errorf(codes.Internal, "failed to delete media: %v", err)
	}

	return &media.DeleteMediaResponse{
		Success: true,
	}, nil
}

func (s *MediaServiceServer) entityToProto(entity *entity.Media) *media.Media {
	proto := &media.Media{
		Id:               entity.ID,
		CreatedBy:        entity.CreatedBy,
		Name:             entity.Name,
		Size:             entity.Size,
		Url:              entity.URL,
		MimeType:         entity.MimeType,
		Type:             string(entity.Type),
		ProcessingStatus: string(entity.ProcessingStatus),
		Metadata:         entity.Metadata,
		CreatedAt:        timestamppb.New(entity.CreatedAt),
		UpdatedAt:        timestamppb.New(entity.UpdatedAt),
	}

	if entity.Width != nil {
		proto.Width = int32(*entity.Width)
	}
	if entity.Height != nil {
		proto.Height = int32(*entity.Height)
	}
	if entity.Duration != nil {
		proto.Duration = int32(*entity.Duration)
	}

	return proto
}
