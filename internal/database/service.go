package database

import (
	"errors"
	"math/rand"

	"gorm.io/gorm"

	"github.com/paran01d/lddb/internal/models"
)

// Service handles all database operations
type Service struct {
	db *gorm.DB
}

// NewService creates a new database service
func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

// GetAllLaserDiscs retrieves all LaserDiscs from the database
func (s *Service) GetAllLaserDiscs() ([]models.LaserDisc, error) {
	var laserdiscs []models.LaserDisc
	result := s.db.Order("title ASC").Find(&laserdiscs)
	return laserdiscs, result.Error
}

// GetLaserDiscByID retrieves a LaserDisc by its ID
func (s *Service) GetLaserDiscByID(id uint) (*models.LaserDisc, error) {
	var laserdisc models.LaserDisc
	result := s.db.First(&laserdisc, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &laserdisc, nil
}

// GetLaserDiscByUPC retrieves a LaserDisc by its UPC
func (s *Service) GetLaserDiscByUPC(upc string) (*models.LaserDisc, error) {
	var laserdisc models.LaserDisc
	result := s.db.Where("upc = ?", upc).First(&laserdisc)
	if result.Error != nil {
		return nil, result.Error
	}
	return &laserdisc, nil
}

// CreateLaserDisc creates a new LaserDisc in the database
func (s *Service) CreateLaserDisc(req *models.CreateLaserDiscRequest) (*models.LaserDisc, error) {
	// Check if UPC already exists
	var existing models.LaserDisc
	result := s.db.Where("upc = ?", req.UPC).First(&existing)
	if result.Error == nil {
		return nil, errors.New("laserdisc with this UPC already exists")
	}

	laserdisc := &models.LaserDisc{
		UPC:           req.UPC,
		Title:         req.Title,
		Year:          req.Year,
		Director:      req.Director,
		Genre:         req.Genre,
		Format:        req.Format,
		Sides:         req.Sides,
		Runtime:       req.Runtime,
		CoverImageURL: req.CoverImageURL,
		Notes:         req.Notes,
	}

	result = s.db.Create(laserdisc)
	if result.Error != nil {
		return nil, result.Error
	}

	return laserdisc, nil
}

// UpdateLaserDisc updates an existing LaserDisc
func (s *Service) UpdateLaserDisc(id uint, req *models.UpdateLaserDiscRequest) (*models.LaserDisc, error) {
	var laserdisc models.LaserDisc
	result := s.db.First(&laserdisc, id)
	if result.Error != nil {
		return nil, result.Error
	}

	// Update only non-nil fields
	updates := make(map[string]interface{})
	
	if req.Title != nil {
		updates["title"] = *req.Title
	}
	if req.Year != nil {
		updates["year"] = *req.Year
	}
	if req.Director != nil {
		updates["director"] = *req.Director
	}
	if req.Genre != nil {
		updates["genre"] = *req.Genre
	}
	if req.Format != nil {
		updates["format"] = *req.Format
	}
	if req.Sides != nil {
		updates["sides"] = *req.Sides
	}
	if req.Runtime != nil {
		updates["runtime"] = *req.Runtime
	}
	if req.CoverImageURL != nil {
		updates["cover_image_url"] = *req.CoverImageURL
	}
	if req.Watched != nil {
		updates["watched"] = *req.Watched
	}
	if req.Notes != nil {
		updates["notes"] = *req.Notes
	}

	result = s.db.Model(&laserdisc).Updates(updates)
	if result.Error != nil {
		return nil, result.Error
	}

	return &laserdisc, nil
}

// DeleteLaserDisc deletes a LaserDisc from the database
func (s *Service) DeleteLaserDisc(id uint) error {
	result := s.db.Delete(&models.LaserDisc{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// ToggleWatched toggles the watched status of a LaserDisc
func (s *Service) ToggleWatched(id uint) (*models.LaserDisc, error) {
	var laserdisc models.LaserDisc
	result := s.db.First(&laserdisc, id)
	if result.Error != nil {
		return nil, result.Error
	}

	laserdisc.Watched = !laserdisc.Watched
	result = s.db.Save(&laserdisc)
	if result.Error != nil {
		return nil, result.Error
	}

	return &laserdisc, nil
}

// GetRandomUnwatched returns a random unwatched LaserDisc
func (s *Service) GetRandomUnwatched() (*models.LaserDisc, error) {
	var unwatched []models.LaserDisc
	result := s.db.Where("watched = ?", false).Find(&unwatched)
	if result.Error != nil {
		return nil, result.Error
	}

	if len(unwatched) == 0 {
		return nil, errors.New("no unwatched laserdiscs found")
	}

	// Return random unwatched LaserDisc
	randomIndex := rand.Intn(len(unwatched))
	return &unwatched[randomIndex], nil
}

// SearchLaserDiscs searches for LaserDiscs by title, director, or genre
func (s *Service) SearchLaserDiscs(query string) ([]models.LaserDisc, error) {
	var laserdiscs []models.LaserDisc
	searchPattern := "%" + query + "%"
	
	result := s.db.Where(
		"title LIKE ? OR director LIKE ? OR genre LIKE ?",
		searchPattern, searchPattern, searchPattern,
	).Order("title ASC").Find(&laserdiscs)
	
	return laserdiscs, result.Error
}

// GetStats returns collection statistics
func (s *Service) GetStats() (map[string]interface{}, error) {
	var total, watched, unwatched int64
	
	// Get total count
	result := s.db.Model(&models.LaserDisc{}).Count(&total)
	if result.Error != nil {
		return nil, result.Error
	}
	
	// Get watched count
	result = s.db.Model(&models.LaserDisc{}).Where("watched = ?", true).Count(&watched)
	if result.Error != nil {
		return nil, result.Error
	}
	
	unwatched = total - watched
	
	stats := map[string]interface{}{
		"total":     total,
		"watched":   watched,
		"unwatched": unwatched,
	}
	
	return stats, nil
}