package response

import (
	"encoding/json"
	"net/http"

	"github.com/SidahmedSeg/document-manager/backend/pkg/errors"
)

// Response represents a standardized API response
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *ErrorData  `json:"error,omitempty"`
	Meta    *Meta       `json:"meta,omitempty"`
}

// ErrorData represents error information in the response
type ErrorData struct {
	Code    errors.ErrorCode   `json:"code"`
	Message string             `json:"message"`
	Fields  map[string]string  `json:"fields,omitempty"`
	Meta    map[string]interface{} `json:"meta,omitempty"`
}

// Meta represents pagination and other metadata
type Meta struct {
	Page       int   `json:"page,omitempty"`
	Limit      int   `json:"limit,omitempty"`
	Total      int64 `json:"total,omitempty"`
	TotalPages int   `json:"total_pages,omitempty"`
}

// JSON writes a JSON response
func JSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := Response{
		Success: statusCode >= 200 && statusCode < 300,
		Data:    data,
	}

	_ = json.NewEncoder(w).Encode(response)
}

// Success writes a successful response
func Success(w http.ResponseWriter, data interface{}) {
	JSON(w, http.StatusOK, data)
}

// Created writes a 201 Created response
func Created(w http.ResponseWriter, data interface{}) {
	JSON(w, http.StatusCreated, data)
}

// NoContent writes a 204 No Content response
func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

// Error writes an error response
func Error(w http.ResponseWriter, err error) {
	appErr := errors.FromError(err)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(appErr.StatusCode)

	response := Response{
		Success: false,
		Error: &ErrorData{
			Code:    appErr.Code,
			Message: appErr.Message,
			Fields:  appErr.Fields,
			Meta:    appErr.Meta,
		},
	}

	_ = json.NewEncoder(w).Encode(response)
}

// WithMeta writes a response with pagination metadata
func WithMeta(w http.ResponseWriter, data interface{}, meta *Meta) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := Response{
		Success: true,
		Data:    data,
		Meta:    meta,
	}

	_ = json.NewEncoder(w).Encode(response)
}

// Paginated writes a paginated response
func Paginated(w http.ResponseWriter, data interface{}, page, limit int, total int64) {
	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}

	meta := &Meta{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
	}

	WithMeta(w, data, meta)
}

// BadRequest writes a 400 Bad Request response
func BadRequest(w http.ResponseWriter, message string) {
	Error(w, errors.New(errors.ErrCodeBadRequest, message))
}

// NotFound writes a 404 Not Found response
func NotFound(w http.ResponseWriter, message string) {
	if message == "" {
		message = "Resource not found"
	}
	Error(w, errors.NotFoundf(message))
}

// Unauthorized writes a 401 Unauthorized response
func Unauthorized(w http.ResponseWriter, message string) {
	if message == "" {
		message = "Authentication required"
	}
	Error(w, errors.Unauthorizedf(message))
}

// Forbidden writes a 403 Forbidden response
func Forbidden(w http.ResponseWriter, message string) {
	if message == "" {
		message = "Access denied"
	}
	Error(w, errors.Forbiddenf(message))
}

// Conflict writes a 409 Conflict response
func Conflict(w http.ResponseWriter, message string) {
	Error(w, errors.Conflictf(message))
}

// InternalServerError writes a 500 Internal Server Error response
func InternalServerError(w http.ResponseWriter, message string) {
	if message == "" {
		message = "Internal server error"
	}
	Error(w, errors.Internalf(nil, message))
}

// ValidationError writes a validation error response
func ValidationError(w http.ResponseWriter, err error) {
	if appErr, ok := err.(*errors.AppError); ok {
		Error(w, appErr)
	} else {
		Error(w, errors.Validationf(err.Error()))
	}
}

// CalculatePagination calculates pagination values
func CalculatePagination(page, limit int, total int64) *Meta {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}

	return &Meta{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
	}
}

// GetOffset calculates the database offset from page and limit
func GetOffset(page, limit int) int {
	if page < 1 {
		page = 1
	}
	return (page - 1) * limit
}

// PaginationParams represents pagination request parameters
type PaginationParams struct {
	Page  int `json:"page" form:"page" validate:"omitempty,gte=1"`
	Limit int `json:"limit" form:"limit" validate:"omitempty,gte=1,lte=100"`
}

// Normalize sets default values for pagination parameters
func (p *PaginationParams) Normalize() {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.Limit < 1 {
		p.Limit = 20
	}
	if p.Limit > 100 {
		p.Limit = 100
	}
}

// GetOffset returns the database offset
func (p *PaginationParams) GetOffset() int {
	return GetOffset(p.Page, p.Limit)
}

// GetLimit returns the limit
func (p *PaginationParams) GetLimit() int {
	return p.Limit
}

// SortParams represents sorting request parameters
type SortParams struct {
	SortBy    string `json:"sort_by" form:"sort_by" validate:"omitempty"`
	SortOrder string `json:"sort_order" form:"sort_order" validate:"omitempty,oneof=asc desc"`
}

// Normalize sets default values for sorting parameters
func (s *SortParams) Normalize(defaultSortBy string) {
	if s.SortBy == "" {
		s.SortBy = defaultSortBy
	}
	if s.SortOrder == "" {
		s.SortOrder = "desc"
	}
}

// GetOrderBy returns the SQL ORDER BY clause
func (s *SortParams) GetOrderBy() string {
	if s.SortBy == "" {
		return ""
	}
	return s.SortBy + " " + s.SortOrder
}

// ListParams combines pagination and sorting
type ListParams struct {
	PaginationParams
	SortParams
}

// Normalize sets default values
func (l *ListParams) Normalize(defaultSortBy string) {
	l.PaginationParams.Normalize()
	l.SortParams.Normalize(defaultSortBy)
}
