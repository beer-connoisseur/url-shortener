package urlshort_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	generated "pkg/generated/urlshort/api/urlshort/v1"
	"urlshort/internal/controller/urlshort"
	"urlshort/internal/controller/urlshort/mocks"
	"urlshort/internal/entity"
)

func Test_urlshortServer_ShortenLink_validate(t *testing.T) {
	t.Parallel()

	type args struct {
		ctx context.Context
		req *generated.ShortenLinkRequest
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "correctness",
			args: args{
				ctx: t.Context(),
				req: &generated.ShortenLinkRequest{
					Link: "https://google.com",
				},
			},
			wantErr: false,
		},
		{
			name: "bad url",
			args: args{
				ctx: t.Context(),
				req: &generated.ShortenLinkRequest{
					Link: "bad",
				},
			},
			wantErr: true,
		},
		{
			name: "empty link",
			args: args{
				ctx: t.Context(),
				req: &generated.ShortenLinkRequest{},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.args.req.ValidateAll()

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func Test_urlshortServer_ShortenLink_generate(t *testing.T) {
	t.Parallel()

	type args struct {
		ctx context.Context
		req *generated.ShortenLinkRequest
	}

	tests := []struct {
		name        string
		args        args
		want        *generated.ShortenLinkResponse
		err         error
		getMockFunc func(controller *gomock.Controller) *mocks.MockurlshortService
	}{
		{
			name: "correctness",
			args: args{
				ctx: t.Context(),
				req: &generated.ShortenLinkRequest{
					Link: "https://google.com",
				},
			},
			want: &generated.ShortenLinkResponse{
				Link: "short-link",
			},
			err: nil,
			getMockFunc: func(controller *gomock.Controller) *mocks.MockurlshortService {
				serviceMock := mocks.NewMockurlshortService(controller)

				serviceMock.
					EXPECT().
					ShortenLink(t.Context(), "https://google.com").
					Return("short-link", nil).
					Times(1)

				return serviceMock
			},
		},
		{
			name: "bad url",
			args: args{
				ctx: t.Context(),
				req: &generated.ShortenLinkRequest{
					Link: "bad",
				},
			},
			want: nil,
			err: status.Error(
				codes.InvalidArgument,
				"invalid ShortenLinkRequest.Link: value must be absolute",
			),
			getMockFunc: func(controller *gomock.Controller) *mocks.MockurlshortService {
				serviceMock := mocks.NewMockurlshortService(controller)

				serviceMock.
					EXPECT().
					ShortenLink(gomock.Any(), gomock.Any()).
					Times(0)

				return serviceMock
			},
		},
		{
			name: "service error",
			args: args{
				ctx: t.Context(),
				req: &generated.ShortenLinkRequest{
					Link: "https://google.com",
				},
			},
			want: nil,
			err:  status.Error(codes.Internal, "internal error"),
			getMockFunc: func(controller *gomock.Controller) *mocks.MockurlshortService {
				serviceMock := mocks.NewMockurlshortService(controller)

				serviceMock.
					EXPECT().
					ShortenLink(t.Context(), "https://google.com").
					Return("", errors.New("internal error")).
					Times(1)

				return serviceMock
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			logger := zap.NewNop()

			server := urlshort.NewURLShortServer(
				logger,
				tt.getMockFunc(ctrl),
			)

			got, err := server.ShortenLink(tt.args.ctx, tt.args.req)

			if (err == nil) != (tt.err == nil) {
				t.Errorf("Error expected: %v, but got error: %v", tt.err != nil, err != nil)
				return
			}

			if err != nil && tt.err != nil {
				require.ErrorIs(t, tt.err, err)
			}

			require.Equal(t, tt.want, got)
		})
	}
}

func Test_urlshortServer_ExpandLink_validate(t *testing.T) {
	t.Parallel()

	type args struct {
		ctx context.Context
		req *generated.ExpandLinkRequest
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "correctness",
			args: args{
				ctx: t.Context(),
				req: &generated.ExpandLinkRequest{
					Link: "short_ly77",
				},
			},
			wantErr: false,
		},
		{
			name: "bad url",
			args: args{
				ctx: t.Context(),
				req: &generated.ExpandLinkRequest{
					Link: "bad/_",
				},
			},
			wantErr: true,
		},
		{
			name: "empty link",
			args: args{
				ctx: t.Context(),
				req: &generated.ExpandLinkRequest{},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.args.req.ValidateAll()

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func Test_urlshortServer_ExpandLink_generate(t *testing.T) {
	t.Parallel()

	type args struct {
		ctx context.Context
		req *generated.ExpandLinkRequest
	}

	tests := []struct {
		name        string
		args        args
		want        *generated.ExpandLinkResponse
		err         error
		getMockFunc func(controller *gomock.Controller) *mocks.MockurlshortService
	}{
		{
			name: "correctness",
			args: args{
				ctx: t.Context(),
				req: &generated.ExpandLinkRequest{
					Link: "httly_aabc",
				},
			},
			want: &generated.ExpandLinkResponse{
				Link: "https://google.com",
			},
			err: nil,
			getMockFunc: func(controller *gomock.Controller) *mocks.MockurlshortService {
				serviceMock := mocks.NewMockurlshortService(controller)

				serviceMock.
					EXPECT().
					ExpandLink(t.Context(), "httly_aabc").
					Return("https://google.com", nil).
					Times(1)

				return serviceMock
			},
		},
		{
			name: "bad url",
			args: args{
				ctx: t.Context(),
				req: &generated.ExpandLinkRequest{
					Link: "bad",
				},
			},
			want: nil,
			err: status.Error(
				codes.InvalidArgument,
				"invalid ExpandLinkRequest.Link: value length must be 10 runes",
			),
			getMockFunc: func(controller *gomock.Controller) *mocks.MockurlshortService {
				serviceMock := mocks.NewMockurlshortService(controller)

				serviceMock.
					EXPECT().
					ExpandLink(gomock.Any(), gomock.Any()).
					Times(0)

				return serviceMock
			},
		},
		{
			name: "short link not found",
			args: args{
				ctx: t.Context(),
				req: &generated.ExpandLinkRequest{
					Link: "short_ly11",
				},
			},
			want: nil,
			err:  status.Error(codes.NotFound, entity.ErrShortLinkNotFound.Error()),
			getMockFunc: func(controller *gomock.Controller) *mocks.MockurlshortService {
				serviceMock := mocks.NewMockurlshortService(controller)

				serviceMock.
					EXPECT().
					ExpandLink(t.Context(), "short_ly11").
					Return("", entity.ErrShortLinkNotFound).
					Times(1)

				return serviceMock
			},
		},
		{
			name: "internal error",
			args: args{
				ctx: t.Context(),
				req: &generated.ExpandLinkRequest{
					Link: "short_lybc",
				},
			},
			want: nil,
			err:  status.Error(codes.Internal, "internal error"),
			getMockFunc: func(controller *gomock.Controller) *mocks.MockurlshortService {
				serviceMock := mocks.NewMockurlshortService(controller)

				serviceMock.
					EXPECT().
					ExpandLink(t.Context(), "short_lybc").
					Return("", errors.New("internal error")).
					Times(1)

				return serviceMock
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			logger := zap.NewNop()

			server := urlshort.NewURLShortServer(
				logger,
				tt.getMockFunc(ctrl),
			)

			got, err := server.ExpandLink(tt.args.ctx, tt.args.req)

			if (err == nil) != (tt.err == nil) {
				t.Errorf("Error expected: %v, but got error: %v", tt.err != nil, err != nil)
				return
			}

			if err != nil && tt.err != nil {
				require.ErrorIs(t, tt.err, err)
			}

			require.Equal(t, tt.want, got)
		})
	}
}
