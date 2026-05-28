package integration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/network"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	grpcPort = "50051"
	httpPort = "8080"
	dbPort   = "5432"
)

func TestIntegrationWithPostgres(t *testing.T) {
	ctx := t.Context()

	net, err := network.New(ctx, network.WithDriver("bridge"))
	require.NoError(t, err)
	t.Cleanup(func() {
		if err = net.Remove(ctx); err != nil {
			t.Logf("failed to remove network: %v", err)
		}
	})

	postgresC, err := runPostgres(ctx, net.Name)
	require.NoError(t, err)
	t.Cleanup(func() {
		if err = postgresC.Terminate(ctx); err != nil {
			t.Logf("failed to terminate postgres: %v", err)
		}
	})

	appC, err := runApp(ctx, net.Name, "POSTGRES")
	require.NoError(t, err)
	t.Cleanup(func() {
		if err = appC.Terminate(ctx); err != nil {
			t.Logf("failed to terminate app: %v", err)
		}
	})

	appHost, err := appC.Host(ctx)
	require.NoError(t, err)
	appPort, err := appC.MappedPort(ctx, httpPort)
	require.NoError(t, err)
	baseURL := fmt.Sprintf("http://%s:%s", appHost, appPort.Port())

	time.Sleep(1 * time.Second)

	t.Run("API flow", func(t *testing.T) {
		testAPI(t, baseURL)
	})
}

func TestIntegrationWithInMemory(t *testing.T) {
	ctx := t.Context()

	net, err := network.New(ctx, network.WithDriver("bridge"))
	require.NoError(t, err)
	t.Cleanup(func() {
		if err = net.Remove(ctx); err != nil {
			t.Logf("failed to remove network: %v", err)
		}
	})

	appC, err := runApp(ctx, net.Name, "INMEMORY")
	require.NoError(t, err)
	t.Cleanup(func() {
		if err = appC.Terminate(ctx); err != nil {
			t.Logf("failed to terminate app: %v", err)
		}
	})

	appHost, err := appC.Host(ctx)
	require.NoError(t, err)
	appPort, err := appC.MappedPort(ctx, httpPort)
	require.NoError(t, err)
	baseURL := fmt.Sprintf("http://%s:%s", appHost, appPort.Port())

	time.Sleep(1 * time.Second)

	t.Run("API flow", func(t *testing.T) {
		testAPI(t, baseURL)
	})
}

func runPostgres(ctx context.Context, networkName string) (testcontainers.Container, error) {
	req := testcontainers.ContainerRequest{
		Image:        "postgres:17",
		ExposedPorts: []string{dbPort + "/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "postgres",
			"POSTGRES_PASSWORD": "12345",
			"POSTGRES_DB":       "urls",
		},
		WaitingFor: wait.ForListeningPort(dbPort + "/tcp").WithPollInterval(1 * time.Second),
		Networks:   []string{networkName},
		NetworkAliases: map[string][]string{
			networkName: {"postgres"},
		},
	}
	return testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
}

func runApp(ctx context.Context, networkName, accessType string) (testcontainers.Container, error) {
	req := testcontainers.ContainerRequest{
		FromDockerfile: testcontainers.FromDockerfile{
			Context:        "../",
			Dockerfile:     "./urlshort/Dockerfile",
			BuildLogWriter: os.Stdout,
		},
		ExposedPorts: []string{httpPort + "/tcp"},
		Env: map[string]string{
			"GRPC_PORT":         grpcPort,
			"GRPC_GATEWAY_PORT": httpPort,
			"POSTGRES_HOST":     "postgres",
			"POSTGRES_PORT":     dbPort,
			"POSTGRES_USER":     "postgres",
			"POSTGRES_PASSWORD": "12345",
			"POSTGRES_DB":       "urls",
			"POSTGRES_MAX_CONN": "10",
			"ACCESS_TYPE":       accessType,
		},
		WaitingFor: wait.ForListeningPort(httpPort + "/tcp").WithPollInterval(1 * time.Second),
		Networks:   []string{networkName},
		NetworkAliases: map[string][]string{
			networkName: {"app"},
		},
	}
	return testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
}

