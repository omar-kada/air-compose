package server

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"omar-kada/air-compose/api"
	"omar-kada/air-compose/internal/config"
	"omar-kada/air-compose/internal/deployments"
	"omar-kada/air-compose/internal/docker"
	"omar-kada/air-compose/internal/events"
	"omar-kada/air-compose/internal/git"
	"omar-kada/air-compose/internal/logs"
	"omar-kada/air-compose/internal/models"
	"omar-kada/air-compose/internal/process"
	"omar-kada/air-compose/internal/server/handlers"
	"omar-kada/air-compose/internal/server/socket"
	"omar-kada/air-compose/internal/shell"
	"omar-kada/air-compose/internal/users"
	"omar-kada/air-compose/testutil"
)

// testInspector is a minimal docker.Inspector for tests (no Docker daemon needed).
type testInspector struct{}

func (testInspector) GetManagedStacks() (models.StacksState, error) {
	return models.NewStacksState(), nil
}
func (testInspector) GetCurrentStacks(_ []string) (models.StacksState, error) {
	return models.NewStacksState(), nil
}

// serverDeps holds real implementations for all Serve dependencies.
type serverDeps struct {
	businessHandler api.StrictServerInterface
	socketHandler   socket.WebSocketHandler
	userService     users.Service
	oidcService     users.OidcService
	configStore     config.Store
}

func newServerDeps(t *testing.T) *serverDeps {
	t.Helper()

	db := testutil.NewMemoryStorage(t)
	eventBus := events.NewBus(1)

	configStore, err := config.NewConfigStore(
		filepath.Join(t.TempDir(), "config.yaml"), eventBus,
	)
	require.NoError(t, err)

	eventStore, err := events.NewEventStorage(db)
	require.NoError(t, err, "event store")

	deploymentStore, err := deployments.NewDeploymentStorage(db)
	require.NoError(t, err, "deployment store")

	userStore, err := users.NewUsersStorage(db)
	require.NoError(t, err, "user store")

	sessionStore, err := users.NewSessionStorage(db)
	require.NoError(t, err, "session store")

	authStore, err := users.NewAuthStorage(userStore, sessionStore, users.NewTokenHolder())
	require.NoError(t, err, "auth store")

	fetcher := git.NewFetcher(0o777, t.TempDir(), configStore)
	deployer := docker.NewDeployer(eventBus, shell.NewExecutor())
	deploymentSvc := process.NewDeploymentService(
		models.DeploymentParams{}, deployer, fetcher,
		deploymentStore, configStore, eventBus,
	)
	repoWatcher := process.NewRepoWatcher(
		fetcher, configStore, deploymentSvc, eventBus, process.NewCronScheduler(),
	)
	healthChecker := docker.NewHealthChecker(configStore, testInspector{}, eventBus)

	userService := users.NewService(authStore, eventBus)

	bh := handlers.NewBusinessHandler(
		configStore, deploymentSvc, userService,
		fetcher, healthChecker, repoWatcher, eventStore, deploymentStore,
	)

	return &serverDeps{
		businessHandler: bh,
		socketHandler:   socket.NewWebSocketHandler(logs.NewHistoryHub(0)),
		userService:     userService,
		oidcService:     users.NewOidcService(configStore, authStore),
		configStore:     configStore,
	}
}

// --- Tests ---

func TestNewServer(t *testing.T) {
	require.NotNil(t, NewServer())
}

func TestApplyMiddlewares_Order(t *testing.T) {
	var order []string

	h := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		order = append(order, "handler")
	})

	mw := func(name string) func(http.Handler) http.Handler {
		return func(next http.Handler) http.Handler {
			order = append(order, name+"-wrap")
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				order = append(order, name+"-call")
				next.ServeHTTP(w, r)
				order = append(order, name+"-after")
			})
		}
	}

	result := applyMiddlewares(h, mw("mw1"), mw("mw2"), mw("mw3"))
	result.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/test", nil))

	expected := []string{
		"mw3-wrap", "mw2-wrap", "mw1-wrap",
		"mw1-call", "mw2-call", "mw3-call",
		"handler",
		"mw3-after", "mw2-after", "mw1-after",
	}
	require.Len(t, order, len(expected))
	for i, exp := range expected {
		assert.Equal(t, exp, order[i], "position %d", i)
	}
}

func TestApplyMiddlewares_Empty(t *testing.T) {
	var called bool
	h := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		called = true
	})
	result := applyMiddlewares(h)
	result.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/test", nil))
	assert.True(t, called, "handler should be called when no middlewares")
}

func TestServe_SPARoute(t *testing.T) {
	frontDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(frontDir, "index.html"), []byte("<html>Hello SPA</html>"), 0644))

	deps := newServerDeps(t)
	srv := NewServer()
	go func() {
		_ = srv.Serve(models.ServerParams{Port: 18090, FrontDir: frontDir},
			deps.businessHandler, deps.socketHandler, deps.userService, deps.oidcService)
	}()
	t.Cleanup(func() { srv.Shutdown(context.Background()) })

	resp := waitForResponse(t, &http.Client{Timeout: 2 * time.Second},
		mustNewRequest("GET", "http://127.0.0.1:18090/app/dashboard", nil), 10*time.Second)
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, string(body), "Hello SPA")
}

