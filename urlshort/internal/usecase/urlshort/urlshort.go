package urlshort

import (
	"context"
	"errors"

	"urlshort/internal/entity"
	"urlshort/internal/usecase/generator"
)

type (
	urlshortRepository interface {
		GetShortLinkByOriginalLink(ctx context.Context, link string) (string, error)
		GetOriginalLinkByShortLink(ctx context.Context, link string) (string, error)
		CreateShortLink(ctx context.Context, link entity.Link) error
	}
)

//go:generate mockgen -source=urlshort.go -destination=mocks/mock.go -package=mocks

type urlshortService struct {
	urlshortRepository urlshortRepository
}

func NewURLShortService(urlshortRepository urlshortRepository) *urlshortService {
	return &urlshortService{
		urlshortRepository: urlshortRepository,
	}
}

func (s *urlshortService) ShortenLink(ctx context.Context, link string) (string, error) {
	const maxAttempts = 5

	for i := 0; i < maxAttempts; i++ {
		shortLink, err := generator.GenerateShortLink()
		if err != nil {
			return "", err
		}

		err = s.urlshortRepository.CreateShortLink(ctx, entity.Link{
			OriginalLink: link,
			ShortLink:    shortLink,
		})
		if err == nil {
			return shortLink, nil
		}

		if errors.Is(err, entity.ErrShortLinkAlreadyExists) {
			continue
		}

		if errors.Is(err, entity.ErrOriginalLinkAlreadyExists) {
			shortLink, err = s.urlshortRepository.GetShortLinkByOriginalLink(ctx, link)
			if err != nil {
				return "", err
			}

			return shortLink, nil
		}

		return "", err
	}

	return "", errors.New("failed to generate unique short link")
}

func (s *urlshortService) ExpandLink(ctx context.Context, link string) (string, error) {
	return s.urlshortRepository.GetOriginalLinkByShortLink(ctx, link)
}
