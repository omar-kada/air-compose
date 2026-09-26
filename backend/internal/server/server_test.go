package server

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"omar-kada/air-compose/api"
	"omar-kada/air-compose/internal/models"
	"omar-kada/air-compose/internal/server/socket"
)

// --- Stubs ---

type stubHandler struct {
	registeredCalled bool
}

func (s *stubHandler) AuthAPILogin(_ context.Context, _ api.AuthAPILoginRequestObject) (api.AuthAPILoginResponseObject, error) {
	return api.AuthAPILogin200JSONResponse{}, nil
}
func (s *stubHandler) AuthAPILogout(_ context.Context, _ api.AuthAPILogoutRequestObject) (api.AuthAPILogoutResponseObject, error) {
	return api.AuthAPILogout200JSONResponse{}, nil
}
func (s *stubHandler) AuthAPIRefresh(_ context.Context, _ api.AuthAPIRefreshRequestObject) (api.AuthAPIRefreshResponseObject, error) {
	return api.AuthAPIRefresh200JSONResponse{}, nil
}
func (s *stubHandler) AuthAPIRegistered(_ context.Context, _ api.AuthAPIRegisteredRequestObject) (api.AuthAPIRegisteredResponseObject, error) {
	s.registeredCalled = true
	return api.AuthAPIRegistered200JSONResponse{}, nil
}
func (s *stubHandler) AuthAPIRegister(_ context.Context, _ api.AuthAPIRegisterRequestObject) (api.AuthAPIRegisterResponseObject, error) {
	return api.AuthAPIRegister200JSONResponse{}, nil
}
func (s *stubHandler) ConfigAPIGet(_ context.Context, _ api.ConfigAPIGetRequestObject) (api.ConfigAPIGetResponseObject, error) {
	return api.ConfigAPIGet200JSONResponse{}, nil
}
func (s *stubHandler) ConfigAPISet(_ context.Context, _ api.ConfigAPISetRequestObject) (api.ConfigAPISetResponseObject, error) {
	return api.ConfigAPISet200JSONResponse{}, nil
}
func (s *stubHandler) DeployementAPIList(_ context.Context, _ api.DeployementAPIListRequestObject) (api.DeployementAPIListResponseObject, error) {
	return api.DeployementAPIList200JSONResponse{}, nil
}
func (s *stubHandler) DeployementAPISync(_ context.Context, _ api.DeployementAPISyncRequestObject) (api.DeployementAPISyncResponseObject, error) {
	return api.DeployementAPISync200JSONResponse{}, nil
}
func (s *stubHandler) DeployementAPIRead(_ context.Context, _ api.DeployementAPIReadRequestObject) (api.DeployementAPIReadResponseObject, error) {
	return api.DeployementAPIRead200JSONResponse{}, nil
}
func (s *stubHandler) DiffAPIGet(_ context.Context, _ api.DiffAPIGetRequestObject) (api.DiffAPIGetResponseObject, error) {
	return api.DiffAPIGet200JSONResponse{}, nil
}
func (s *stubHandler) FeaturesAPIGet(_ context.Context, _ api.FeaturesAPIGetRequestObject) (api.FeaturesAPIGetResponseObject, error) {
	return api.FeaturesAPIGet200JSONResponse{}, nil
}
func (s *stubHandler) NotificationsAPIList(_ context.Context, _ api.NotificationsAPIListRequestObject) (api.NotificationsAPIListResponseObject, error) {
	return api.NotificationsAPIList200JSONResponse{}, nil
}
func (s *stubHandler) OIDCAPIOidcCallback(_ context.Context, _ api.OIDCAPIOidcCallbackRequestObject) (api.OIDCAPIOidcCallbackResponseObject, error) {
	return api.OIDCAPIOidcCallbackdefaultJSONResponse{}, nil
}
func (s *stubHandler) OIDCAPIOidcLogin(_ context.Context, _ api.OIDCAPIOidcLoginRequestObject) (api.OIDCAPIOidcLoginResponseObject, error) {
	return api.OIDCAPIOidcLogindefaultJSONResponse{}, nil
}
func (s *stubHandler) SettingsAPIGet(_ context.Context, _ api.SettingsAPIGetRequestObject) (api.SettingsAPIGetResponseObject, error) {
	return api.SettingsAPIGet200JSONResponse{}, nil
}
func (s *stubHandler) SettingsAPISet(_ context.Context, _ api.SettingsAPISetRequestObject) (api.SettingsAPISetResponseObject, error) {
	return api.SettingsAPISet200JSONResponse{}, nil
}
func (s *stubHandler) SettingsAPITestGitConnection(_ context.Context, _ api.SettingsAPITestGitConnectionRequestObject) (api.SettingsAPITestGitConnectionResponseObject, error) {
	return api.SettingsAPITestGitConnection200JSONResponse{}, nil
}
func (s *stubHandler) StateAPIGet(_ context.Context, _ api.StateAPIGetRequestObject) (api.StateAPIGetResponseObject, error) {
	return api.StateAPIGet200JSONResponse{}, nil
}
func (s *stubHandler) StatusAPIGet(_ context.Context, _ api.StatusAPIGetRequestObject) (api.StatusAPIGetResponseObject, error) {
	return api.StatusAPIGet200JSONResponse{}, nil
}
func (s *stubHandler) UserAPIDelete(_ context.Context, _ api.UserAPIDeleteRequestObject) (api.UserAPIDeleteResponseObject, error) {
	return api.UserAPIDelete200JSONResponse{}, nil
}
func (s *stubHandler) UserAPIGet(_ context.Context, _ api.UserAPIGetRequestObject) (api.UserAPIGetResponseObject, error) {
	return api.UserAPIGet200JSONResponse{}, nil
}
func (s *stubHandler) UserAPIChangePassword(_ context.Context, _ api.UserAPIChangePasswordRequestObject) (api.UserAPIChangePasswordResponseObject, error) {
	return api.UserAPIChangePassword200JSONResponse{}, nil
}
func (s *stubHandler) WebSocketConnect(_ context.Context, _ api.WebSocketConnectRequestObject) (api.WebSocketConnectResponseObject, error) {
	return api.WebSocketConnect401Response{}, nil
}

