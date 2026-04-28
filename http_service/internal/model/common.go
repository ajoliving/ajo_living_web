/*
 * Shared data model primitives.
 * 1. Define reusable timestamp and pagination structures.
 * 2. Keep common database fields consistent across tables.
 */
package model

import "time"

// 1. TimestampModel defines shared created and updated timestamps.
type TimestampModel struct {
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// 2. Pagination defines a shared pagination payload.
type Pagination struct {
	Page     int   `json:"page"`
	PageSize int   `json:"page_size"`
	Total    int64 `json:"total"`
}
