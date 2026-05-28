package urlshort_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"urlshort/internal/entity"
	"urlshort/internal/usecase/urlshort"
	"urlshort/internal/usecase/urlshort/mocks"
)

func Test_urlshortService_ShortenLink(t *testing.T) {
	t.Parallel()

	type args struct {
		ctx  context.Context
		link string
	}

	tests := []struct {
		name       string
		args       args
		want       string
		wantErr    bool
		setupMocks func(repo *mocks.MockurlshortRepository)
	}{
		{
			name: "successful create",
			args: args{
				ctx:  t.Context(),
				link: "https://google.com",
			},
			wantErr: false,
			setupMocks: func(repo *mocks.MockurlshortRepository) {
				repo.
					EXPECT().
					CreateShortLink(
						t.Context(),
						gomock.AssignableToTypeOf(entity.Link{}),
					).
					DoAndReturn(func(_ context.Context, link entity.Link) error {
						require.Equal(t, "https://google.com", link.OriginalLink)
						require.NotEmpty(t, link.ShortLink)

						return nil
					}).
					Times(1)

				repo.
					EXPECT().
					GetShortLinkByOriginalLink(gomock.Any(), gomock.Any()).
					Times(0)
			},
		},
		{
			name: "original link already exists",
			args: args{
				ctx:  t.Context(),
				link: "https://google.com",
			},
			want:    "existing-short",
			wantErr: false,
			setupMocks: func(repo *mocks.MockurlshortRepository) {
				repo.
					EXPECT().
					CreateShortLink(
						t.Context(),
						gomock.AssignableToTypeOf(entity.Link{}),
					).
					Return(entity.ErrOriginalLinkAlreadyExists).
					Times(1)

				repo.
					EXPECT().
					GetShortLinkByOriginalLink(t.Context(), "https://google.com").
					Return("existing-short", nil).
					Times(1)
			},
		},
		{
			name: "short link collision then success",
			args: args{
				ctx:  t.Context(),
				link: "https://google.com",
			},
			wantErr: false,
			setupMocks: func(repo *mocks.MockurlshortRepository) {
				gomock.InOrder(
					repo.
						EXPECT().
						CreateShortLink(
							t.Context(),
							gomock.AssignableToTypeOf(entity.Link{}),
						).
						Return(entity.ErrShortLinkAlreadyExists),

					repo.
						EXPECT().
						CreateShortLink(
							t.Context(),
							gomock.AssignableToTypeOf(entity.Link{}),
						).
						Return(nil),
				)

				repo.
					EXPECT().
					GetShortLinkByOriginalLink(gomock.Any(), gomock.Any()).
					Times(0)
			},
		},
		{
			name: "all attempts exhausted",
			args: args{
				ctx:  t.Context(),
				link: "https://google.com",
			},
			want:    "",
			wantErr: true,
			setupMocks: func(repo *mocks.MockurlshortRepository) {
				repo.
					EXPECT().
					CreateShortLink(
						t.Context(),
						gomock.AssignableToTypeOf(entity.Link{}),
					).
					Return(entity.ErrShortLinkAlreadyExists).
					Times(5)

				repo.
					EXPECT().
					GetShortLinkByOriginalLink(gomock.Any(), gomock.Any()).
					Times(0)
			},
		},
		{
			name: "repository internal error",
			args: args{
				ctx:  t.Context(),
				link: "https://google.com",
			},
			want:    "",
			wantErr: true,
			setupMocks: func(repo *mocks.MockurlshortRepository) {
				repo.
					EXPECT().
					CreateShortLink(
						t.Context(),
						gomock.AssignableToTypeOf(entity.Link{}),
					).
					Return(errors.New("internal error")).
					Times(1)

				repo.
					EXPECT().
					GetShortLinkByOriginalLink(gomock.Any(), gomock.Any()).
					Times(0)
			},
		},
		{
			name: "get existing short link failed",
			args: args{
				ctx:  t.Context(),
				link: "https://google.com",
			},
			want:    "",
			wantErr: true,
			setupMocks: func(repo *mocks.MockurlshortRepository) {
				repo.
					EXPECT().
					CreateShortLink(
						t.Context(),
						gomock.AssignableToTypeOf(entity.Link{}),
					).
					Return(entity.ErrOriginalLinkAlreadyExists).
					Times(1)

				repo.
					EXPECT().
					GetShortLinkByOriginalLink(t.Context(), "https://google.com").
					Return("", errors.New("internal error")).
					Times(1)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			repo := mocks.NewMockurlshortRepository(ctrl)

			tt.setupMocks(repo)

			service := urlshort.NewURLShortService(repo)

			got, err := service.ShortenLink(tt.args.ctx, tt.args.link)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			if tt.want != "" {
				require.Equal(t, tt.want, got)
			} else if !tt.wantErr {
				require.NotEmpty(t, got)
			}
		})
	}
}

func Test_urlshortService_ExpandLink(t *testing.T) {
	t.Parallel()

	type args struct {
		ctx  context.Context
		link string
	}

	tests := []struct {
		name       string
		args       args
		want       string
		wantErr    bool
		setupMocks func(repo *mocks.MockurlshortRepository)
	}{
		{
			name: "correctness",
			args: args{
				ctx:  t.Context(),
				link: "short123_1",
			},
			want:    "https://google.com",
			wantErr: false,
			setupMocks: func(repo *mocks.MockurlshortRepository) {
				repo.
					EXPECT().
					GetOriginalLinkByShortLink(t.Context(), "short123_1").
					Return("https://google.com", nil).
					Times(1)
			},
		},
		{
			name: "short link not found",
			args: args{
				ctx:  t.Context(),
				link: "short123_1",
			},
			want:    "",
			wantErr: true,
			setupMocks: func(repo *mocks.MockurlshortRepository) {
				repo.
					EXPECT().
					GetOriginalLinkByShortLink(t.Context(), "short123_1").
					Return("", entity.ErrShortLinkNotFound).
					Times(1)
			},
		},
		{
			name: "internal error",
			args: args{
				ctx:  t.Context(),
				link: "short123_1",
			},
			want:    "",
			wantErr: true,
			setupMocks: func(repo *mocks.MockurlshortRepository) {
				repo.
					EXPECT().
					GetOriginalLinkByShortLink(t.Context(), "short123_1").
					Return("", errors.New("internal error")).
					Times(1)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			repo := mocks.NewMockurlshortRepository(ctrl)

			tt.setupMocks(repo)

			service := urlshort.NewURLShortService(repo)

			got, err := service.ExpandLink(tt.args.ctx, tt.args.link)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, tt.want, got)
		})
	}
}