func testAPI(t *testing.T, baseURL string) {
	doRequest := func(method, path string, body interface{}) *http.Response {
		var buf bytes.Buffer
		if body != nil {
			err := json.NewEncoder(&buf).Encode(body)
			require.NoError(t, err)
		}
		req, err := http.NewRequest(method, baseURL+path, &buf)
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		return resp
	}

	t.Run("shorten and expand link", func(t *testing.T) {
		originalLink := "https://google.com/some/very/long/url"

		reqBody := map[string]interface{}{
			"link": originalLink,
		}

		resp := doRequest(
			http.MethodPost,
			"/v1/urlshort/shorten",
			reqBody,
		)
		defer func() {
			err := resp.Body.Close()
			if err != nil {
				t.Logf("failed to close response body: %v", err)
			}
		}()

		require.Equal(
			t,
			http.StatusOK,
			resp.StatusCode,
		)

		var shortenResp struct {
			Link string `json:"link"`
		}

		err := json.NewDecoder(resp.Body).Decode(&shortenResp)
		require.NoError(t, err)

		require.NotEmpty(t, shortenResp.Link)

		assert.Len(t, shortenResp.Link, 10)

		matched, err := regexp.MatchString(
			`^[a-zA-Z0-9_]{10}$`,
			shortenResp.Link,
		)

		require.NoError(t, err)
		assert.True(t, matched)

		resp = doRequest(
			http.MethodGet,
			"/v1/urlshort/expand/"+shortenResp.Link,
			nil,
		)
		defer func() {
			err = resp.Body.Close()
			if err != nil {
				t.Logf("failed to close response body: %v", err)
			}
		}()

		require.Equal(
			t,
			http.StatusOK,
			resp.StatusCode,
		)

		var expandResp struct {
			Link string `json:"link"`
		}

		err = json.NewDecoder(resp.Body).Decode(&expandResp)
		require.NoError(t, err)

		assert.Equal(
			t,
			originalLink,
			expandResp.Link,
		)
	})

	t.Run("same original link returns same short link", func(t *testing.T) {
		originalLink := "https://youtube.com/watch?v=test"

		reqBody := map[string]interface{}{
			"link": originalLink,
		}

		resp1 := doRequest(
			http.MethodPost,
			"/v1/urlshort/shorten",
			reqBody,
		)
		defer func() {
			err := resp1.Body.Close()
			if err != nil {
				t.Logf("failed to close response body: %v", err)
			}
		}()

		require.Equal(
			t,
			http.StatusOK,
			resp1.StatusCode,
		)

		var r1 struct {
			Link string `json:"link"`
		}

		err := json.NewDecoder(resp1.Body).Decode(&r1)
		require.NoError(t, err)

		resp2 := doRequest(
			http.MethodPost,
			"/v1/urlshort/shorten",
			reqBody,
		)
		defer func() {
			err = resp2.Body.Close()
			if err != nil {
				t.Logf("failed to close response body: %v", err)
			}
		}()

		require.Equal(
			t,
			http.StatusOK,
			resp2.StatusCode,
		)

		var r2 struct {
			Link string `json:"link"`
		}

		err = json.NewDecoder(resp2.Body).Decode(&r2)
		require.NoError(t, err)

		assert.Equal(t, r1.Link, r2.Link)
	})

	t.Run("expand non existent link", func(t *testing.T) {
		resp := doRequest(
			http.MethodGet,
			"/v1/urlshort/expand/A1AAA_2BBB",
			nil,
		)
		defer func() {
			err := resp.Body.Close()
			if err != nil {
				t.Logf("failed to close response body: %v", err)
			}
		}()

		assert.Equal(
			t,
			http.StatusNotFound,
			resp.StatusCode,
		)
	})

	t.Run("invalid url", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"url": "not-a-valid-url",
		}

		resp := doRequest(
			http.MethodPost,
			"/v1/urlshort/shorten",
			reqBody,
		)
		defer func() {
			err := resp.Body.Close()
			if err != nil {
				t.Logf("failed to close response body: %v", err)
			}
		}()

		assert.Equal(
			t,
			http.StatusBadRequest,
			resp.StatusCode,
		)
	})

	t.Run("empty url", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"url": "",
		}

		resp := doRequest(
			http.MethodPost,
			"/v1/urlshort/shorten",
			reqBody,
		)
		defer func() {
			err := resp.Body.Close()
			if err != nil {
				t.Logf("failed to close response body: %v", err)
			}
		}()

		assert.Equal(
			t,
			http.StatusBadRequest,
			resp.StatusCode,
		)
	})

	t.Run("concurrent shorten same link", func(t *testing.T) {
		const goroutines = 2

		originalLink := "https://concurrent-test.com/page"
		reqBody := map[string]interface{}{
			"link": originalLink,
		}

		type result struct {
			shortLink string
			err       error
			status    int
		}

		results := make(chan result, goroutines)
		var wg sync.WaitGroup

		wg.Add(goroutines)
		for range goroutines {
			go func() {
				defer wg.Done()

				resp := doRequest(
					http.MethodPost,
					"/v1/urlshort/shorten",
					reqBody,
				)
				defer func() {
					err := resp.Body.Close()
					if err != nil {
						t.Logf("failed to close response body: %v", err)
					}
				}()

				var shortenResp struct {
					Link string `json:"link"`
				}

				err := json.NewDecoder(resp.Body).Decode(&shortenResp)

				results <- result{
					shortLink: shortenResp.Link,
					err:       err,
					status:    resp.StatusCode,
				}
			}()
		}

		wg.Wait()
		close(results)

		var links []string
		for r := range results {
			require.NoError(t, r.err)

			require.Equal(t, http.StatusOK, r.status)

			require.NotEmpty(t, r.shortLink)

			links = append(links, r.shortLink)
		}

		require.Len(t, links, goroutines)

		assert.Equal(t, links[0], links[1], "both users should receive same short link")
	})
}