type stubSocketHandler struct{}

func (s *stubSocketHandler) Handle(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}
func (s *stubSocketHandler) BroadcastEvent(_ context.Context, _ models.Event) {}

type stubUserService struct {
	loginCalled bool
	loginInput  models.Credentials
	loginError  error
}

func (s *stubUserService) Login(creds models.Credentials) (models.Token, error) {
	s.loginCalled = true
	s.loginInput = creds
	return models.Token{}, s.loginError
}
func (s *stubUserService) Register(_ models.Credentials) (models.Token, error) {
	return models.Token{}, nil
}
func (s *stubUserService) Logout(_ models.Token) error                       { return nil }
func (s *stubUserService) GetUsernameByToken(_ models.Token) (string, error) { return "", nil }
func (s *stubUserService) RefreshToken(_ models.Token) (models.Token, error) {
	return models.Token{}, nil
}
func (s *stubUserService) IsRegistered() (bool, error)                 { return false, nil }
func (s *stubUserService) GetUser(_ string) (models.User, error)       { return models.User{}, nil }
func (s *stubUserService) DeleteUser(_ string) (bool, error)           { return false, nil }
func (s *stubUserService) ChangePassword(_, _, _ string) (bool, error) { return false, nil }

type stubOidcService struct {
	authURL string
}

func (s *stubOidcService) GetAuthURL(_, _, _ string) (string, error) {
	return s.authURL, nil
}
func (s *stubOidcService) LoginOidc(_, _, _ string) (models.Token, error) {
	return models.Token{}, nil
}

