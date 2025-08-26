package query

import (
	"context"
	"time"

	"github.com/danghamo/life/internal/domain/shared"
)

// Query represents a query in CQRS pattern
type Query interface {
	QueryID() string
	QueryType() string
	CreatedAt() time.Time
}

// BaseQuery provides common query functionality
type BaseQuery struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	Timestamp time.Time `json:"created_at"`
}

// QueryID returns the query ID
func (q BaseQuery) QueryID() string {
	return q.ID
}

// QueryType returns the query type
func (q BaseQuery) QueryType() string {
	return q.Type
}

// CreatedAt returns when the query was created
func (q BaseQuery) CreatedAt() time.Time {
	return q.Timestamp
}

// NewBaseQuery creates a new base query
func NewBaseQuery(queryType string) BaseQuery {
	return BaseQuery{
		ID:        shared.NewID().String(),
		Type:      queryType,
		Timestamp: time.Now(),
	}
}

// QueryHandler handles queries
type QueryHandler interface {
	Handle(ctx context.Context, query Query) (interface{}, error)
}

// QueryBus dispatches queries to handlers
type QueryBus interface {
	Send(ctx context.Context, query Query) (interface{}, error)
	Register(queryType string, handler QueryHandler)
}

// QueryResult represents the result of a query execution
type QueryResult struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   error       `json:"error,omitempty"`
}

// NewSuccessQueryResult creates a successful query result
func NewSuccessQueryResult(data interface{}) QueryResult {
	return QueryResult{
		Success: true,
		Data:    data,
	}
}

// NewErrorQueryResult creates an error query result
func NewErrorQueryResult(err error) QueryResult {
	return QueryResult{
		Success: false,
		Error:   err,
	}
}

// Pagination represents pagination parameters
type Pagination struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
	Offset   int `json:"offset"`
}

// NewPagination creates a new pagination
func NewPagination(page, pageSize int) Pagination {
	offset := (page - 1) * pageSize
	if offset < 0 {
		offset = 0
	}
	return Pagination{
		Page:     page,
		PageSize: pageSize,
		Offset:   offset,
	}
}

// PaginatedResult represents a paginated query result
type PaginatedResult struct {
	Data       interface{} `json:"data"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	TotalCount int         `json:"total_count"`
	TotalPages int         `json:"total_pages"`
}

// NewPaginatedResult creates a new paginated result
func NewPaginatedResult(data interface{}, page, pageSize, totalCount int) PaginatedResult {
	totalPages := (totalCount + pageSize - 1) / pageSize
	return PaginatedResult{
		Data:       data,
		Page:       page,
		PageSize:   pageSize,
		TotalCount: totalCount,
		TotalPages: totalPages,
	}
}
