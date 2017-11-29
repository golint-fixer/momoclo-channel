package persistence

import (
	"github.com/utahta/momoclo-channel/domain"
	"github.com/utahta/momoclo-channel/domain/model"
)

// LatestEntryRepository operates LatestEntry entity
type LatestEntryRepository struct {
	model.PersistenceHandler
}

// NewLatestEntryRepository returns the LatestEntryRepository
func NewLatestEntryRepository(h model.PersistenceHandler) model.LatestEntryRepository {
	return &LatestEntryRepository{h}
}

// Save saves LatestEntry
func (repo *LatestEntryRepository) Save(l *model.LatestEntry) error {
	return repo.Put(l)
}

// FindOrNewByURL finds LatestEntry given url
// if not found, returns new LatestEntry
func (repo *LatestEntryRepository) FindOrNewByURL(urlStr string) (*model.LatestEntry, error) {
	l, err := model.NewLatestEntry(urlStr)
	if err != nil {
		return nil, err
	}

	err = repo.Get(l)
	if err == domain.ErrNoSuchEntity {
		return l, nil
	}
	if err != nil {
		return nil, err
	}
	return l, nil
}

// GetURL returns URL given code
func (repo *LatestEntryRepository) GetURL(code string) string {
	l := &model.LatestEntry{ID: code}
	if err := repo.Get(l); err != nil {
		return ""
	}
	return l.URL
}