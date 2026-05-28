package inmemory

import (
	"context"
	"sync"

	"go.uber.org/zap"
	"urlshort/internal/entity"
)

type inMemoryRepository struct {
	logger      *zap.Logger
	mx          sync.RWMutex
	shortToLong map[string]string // short_link -> original_link
	longToShort map[string]string // original_link -> short_link
}

func NewInMemoryRepository(logger *zap.Logger) *inMemoryRepository {
	return &inMemoryRepository{
		logger:      logger,
		shortToLong: make(map[string]string),
		longToShort: make(map[string]string),
	}
}

func (r *inMemoryRepository) GetShortLinkByOriginalLink(_ context.Context, link string) (string, error) {
	r.mx.RLock()
	defer r.mx.RUnlock()

	shortLink, exists := r.longToShort[link]
	if !exists {
		return "", entity.ErrOriginalLinkNotFound
	}

	return shortLink, nil
}

func (r *inMemoryRepository) GetOriginalLinkByShortLink(_ context.Context, link string) (string, error) {
	r.mx.RLock()
	defer r.mx.RUnlock()

	originalLink, exists := r.shortToLong[link]
	if !exists {
		return "", entity.ErrShortLinkNotFound
	}

	return originalLink, nil
}

func (r *inMemoryRepository) CreateShortLink(_ context.Context, link entity.Link) error {
	r.mx.Lock()
	defer r.mx.Unlock()

	_, exists := r.longToShort[link.OriginalLink]
	if exists {
		return entity.ErrOriginalLinkAlreadyExists
	}

	_, exists = r.shortToLong[link.ShortLink]
	if exists {
		return entity.ErrShortLinkAlreadyExists
	}

	r.longToShort[link.OriginalLink] = link.ShortLink
	r.shortToLong[link.ShortLink] = link.OriginalLink

	return nil
}