var (
	_ api.StrictServerInterface = (*stubHandler)(nil)
	_ socket.WebSocketHandler   = (*stubSocketHandler)(nil)
)

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
	err := os.WriteFile(filepath.Join(frontDir, "index.html"), []byte("<html>Hello SPA</html>"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	srv := NewServer()
	params := models.ServerParams{Port: 18090, FrontDir: frontDir}
	go func() {
		_ = srv.Serve(params, &stubHandler{}, &stubSocketHandler{}, &stubUserService{}, &stubOidcService{})
	}()

	resp := waitForResponse(t, &http.Client{Timeout: 5 * time.Second}, mustNewRequest("GET", "http://127.0.0.1:18090/app/dashboard", nil), 5*time.Second)
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
	srv := NewServer()
	oidc := &stubOidcService{authURL: "https://oidc.example.com/auth"}
	go func() {
		_ = srv.Serve(models.ServerParams{Port: 18091, FrontDir: t.TempDir()}, &stubHandler{}, &stubSocketHandler{}, &stubUserService{}, oidc)
	}()

	client := &http.Client{
		Timeout: 5 * time.Second,
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	resp := waitForResponse(t, client, mustNewRequest("GET", "http://127.0.0.1:18091/api/oidc/login", nil), 5*time.Second)

	if resp.StatusCode != http.StatusFound {
		t.Fatalf("expected 302, got %d", resp.StatusCode)
	}
	if loc := resp.Header.Get("Location"); !strings.Contains(loc, "oidc.example.com") {
		t.Fatalf("expected OIDC redirect, got location: %s", loc)
	}
	resp.Body.Close()
	srv.Shutdown(context.Background())
}

func TestServe_AuthRegisterGet(t *testing.T) {
	srv := NewServer()
	h := &stubHandler{}
	go func() {
		_ = srv.Serve(models.ServerParams{Port: 18092, FrontDir: t.TempDir()}, h, &stubSocketHandler{}, &stubUserService{}, &stubOidcService{})
	}()

	resp := waitForResponse(t, &http.Client{Timeout: 5 * time.Second}, mustNewRequest("GET", "http://127.0.0.1:18092/api/auth/register", nil), 5*time.Second)
	resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if !h.registeredCalled {
		t.Fatal("AuthAPIRegistered handler was not called")
	}
	srv.Shutdown(context.Background())
}

func TestServe_AuthLoginPost(t *testing.T) {
	srv := NewServer()
	us := &stubUserService{loginError: errors.New("login failed")}
	go func() {
		_ = srv.Serve(models.ServerParams{Port: 18093, FrontDir: t.TempDir()}, &stubHandler{}, &stubSocketHandler{}, us, &stubOidcService{})
	}()

	client := &http.Client{Timeout: 5 * time.Second}
	req, _ := http.NewRequest("POST", "http://127.0.0.1:18093/api/auth/login",
		strings.NewReader(`{"username":"test","password":"pass"}`))
	req.Header.Set("Content-Type", "application/json")
	resp := waitForResponse(t, client, req, 5*time.Second)
	resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
	if !us.loginCalled {
		t.Fatal("Login was not called")
	}
	if us.loginInput.Username != "test" || us.loginInput.Password != "pass" {
		t.Fatalf("unexpected credentials: %+v", us.loginInput)
	}
	srv.Shutdown(context.Background())
}

func TestServe_CORSHeaders(t *testing.T) {
	srv := NewServer()
	us := &stubUserService{loginError: errors.New("login failed")}
	go func() {
		_ = srv.Serve(models.ServerParams{Port: 18094, FrontDir: t.TempDir()}, &stubHandler{}, &stubSocketHandler{}, us, &stubOidcService{})
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
	srv := NewServer()
	go func() {
		_ = srv.Serve(models.ServerParams{Port: 18095, FrontDir: t.TempDir()}, &stubHandler{}, &stubSocketHandler{}, &stubUserService{}, &stubOidcService{})
	}()

	resp := waitForResponse(t, &http.Client{Timeout: 5 * time.Second}, mustNewRequest("GET", "http://127.0.0.1:18095/api/features", nil), 5*time.Second)
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
