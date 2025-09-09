package models

import (
	"time"
)

// LaserDisc represents a LaserDisc in the collection
type LaserDisc struct {
	ID            uint      `json:"id" gorm:"primaryKey"`
	UPC           string    `json:"upc" gorm:"uniqueIndex;not null" binding:"required"`
	Title         string    `json:"title" gorm:"not null" binding:"required"`
	Year          int       `json:"year"`
	Director      string    `json:"director"`
	Genre         string    `json:"genre"`
	Format        string    `json:"format"`        // CLV, CAV, etc.
	Sides         int       `json:"sides"`         // 1 or 2
	Runtime       int       `json:"runtime"`       // minutes
	CoverImageURL string    `json:"cover_image_url"`
	LDDBUrl       string    `json:"lddb_url"`
	Watched       bool      `json:"watched" gorm:"default:false"`
	Notes         string    `json:"notes"`
	AddedDate     time.Time `json:"added_date" gorm:"autoCreateTime"`
	UpdatedDate   time.Time `json:"updated_date" gorm:"autoUpdateTime"`
}

// TableName returns the table name for the LaserDisc model
func (LaserDisc) TableName() string {
	return "laserdiscs"
}

// CreateLaserDiscRequest represents the request payload for creating a LaserDisc
type CreateLaserDiscRequest struct {
	UPC           string `json:"upc" binding:"required"`
	Title         string `json:"title" binding:"required"`
	Year          int    `json:"year"`
	Director      string `json:"director"`
	Genre         string `json:"genre"`
	Format        string `json:"format"`
	Sides         int    `json:"sides"`
	Runtime       int    `json:"runtime"`
	CoverImageURL string `json:"cover_image_url"`
	LDDBUrl       string `json:"lddb_url"`
	Notes         string `json:"notes"`
}

// UpdateLaserDiscRequest represents the request payload for updating a LaserDisc
type UpdateLaserDiscRequest struct {
	Title         *string `json:"title"`
	Year          *int    `json:"year"`
	Director      *string `json:"director"`
	Genre         *string `json:"genre"`
	Format        *string `json:"format"`
	Sides         *int    `json:"sides"`
	Runtime       *int    `json:"runtime"`
	CoverImageURL *string `json:"cover_image_url"`
	LDDBUrl       *string `json:"lddb_url"`
	Watched       *bool   `json:"watched"`
	Notes         *string `json:"notes"`
}

// LookupResult represents the result of a UPC lookup from lddb.com
type LookupResult struct {
	UPC           string `json:"upc"`
	Title         string `json:"title"`
	Year          int    `json:"year"`
	Director      string `json:"director"`
	Genre         string `json:"genre"`
	Format        string `json:"format"`
	Sides         int    `json:"sides"`
	Runtime       int    `json:"runtime"`
	CoverImageURL string `json:"cover_image_url"`
	LDDBUrl       string `json:"lddb_url"`
	Found         bool   `json:"found"`
	Error         string `json:"error,omitempty"`
}