func TestServe_OidcLoginRedirect(t *testing.T) {
	oidcServer := testutil.NewOidcTestServerWithToken(t)
	defer oidcServer.Server.Close()

	deps := newServerDeps(t)
	require.NoError(t, deps.configStore.Update(models.Config{
		Settings: models.Settings{
			Oidc: models.OidcConfig{
				IssuerURL: oidcServer.IssuerURL,
				ClientID:  testutil.ClientID,
			},
		},
	}))

	srv := NewServer()
	go func() {
		_ = srv.Serve(models.ServerParams{Port: 18091, FrontDir: t.TempDir()},
			deps.businessHandler, deps.socketHandler, deps.userService, deps.oidcService)
	}()
	t.Cleanup(func() { srv.Shutdown(context.Background()) })

	client := &http.Client{
		Timeout: 2 * time.Second,
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	resp := waitForResponse(t, client,
		mustNewRequest("GET", "http://127.0.0.1:18091/api/oidc/login", nil), 10*time.Second)

	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Contains(t, resp.Header.Get("Location"), oidcServer.IssuerURL)
	resp.Body.Close()
}

func TestServe_AuthRegisterGet(t *testing.T) {
	deps := newServerDeps(t)
	srv := NewServer()
	go func() {
		_ = srv.Serve(models.ServerParams{Port: 18092, FrontDir: t.TempDir()},
			deps.businessHandler, deps.socketHandler, deps.userService, deps.oidcService)
	}()
	t.Cleanup(func() { srv.Shutdown(context.Background()) })

	resp := waitForResponse(t, &http.Client{Timeout: 2 * time.Second},
		mustNewRequest("GET", "http://127.0.0.1:18092/api/auth/register", nil), 10*time.Second)
	resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestServe_AuthLoginPost(t *testing.T) {
	deps := newServerDeps(t)
	srv := NewServer()
	go func() {
		_ = srv.Serve(models.ServerParams{Port: 18093, FrontDir: t.TempDir()},
			deps.businessHandler, deps.socketHandler, deps.userService, deps.oidcService)
	}()
	t.Cleanup(func() { srv.Shutdown(context.Background()) })

	req, _ := http.NewRequest("POST", "http://127.0.0.1:18093/api/auth/login",
		strings.NewReader(`{"username":"test","password":"pass"}`))
	req.Header.Set("Content-Type", "application/json")
	resp := waitForResponse(t, &http.Client{Timeout: 2 * time.Second}, req, 10*time.Second)
	resp.Body.Close()

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestServe_CORSHeaders(t *testing.T) {
	deps := newServerDeps(t)
	srv := NewServer()
	go func() {
		_ = srv.Serve(models.ServerParams{Port: 18094, FrontDir: t.TempDir()},
			deps.businessHandler, deps.socketHandler, deps.userService, deps.oidcService)
	}()
	t.Cleanup(func() { srv.Shutdown(context.Background()) })

	req, _ := http.NewRequest("POST", "http://127.0.0.1:18094/api/auth/login",
		strings.NewReader(`{"username":"test","password":"pass"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", "127.0.0.1:18094")

	resp := waitForResponse(t, &http.Client{Timeout: 2 * time.Second}, req, 10*time.Second)
	resp.Body.Close()

	assert.NotEmpty(t, resp.Header.Get("Access-Control-Allow-Origin"),
		"expected CORS Access-Control-Allow-Origin header")
}

func TestServe_UnauthorizedApiRoute(t *testing.T) {
	deps := newServerDeps(t)
	srv := NewServer()
	go func() {
		_ = srv.Serve(models.ServerParams{Port: 18095, FrontDir: t.TempDir()},
			deps.businessHandler, deps.socketHandler, deps.userService, deps.oidcService)
	}()
	t.Cleanup(func() { srv.Shutdown(context.Background()) })

	resp := waitForResponse(t, &http.Client{Timeout: 2 * time.Second},
		mustNewRequest("GET", "http://127.0.0.1:18095/api/features", nil), 10*time.Second)
	resp.Body.Close()

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestShutdown_NoPanic(t *testing.T) {
	srv := NewServer().(*HTTPServer)
	assert.NotPanics(t, func() {
		srv.Shutdown(context.Background())
	})
}

// --- Helpers ---

func waitForResponse(t *testing.T, client *http.Client, req *http.Request, timeout time.Duration) *http.Response {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		resp, err := client.Do(req)
		if err == nil && resp != nil {
			return resp
		}
		time.Sleep(200 * time.Millisecond)
	}
	t.Fatal("server did not respond within timeout")
	return nil
}

func mustNewRequest(method, url string, body io.Reader) *http.Request {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		panic(err)
	}
	return req
}
