package usecase

import (
	"context"
	"fmt"
	"media-service/domain/entity"
	"media-service/domain/repository"

	"github.com/anhvanhoa/service-core/domain/log"
)

// ListMediaUsecase handles listing media with filters and pagination
type ListMediaUsecase struct {
	mediaRepo repository.MediaRepository
	logger    *log.LogGRPCImpl
}

// ListMediaRequest represents the request for listing media
type ListMediaRequest struct {
	CreatedBy string
	Type      entity.MediaType
	MimeType  string
	Limit     int
	Offset    int
	SortBy    string // created_at, name, size
	SortOrder string // asc, desc
}

// ListMediaResponse represents the response for listing media
type ListMediaResponse struct {
	Media  []*entity.Media
	Total  int
	Limit  int
	Offset int
}

// NewListMediaUsecase creates a new list media usecase
func NewListMediaUsecase(
	mediaRepo repository.MediaRepository,
	logger *log.LogGRPCImpl,
) *ListMediaUsecase {
	return &ListMediaUsecase{
		mediaRepo: mediaRepo,
		logger:    logger,
	}
}

// Execute lists media with filters and pagination
func (uc *ListMediaUsecase) Execute(ctx context.Context, req *ListMediaRequest) (*ListMediaResponse, error) {
	uc.logger.Info(fmt.Sprintf("Listing media: %v", req))

	// Step 1: Validate and normalize input
	if err := uc.validateAndNormalizeInput(req); err != nil {
		uc.logger.Error(fmt.Sprintf("Input validation failed: %v", err))
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Step 2: Build repository filters
	filters := uc.buildRepositoryFilters(req)

	// Step 3: Retrieve media from database
	media, total, err := uc.retrieveFromDatabase(ctx, filters)
	if err != nil {
		uc.logger.Error(fmt.Sprintf("Failed to retrieve media from database: %v", err))
		return nil, fmt.Errorf("database retrieval failed: %w", err)
	}

	// Step 4: Build response
	response := uc.buildResponse(media, total, req)

	// Step 5: Log successful retrieval
	uc.logger.Info(fmt.Sprintf("Media list retrieved successfully: %v", response))

	return response, nil
}

// Step 1: Validate and normalize input
func (uc *ListMediaUsecase) validateAndNormalizeInput(req *ListMediaRequest) error {
	// Set default values
	if req.Limit <= 0 {
		req.Limit = 20 // Default limit
	}
	if req.Limit > 100 {
		req.Limit = 100 // Max limit
	}

	if req.Offset < 0 {
		req.Offset = 0
	}

	// Validate sort fields
	if req.SortBy == "" {
		req.SortBy = "created_at"
	}
	validSortFields := map[string]bool{
		"created_at": true,
		"name":       true,
		"size":       true,
		"updated_at": true,
	}
	if !validSortFields[req.SortBy] {
		return fmt.Errorf("invalid sort field: %s", req.SortBy)
	}

	// Validate sort order
	if req.SortOrder == "" {
		req.SortOrder = "desc"
	}
	if req.SortOrder != "asc" && req.SortOrder != "desc" {
		return fmt.Errorf("invalid sort order: %s (must be 'asc' or 'desc')", req.SortOrder)
	}

	return nil
}

// Step 2: Build repository filters
func (uc *ListMediaUsecase) buildRepositoryFilters(req *ListMediaRequest) repository.MediaFilters {
	return repository.MediaFilters{
		CreatedBy: req.CreatedBy,
		Type:      req.Type,
		MimeType:  req.MimeType,
		Limit:     req.Limit,
		Offset:    req.Offset,
		SortBy:    req.SortBy,
		SortOrder: req.SortOrder,
	}
}

// Step 3: Retrieve media from database
func (uc *ListMediaUsecase) retrieveFromDatabase(ctx context.Context, filters repository.MediaFilters) ([]*entity.Media, int, error) {
	return uc.mediaRepo.List(ctx, filters)
}

// Step 4: Build response
func (uc *ListMediaUsecase) buildResponse(media []*entity.Media, total int, req *ListMediaRequest) *ListMediaResponse {
	return &ListMediaResponse{
		Media:  media,
		Total:  total,
		Limit:  req.Limit,
		Offset: req.Offset,
	}
}
