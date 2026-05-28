package urlshort

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
	"urlshort/internal/entity"
	"urlshort/internal/repository/urlshort/inmemory"
	"urlshort/internal/repository/urlshort/postgres"
)

type (
	urlshortRepository interface {
		GetShortLinkByOriginalLink(ctx context.Context, link string) (string, error)
		GetOriginalLinkByShortLink(ctx context.Context, link string) (string, error)
		CreateShortLink(ctx context.Context, link entity.Link) error
	}
)

func New(logger *zap.Logger, db *pgxpool.Pool, accessType string) (urlshortRepository, error) {
	const (
		POSTGRES = "POSTGRES"
		INMEMORY = "INMEMORY"
	)

	switch accessType {
	case POSTGRES:
		return postgres.NewPostgresRepository(logger, db), nil
	case INMEMORY:
		return inmemory.NewInMemoryRepository(logger), nil
	default:
		return nil, fmt.Errorf("unknown access type: %s", accessType)
	}
}
