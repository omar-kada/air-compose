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
	if err != nil {
		t.Fatalf("config store: %v", err)
	}

	eventStore, err := events.NewEventStorage(db)
	if err != nil {
		t.Fatalf("event store: %v", err)
	}
	deploymentStore, err := deployments.NewDeploymentStorage(db)
	if err != nil {
		t.Fatalf("deployment store: %v", err)
	}
	userStore, err := users.NewUsersStorage(db)
	if err != nil {
		t.Fatalf("user store: %v", err)
	}
	sessionStore, err := users.NewSessionStorage(db)
	if err != nil {
		t.Fatalf("session store: %v", err)
	}
	authStore, err := users.NewAuthStorage(userStore, sessionStore, users.NewTokenHolder())
	if err != nil {
		t.Fatalf("auth store: %v", err)
	}

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
	if NewServer() == nil {
		t.Fatal("NewServer returned nil")
	}
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
	if len(order) != len(expected) {
		t.Fatalf("expected %d entries, got %d: %v", len(expected), len(order), order)
	}
	for i, exp := range expected {
		if order[i] != exp {
			t.Fatalf("position %d: expected %q, got %q", i, exp, order[i])
		}
	}
}

func TestApplyMiddlewares_Empty(t *testing.T) {
	var called bool
	h := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		called = true
	})
	result := applyMiddlewares(h)
	result.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/test", nil))
	if !called {
		t.Fatal("handler should be called when no middlewares")
	}
}

func TestServe_SPARoute(t *testing.T) {
	frontDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(frontDir, "index.html"), []byte("<html>Hello SPA</html>"), 0644); err != nil {
		t.Fatal(err)
	}

	deps := newServerDeps(t)
	srv := NewServer()
	go func() {
		_ = srv.Serve(
			models.ServerParams{Port: 18090, FrontDir: frontDir},
			deps.businessHandler, deps.socketHandler, deps.userService, deps.oidcService,
		)
	}()

	resp := waitForResponse(t, &http.Client{Timeout: 5 * time.Second},
		mustNewRequest("GET", "http://127.0.0.1:18090/app/dashboard", nil), 5*time.Second)
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if !strings.Contains(string(body), "Hello SPA") {
		t.Fatalf("expected SPA content, got %s", string(body))
	}
	srv.Shutdown(context.Background())
}

func TestServe_OidcLoginRedirect(t *testing.T) {
	oidcServer := testutil.NewOidcTestServerWithToken(t)
	defer oidcServer.Server.Close()

	deps := newServerDeps(t)
	deps.configStore.Update(models.Config{
		Settings: models.Settings{
			Oidc: models.OidcConfig{
				IssuerURL: oidcServer.IssuerURL,
				ClientID:  testutil.ClientID,
			},
		},
	})

	srv := NewServer()
	go func() {
		_ = srv.Serve(
			models.ServerParams{Port: 18091, FrontDir: t.TempDir()},
			deps.businessHandler, deps.socketHandler, deps.userService, deps.oidcService,
		)
	}()

	client := &http.Client{
		Timeout: 5 * time.Second,
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	resp := waitForResponse(t, client,
		mustNewRequest("GET", "http://127.0.0.1:18091/api/oidc/login", nil), 5*time.Second)

	if resp.StatusCode != http.StatusFound {
		t.Fatalf("expected 302, got %d", resp.StatusCode)
	}
	if loc := resp.Header.Get("Location"); !strings.Contains(loc, oidcServer.IssuerURL) {
		t.Fatalf("expected OIDC redirect to %q, got location: %s", oidcServer.IssuerURL, loc)
	}
	resp.Body.Close()
	srv.Shutdown(context.Background())
}

func TestServe_AuthRegisterGet(t *testing.T) {
	deps := newServerDeps(t)
	srv := NewServer()
	go func() {
		_ = srv.Serve(
			models.ServerParams{Port: 18092, FrontDir: t.TempDir()},
			deps.businessHandler, deps.socketHandler, deps.userService, deps.oidcService,
		)
	}()

	resp := waitForResponse(t, &http.Client{Timeout: 5 * time.Second},
		mustNewRequest("GET", "http://127.0.0.1:18092/api/auth/register", nil), 5*time.Second)
	resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestServe_AuthLoginPost(t *testing.T) {
	deps := newServerDeps(t)
	srv := NewServer()
	go func() {
		_ = srv.Serve(
			models.ServerParams{Port: 18093, FrontDir: t.TempDir()},
			deps.businessHandler, deps.socketHandler, deps.userService, deps.oidcService,
		)
	}()

	req, _ := http.NewRequest("POST", "http://127.0.0.1:18093/api/auth/login",
		strings.NewReader(`{"username":"test","password":"pass"}`))
	req.Header.Set("Content-Type", "application/json")
	resp := waitForResponse(t, &http.Client{Timeout: 5 * time.Second}, req, 5*time.Second)
	resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
}

func TestServe_CORSHeaders(t *testing.T) {
	deps := newServerDeps(t)
	srv := NewServer()
	go func() {
		_ = srv.Serve(
			models.ServerParams{Port: 18094, FrontDir: t.TempDir()},
			deps.businessHandler, deps.socketHandler, deps.userService, deps.oidcService,
		)
	}()

	req, _ := http.NewRequest("POST", "http://127.0.0.1:18094/api/auth/login",
		strings.NewReader(`{"username":"test","password":"pass"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", "127.0.0.1:18094")

	resp := waitForResponse(t, &http.Client{Timeout: 5 * time.Second}, req, 5*time.Second)
	resp.Body.Close()

	if origin := resp.Header.Get("Access-Control-Allow-Origin"); origin == "" {
		t.Fatal("expected CORS Access-Control-Allow-Origin header")
	}
	srv.Shutdown(context.Background())
}

func TestServe_UnauthorizedApiRoute(t *testing.T) {
	deps := newServerDeps(t)
	srv := NewServer()
	go func() {
		_ = srv.Serve(
			models.ServerParams{Port: 18095, FrontDir: t.TempDir()},
			deps.businessHandler, deps.socketHandler, deps.userService, deps.oidcService,
		)
	}()

	resp := waitForResponse(t, &http.Client{Timeout: 5 * time.Second},
		mustNewRequest("GET", "http://127.0.0.1:18095/api/features", nil), 5*time.Second)
	resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
	srv.Shutdown(context.Background())
}

func TestShutdown_NoPanic(t *testing.T) {
	srv := NewServer().(*HTTPServer)
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Shutdown panicked: %v", r)
		}
	}()
	srv.Shutdown(context.Background())
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
		time.Sleep(50 * time.Millisecond)
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
