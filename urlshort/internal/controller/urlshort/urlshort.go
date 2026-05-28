package urlshort

import (
	"context"
	"errors"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"pkg/generated/urlshort/api/urlshort/v1"
	"urlshort/internal/entity"
)

var _ urlshort.UrlshortServer = (*urlshortServer)(nil)

type (
	urlshortService interface {
		ShortenLink(ctx context.Context, link string) (string, error)
		ExpandLink(ctx context.Context, link string) (string, error)
	}
)

//go:generate mockgen -source=urlshort.go -destination=mocks/mock.go -package=mocks

type urlshortServer struct {
	logger          *zap.Logger
	urlshortService urlshortService
}

func NewURLShortServer(logger *zap.Logger, urlshortService urlshortService) *urlshortServer {
	return &urlshortServer{
		logger:          logger,
		urlshortService: urlshortService,
	}
}

func (s *urlshortServer) ShortenLink(ctx context.Context, req *urlshort.ShortenLinkRequest) (*urlshort.ShortenLinkResponse, error) {
	s.logger.Info("rpc ShortenLinkCalled")

	if err := req.ValidateAll(); err != nil {
		s.logger.Error("validation failed", zap.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	shortLink, err := s.urlshortService.ShortenLink(ctx, req.GetLink())
	if err != nil {
		s.logger.Error("shorten link failed", zap.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &urlshort.ShortenLinkResponse{
		Link: shortLink,
	}, nil
}

func (s *urlshortServer) ExpandLink(ctx context.Context, req *urlshort.ExpandLinkRequest) (*urlshort.ExpandLinkResponse, error) {
	s.logger.Info("rpc ExpandLinkCalled")

	if err := req.ValidateAll(); err != nil {
		s.logger.Error("validation failed", zap.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	originalLink, err := s.urlshortService.ExpandLink(ctx, req.GetLink())
	if err != nil {
		s.logger.Error("expand link failed", zap.Error(err))
		if errors.Is(err, entity.ErrShortLinkNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &urlshort.ExpandLinkResponse{
		Link: originalLink,
	}, nil
}
