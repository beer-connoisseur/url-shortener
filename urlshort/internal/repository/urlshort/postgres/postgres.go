package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"urlshort/internal/entity"
)

type postgresRepository struct {
	logger *zap.Logger
	db     *pgxpool.Pool
}

func NewPostgresRepository(logger *zap.Logger, db *pgxpool.Pool) *postgresRepository {
	return &postgresRepository{
		logger: logger,
		db:     db,
	}
}

func (p *postgresRepository) GetShortLinkByOriginalLink(ctx context.Context, link string) (string, error) {
	const query = `
		SELECT short_link FROM link.links
        WHERE original_link = $1
	`
	row := p.db.QueryRow(ctx, query, link)

	var shortLink string
	err := row.Scan(&shortLink)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", entity.ErrOriginalLinkNotFound
		}

		return "", err
	}

	return shortLink, nil
}

func (p *postgresRepository) GetOriginalLinkByShortLink(ctx context.Context, link string) (string, error) {
	const query = `
		SELECT original_link FROM link.links
        WHERE short_link = $1
	`
	row := p.db.QueryRow(ctx, query, link)

	var originalLink string
	err := row.Scan(&originalLink)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", entity.ErrShortLinkNotFound
		}

		return "", err
	}

	return originalLink, nil
}

func (p *postgresRepository) CreateShortLink(ctx context.Context, link entity.Link) error {
	const query = `
		INSERT INTO link.links (original_link, short_link)
        VALUES ($1, $2)
	`

	_, err := p.db.Exec(ctx, query, link.OriginalLink, link.ShortLink)
	if err != nil {
		return checkExistsKeyViolation(err)
	}

	return nil
}

func checkExistsKeyViolation(err error) error {
	const ErrExistsKeyViolation = "23505"
	var pgErr *pgconn.PgError

	if errors.As(err, &pgErr) && pgErr.Code == ErrExistsKeyViolation {
		switch pgErr.ConstraintName {
		case "unique_original_link":
			return entity.ErrOriginalLinkAlreadyExists
		case "unique_short_link":
			return entity.ErrShortLinkAlreadyExists
		}
	}

	return err
}
