package grpc_service

import (
	"context"
	"fmt"
	"io"
	"media-service/domain/entity"

	"media-service/domain/usecase"
	"strings"

	"media-service/proto/media/v1"

	"github.com/anhvanhoa/service-core/domain/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// MediaServiceServer implements the gRPC media service
type MediaServiceServer struct {
	media.UnimplementedMediaServiceServer
	mediaUsecases *usecase.MediaUsecases
	logger        *log.LogGRPCImpl
}

// NewMediaServiceServer creates a new media service server
func NewMediaServiceServer(mediaUsecases *usecase.MediaUsecases, logger *log.LogGRPCImpl) *MediaServiceServer {
	return &MediaServiceServer{
		mediaUsecases: mediaUsecases,
		logger:        logger,
	}
}

func (s *MediaServiceServer) UploadMedia(stream media.MediaService_UploadMediaServer) error {
	var req *media.UploadMediaRequest
	var fileData []byte

	for {
		chunk, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			s.logger.Error(fmt.Sprintf("Failed to receive chunk: %v", err))
			return status.Errorf(codes.Internal, "failed to receive chunk: %v", err)
		}

		if chunk.GetInfo() != nil {
			req = chunk.GetInfo()
		}
		if chunk.GetChunk() != nil {
			fileData = append(fileData, chunk.GetChunk()...)
		}
	}

	if req == nil {
		return status.Errorf(codes.InvalidArgument, "missing upload info")
	}

	// Create upload request
	uploadReq := &usecase.UploadMediaRequest{
		FileName:  req.FileName,
		FileData:  strings.NewReader(string(fileData)),
		FileSize:  int64(len(fileData)),
		MimeType:  req.MimeType,
		CreatedBy: req.CreatedBy,
		Metadata:  req.Metadata,
	}

	// Upload media
	result, err := s.mediaUsecases.UploadMedia(stream.Context(), uploadReq)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to upload media: %v", err))
		return status.Errorf(codes.Internal, "failed to upload media: %v", err)
	}

	// Send response
	response := &media.UploadMediaResponse{
		Media: s.entityToProto(result),
	}

	return stream.SendAndClose(response)
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

func (s *MediaServiceServer) GetMediaVariants(ctx context.Context, req *media.GetMediaVariantsRequest) (*media.GetMediaVariantsResponse, error) {
	variants, err := s.mediaUsecases.GetVariants(ctx, req.MediaId)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to get media variants: %v", err))
		if strings.Contains(err.Error(), "not found") {
			return nil, status.Errorf(codes.NotFound, "media not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get media variants: %v", err)
	}

	protoVariants := make([]*media.MediaVariant, len(variants))
	for i, variant := range variants {
		protoVariants[i] = s.variantEntityToProto(variant)
	}

	return &media.GetMediaVariantsResponse{
		Variants: protoVariants,
	}, nil
}

func (s *MediaServiceServer) ProcessMedia(ctx context.Context, req *media.ProcessMediaRequest) (*media.ProcessMediaResponse, error) {
	err := s.mediaUsecases.ProcessMedia(ctx, req.MediaId)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to process media: %v", err))
		if strings.Contains(err.Error(), "not found") {
			return nil, status.Errorf(codes.NotFound, "media not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to process media: %v", err)
	}

	return &media.ProcessMediaResponse{
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
		proto.Width = *entity.Width
	}
	if entity.Height != nil {
		proto.Height = *entity.Height
	}
	if entity.Duration != nil {
		proto.Duration = *entity.Duration
	}

	return proto
}

func (s *MediaServiceServer) variantEntityToProto(entity *entity.MediaVariant) *media.MediaVariant {
	proto := &media.MediaVariant{
		Id:        entity.ID,
		MediaId:   entity.MediaID,
		Type:      entity.Type,
		Size:      entity.Size,
		Url:       entity.URL,
		FileSize:  entity.FileSize,
		Format:    entity.Format,
		CreatedAt: timestamppb.New(entity.CreatedAt),
	}

	if entity.Width != nil {
		proto.Width = *entity.Width
	}
	if entity.Height != nil {
		proto.Height = *entity.Height
	}
	if entity.Quality != nil {
		proto.Quality = *entity.Quality
	}

	return proto
}
