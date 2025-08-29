package media

// Upload requests/responses
type UploadMediaRequest struct {
	FileName  string            `json:"file_name"`
	MimeType  string            `json:"mime_type"`
	CreatedBy string            `json:"created_by"`
	Metadata  map[string]string `json:"metadata"`
}

type UploadMediaChunk struct {
	Info  *UploadMediaRequest `json:"info,omitempty"`
	Chunk []byte              `json:"chunk,omitempty"`
}

func (x *UploadMediaChunk) GetInfo() *UploadMediaRequest {
	return x.Info
}

func (x *UploadMediaChunk) GetChunk() []byte {
	return x.Chunk
}

type UploadMediaResponse struct {
	Media *Media `json:"media"`
}

// Get requests/responses
type GetMediaRequest struct {
	Id string `json:"id"`
}

type GetMediaResponse struct {
	Media *Media `json:"media"`
}

// List requests/responses
type ListMediaRequest struct {
	CreatedBy string `json:"created_by"`
	Type      string `json:"type"`
	MimeType  string `json:"mime_type"`
	Limit     int32  `json:"limit"`
	Offset    int32  `json:"offset"`
	SortBy    string `json:"sort_by"`
	SortOrder string `json:"sort_order"`
}

type ListMediaResponse struct {
	Media []*Media `json:"media"`
	Total int32    `json:"total"`
}

// Update requests/responses
type UpdateMediaRequest struct {
	Id        string            `json:"id"`
	CreatedBy string            `json:"created_by"`
	Name      string            `json:"name"`
	Metadata  map[string]string `json:"metadata"`
}

type UpdateMediaResponse struct {
	Media *Media `json:"media"`
}

// Delete requests/responses
type DeleteMediaRequest struct {
	Id        string `json:"id"`
	CreatedBy string `json:"created_by"`
}

type DeleteMediaResponse struct {
	Success bool `json:"success"`
}

// Variants requests/responses
type GetMediaVariantsRequest struct {
	MediaId string `json:"media_id"`
}

type GetMediaVariantsResponse struct {
	Variants []*MediaVariant `json:"variants"`
}

// Process requests/responses
type ProcessMediaRequest struct {
	MediaId string `json:"media_id"`
}

type ProcessMediaResponse struct {
	Success bool `json:"success"`
}
