// Package server provides implementations of http and ws handlers.
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
	"omar-kada/air-compose/internal/models"
	"omar-kada/air-compose/internal/users"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// --- Mock implementations ---

// MockStrictServer implements api.StrictServerInterface for testing.
// Only methods that are actually called need expectations set.
type MockStrictServer struct {
	mock.Mock
}

func (m *MockStrictServer) AuthAPILogin(_ context.Context, _ api.AuthAPILoginRequestObject) (api.AuthAPILoginResponseObject, error) {
	args := m.Called()
	return args.Get(0).(api.AuthAPILoginResponseObject), args.Error(1)
}
func (m *MockStrictServer) AuthAPILogout(_ context.Context, _ api.AuthAPILogoutRequestObject) (api.AuthAPILogoutResponseObject, error) {
	args := m.Called()
	return args.Get(0).(api.AuthAPILogoutResponseObject), args.Error(1)
}
func (m *MockStrictServer) AuthAPIRefresh(_ context.Context, _ api.AuthAPIRefreshRequestObject) (api.AuthAPIRefreshResponseObject, error) {
	args := m.Called()
	return args.Get(0).(api.AuthAPIRefreshResponseObject), args.Error(1)
}
func (m *MockStrictServer) AuthAPIRegistered(_ context.Context, _ api.AuthAPIRegisteredRequestObject) (api.AuthAPIRegisteredResponseObject, error) {
	args := m.Called()
	return args.Get(0).(api.AuthAPIRegisteredResponseObject), args.Error(1)
}
func (m *MockStrictServer) AuthAPIRegister(_ context.Context, _ api.AuthAPIRegisterRequestObject) (api.AuthAPIRegisterResponseObject, error) {
	args := m.Called()
	return args.Get(0).(api.AuthAPIRegisterResponseObject), args.Error(1)
}
func (m *MockStrictServer) ConfigAPIGet(_ context.Context, _ api.ConfigAPIGetRequestObject) (api.ConfigAPIGetResponseObject, error) {
	args := m.Called()
	return args.Get(0).(api.ConfigAPIGetResponseObject), args.Error(1)
}
func (m *MockStrictServer) ConfigAPISet(_ context.Context, _ api.ConfigAPISetRequestObject) (api.ConfigAPISetResponseObject, error) {
	args := m.Called()
	return args.Get(0).(api.ConfigAPISetResponseObject), args.Error(1)
}
func (m *MockStrictServer) DeployementAPIList(_ context.Context, _ api.DeployementAPIListRequestObject) (api.DeployementAPIListResponseObject, error) {
	args := m.Called()
	return args.Get(0).(api.DeployementAPIListResponseObject), args.Error(1)
}
func (m *MockStrictServer) DeployementAPISync(_ context.Context, _ api.DeployementAPISyncRequestObject) (api.DeployementAPISyncResponseObject, error) {
	args := m.Called()
	return args.Get(0).(api.DeployementAPISyncResponseObject), args.Error(1)
}
func (m *MockStrictServer) DeployementAPIRead(_ context.Context, _ api.DeployementAPIReadRequestObject) (api.DeployementAPIReadResponseObject, error) {
	args := m.Called()
	return args.Get(0).(api.DeployementAPIReadResponseObject), args.Error(1)
}
func (m *MockStrictServer) DiffAPIGet(_ context.Context, _ api.DiffAPIGetRequestObject) (api.DiffAPIGetResponseObject, error) {
	args := m.Called()
	return args.Get(0).(api.DiffAPIGetResponseObject), args.Error(1)
}
func (m *MockStrictServer) FeaturesAPIGet(_ context.Context, _ api.FeaturesAPIGetRequestObject) (api.FeaturesAPIGetResponseObject, error) {
	args := m.Called()
	return args.Get(0).(api.FeaturesAPIGetResponseObject), args.Error(1)
}
func (m *MockStrictServer) NotificationsAPIList(_ context.Context, _ api.NotificationsAPIListRequestObject) (api.NotificationsAPIListResponseObject, error) {
	args := m.Called()
	return args.Get(0).(api.NotificationsAPIListResponseObject), args.Error(1)
}
func (m *MockStrictServer) OIDCAPIOidcCallback(_ context.Context, _ api.OIDCAPIOidcCallbackRequestObject) (api.OIDCAPIOidcCallbackResponseObject, error) {
	args := m.Called()
	return args.Get(0).(api.OIDCAPIOidcCallbackResponseObject), args.Error(1)
}
func (m *MockStrictServer) OIDCAPIOidcLogin(_ context.Context, _ api.OIDCAPIOidcLoginRequestObject) (api.OIDCAPIOidcLoginResponseObject, error) {
	args := m.Called()
	return args.Get(0).(api.OIDCAPIOidcLoginResponseObject), args.Error(1)
}
func (m *MockStrictServer) SettingsAPIGet(_ context.Context, _ api.SettingsAPIGetRequestObject) (api.SettingsAPIGetResponseObject, error) {
	args := m.Called()
	return args.Get(0).(api.SettingsAPIGetResponseObject), args.Error(1)
}
func (m *MockStrictServer) SettingsAPISet(_ context.Context, _ api.SettingsAPISetRequestObject) (api.SettingsAPISetResponseObject, error) {
	args := m.Called()
	return args.Get(0).(api.SettingsAPISetResponseObject), args.Error(1)
}
func (m *MockStrictServer) SettingsAPITestGitConnection(_ context.Context, _ api.SettingsAPITestGitConnectionRequestObject) (api.SettingsAPITestGitConnectionResponseObject, error) {
	args := m.Called()
	return args.Get(0).(api.SettingsAPITestGitConnectionResponseObject), args.Error(1)
}
func (m *MockStrictServer) StateAPIGet(_ context.Context, _ api.StateAPIGetRequestObject) (api.StateAPIGetResponseObject, error) {
	args := m.Called()
	return args.Get(0).(api.StateAPIGetResponseObject), args.Error(1)
}
func (m *MockStrictServer) StatusAPIGet(_ context.Context, _ api.StatusAPIGetRequestObject) (api.StatusAPIGetResponseObject, error) {
	args := m.Called()
	return args.Get(0).(api.StatusAPIGetResponseObject), args.Error(1)
}
func (m *MockStrictServer) UserAPIDelete(_ context.Context, _ api.UserAPIDeleteRequestObject) (api.UserAPIDeleteResponseObject, error) {
	args := m.Called()
	return args.Get(0).(api.UserAPIDeleteResponseObject), args.Error(1)
}
func (m *MockStrictServer) UserAPIGet(_ context.Context, _ api.UserAPIGetRequestObject) (api.UserAPIGetResponseObject, error) {
	args := m.Called()
	return args.Get(0).(api.UserAPIGetResponseObject), args.Error(1)
}
func (m *MockStrictServer) UserAPIChangePassword(_ context.Context, _ api.UserAPIChangePasswordRequestObject) (api.UserAPIChangePasswordResponseObject, error) {
	args := m.Called()
	return args.Get(0).(api.UserAPIChangePasswordResponseObject), args.Error(1)
}
func (m *MockStrictServer) WebSocketConnect(_ context.Context, _ api.WebSocketConnectRequestObject) (api.WebSocketConnectResponseObject, error) {
	args := m.Called()
	return args.Get(0).(api.WebSocketConnectResponseObject), args.Error(1)
}

// MockWebSocketHandler implements socket.WebSocketHandler for testing.
type MockWebSocketHandler struct {
	mock.Mock
}

func (m *MockWebSocketHandler) Handle(w http.ResponseWriter, r *http.Request) {
	m.Called(w, r)
}
func (m *MockWebSocketHandler) BroadcastEvent(ctx context.Context, event models.Event) {
	m.Called(ctx, event)
}

// MockUserService implements users.Service for testing.
type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) Login(creds models.Credentials) (models.Token, error) {
	args := m.Called(creds)
	return args.Get(0).(models.Token), args.Error(1)
}
func (m *MockUserService) Register(creds models.Credentials) (models.Token, error) {
	args := m.Called(creds)
	return args.Get(0).(models.Token), args.Error(1)
}
func (m *MockUserService) Logout(token models.Token) error {
	args := m.Called(token)
	return args.Error(0)
}
func (m *MockUserService) GetUsernameByToken(token models.Token) (string, error) {
	args := m.Called(token)
	return args.Get(0).(string), args.Error(1)
}
func (m *MockUserService) RefreshToken(token models.Token) (models.Token, error) {
	args := m.Called(token)
	return args.Get(0).(models.Token), args.Error(1)
}
func (m *MockUserService) IsRegistered() (bool, error) {
	args := m.Called()
	return args.Bool(0), args.Error(1)
}
func (m *MockUserService) GetUser(_ string) (models.User, error) {
	args := m.Called()
	return args.Get(0).(models.User), args.Error(1)
}
func (m *MockUserService) DeleteUser(_ string) (bool, error) {
	args := m.Called()
	return args.Bool(0), args.Error(1)
}
func (m *MockUserService) ChangePassword(_, _, _ string) (bool, error) {
	args := m.Called()
	return args.Bool(0), args.Error(1)
}

// MockOidcService implements users.OidcService for testing.
type MockOidcService struct {
	mock.Mock
}

func (m *MockOidcService) GetAuthURL(clientID, redirectURL, state string) (string, error) {
	args := m.Called(clientID, redirectURL, state)
	return args.Get(0).(string), args.Error(1)
}
func (m *MockOidcService) LoginOidc(code, state, codeVerifier string) (models.Token, error) {
	args := m.Called(code, state, codeVerifier)
	return args.Get(0).(models.Token), args.Error(1)
}

// --- Tests ---

func TestNewServer(t *testing.T) {
	srv := NewServer()
	assert.NotNil(t, srv)
	_, ok := srv.(*HTTPServer)
	assert.True(t, ok, "NewServer should return *HTTPServer")
}

func TestApplyMiddlewares_Order(t *testing.T) {
	// applyMiddlewares iterates middlewares in reverse order:
	// mws = [mw1, mw2, mw3] → mw3 wraps first, then mw2, then mw1
	// Execution order: mw1-call → mw2-call → mw3-call → handler → mw3-after → mw2-after → mw1-after
	var callOrder []string

	finalHandler := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		callOrder = append(callOrder, "handler")
	})

	mw := func(name string) func(http.Handler) http.Handler {
		return func(next http.Handler) http.Handler {
			callOrder = append(callOrder, name+"-wrap")
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				callOrder = append(callOrder, name+"-call")
				next.ServeHTTP(w, r)
				callOrder = append(callOrder, name+"-after")
			})
		}
	}

	// Wrapping order: mw3, mw2, mw1 (reverse)
	result := applyMiddlewares(finalHandler, mw("mw1"), mw("mw2"), mw("mw3"))

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	result.ServeHTTP(w, req)

	expected := []string{
		"mw3-wrap", "mw2-wrap", "mw1-wrap",
		"mw1-call", "mw2-call", "mw3-call",
		"handler",
		"mw3-after", "mw2-after", "mw1-after",
	}
	assert.Equal(t, expected, callOrder)
}

func TestApplyMiddlewares_Empty(t *testing.T) {
	var called bool
	handler := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		called = true
	})

	result := applyMiddlewares(handler)
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	result.ServeHTTP(w, req)
	assert.True(t, called, "handler should be called when no middlewares")
}

func TestServe_SPARoute(t *testing.T) {
	// Create a temp front dir with index.html
	frontDir := t.TempDir()
	indexHTML := "<html><body>Hello SPA</body></html>"
	err := os.WriteFile(filepath.Join(frontDir, "index.html"), []byte(indexHTML), 0644)
	assert.NoError(t, err)

	businessHandler := new(MockStrictServer)
	wsHandler := new(MockWebSocketHandler)
	userSvc := new(MockUserService)
	oidcSvc := new(MockOidcService)

	srv := NewServer()
	params := models.ServerParams{
		Port:     18090,
		FrontDir: frontDir,
	}

	go func() {
		_ = srv.Serve(params, businessHandler, wsHandler, userSvc, oidcSvc)
	}()

	client := &http.Client{Timeout: 5 * time.Second}
	var resp *http.Response
	for i := 0; i < 10; i++ {
		resp, _ = client.Get("http://127.0.0.1:18090/app/dashboard")
		if resp != nil {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	assert.NotNil(t, resp, "server should respond")
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	body, _ := io.ReadAll(resp.Body)
	assert.Contains(t, string(body), "Hello SPA")
	resp.Body.Close()

	srv.Shutdown(context.Background())
}

func TestServe_OidcLoginRedirect(t *testing.T) {
	businessHandler := new(MockStrictServer)
	wsHandler := new(MockWebSocketHandler)
	userSvc := new(MockUserService)
	oidcSvc := new(MockOidcService)

	oidcSvc.On("GetAuthURL", mock.Anything, mock.Anything, mock.Anything).Return("https://oidc.example.com/auth", nil).Maybe()

	srv := NewServer()
	params := models.ServerParams{
		Port:     18091,
		FrontDir: t.TempDir(),
	}

	go func() {
		_ = srv.Serve(params, businessHandler, wsHandler, userSvc, oidcSvc)
	}()

	client := &http.Client{Timeout: 5 * time.Second, CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
		return http.ErrUseLastResponse
	}}
	var resp *http.Response
	for i := 0; i < 10; i++ {
		resp, _ = client.Get("http://127.0.0.1:18091/api/oidc/login")
		if resp != nil {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	assert.NotNil(t, resp)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	location := resp.Header.Get("Location")
	assert.Contains(t, location, "oidc.example.com")
	resp.Body.Close()

	oidcSvc.AssertCalled(t, "GetAuthURL", mock.Anything, mock.Anything, mock.Anything)
	oidcSvc.AssertExpectations(t)

	srv.Shutdown(context.Background())
}

func TestServe_AuthRegisterGet(t *testing.T) {
	// GET /api/auth/register bypasses auth middleware and goes to the handler
	businessHandler := new(MockStrictServer)
	wsHandler := new(MockWebSocketHandler)
	userSvc := new(MockUserService)
	oidcSvc := new(MockOidcService)

	businessHandler.On("AuthAPIRegistered", mock.Anything, mock.Anything).Return(
		api.AuthAPIRegistered200JSONResponse{Registered: false, Oidc: false}, nil)

	srv := NewServer()
	params := models.ServerParams{
		Port:     18092,
		FrontDir: t.TempDir(),
	}

	go func() {
		_ = srv.Serve(params, businessHandler, wsHandler, userSvc, oidcSvc)
	}()

	client := &http.Client{Timeout: 5 * time.Second}
	var resp *http.Response
	for i := 0; i < 10; i++ {
		resp, _ = client.Get("http://127.0.0.1:18092/api/auth/register")
		if resp != nil {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	assert.NotNil(t, resp)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()

	businessHandler.AssertCalled(t, "AuthAPIRegistered", mock.Anything, mock.Anything)
	businessHandler.AssertExpectations(t)

	srv.Shutdown(context.Background())
}

func TestServe_AuthLoginPost(t *testing.T) {
	// POST /api/auth/login is handled by auth middleware (loginHandler), not the strict handler
	businessHandler := new(MockStrictServer)
	wsHandler := new(MockWebSocketHandler)
	userSvc := new(MockUserService)
	oidcSvc := new(MockOidcService)

	// Send valid credentials so loginHandler calls authService.Login
	loginCreds := models.Credentials{Username: "testuser", Password: "testpass"}
	userSvc.On("Login", loginCreds).Return(models.Token{}, assert.AnError)

	srv := NewServer()
	params := models.ServerParams{
		Port:     18093,
		FrontDir: t.TempDir(),
	}

	go func() {
		_ = srv.Serve(params, businessHandler, wsHandler, userSvc, oidcSvc)
	}()

	client := &http.Client{Timeout: 5 * time.Second}
	var resp *http.Response
	for i := 0; i < 10; i++ {
		resp, _ = client.Post("http://127.0.0.1:18093/api/auth/login", "application/json", strings.NewReader(`{"username":"testuser","password":"testpass"}`))
		if resp != nil {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	assert.NotNil(t, resp)
	// loginHandler calls authService.Login which returns error → sends 401
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	resp.Body.Close()

	userSvc.AssertCalled(t, "Login", loginCreds)
	userSvc.AssertExpectations(t)

	srv.Shutdown(context.Background())
}

func TestServe_CORSHeaders(t *testing.T) {
	// POST /api/auth/login goes through CORS middleware and auth middleware
	businessHandler := new(MockStrictServer)
	wsHandler := new(MockWebSocketHandler)
	userSvc := new(MockUserService)
	oidcSvc := new(MockOidcService)

	loginCreds := models.Credentials{Username: "testuser", Password: "testpass"}
	userSvc.On("Login", loginCreds).Return(models.Token{}, assert.AnError)

	srv := NewServer()
	params := models.ServerParams{
		Port:     18094,
		FrontDir: t.TempDir(),
	}

	go func() {
		_ = srv.Serve(params, businessHandler, wsHandler, userSvc, oidcSvc)
	}()

	client := &http.Client{Timeout: 5 * time.Second}
	var resp *http.Response
	for i := 0; i < 10; i++ {
		req, _ := http.NewRequest("POST", "http://127.0.0.1:18094/api/auth/login", strings.NewReader(`{"username":"testuser","password":"testpass"}`))
		req.Header.Set("Content-Type", "application/json")
		// CORS middleware pattern "127.0.0.1:*" matches origins without protocol prefix
		req.Header.Set("Origin", "127.0.0.1:18094")
		resp, _ = client.Do(req)
		if resp != nil {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	assert.NotNil(t, resp)
	// CORS headers should be present when Origin header matches allowed origins
	origin := resp.Header.Get("Access-Control-Allow-Origin")
	assert.NotEmpty(t, origin, "CORS origin header should be set for allowed origin")
	resp.Body.Close()

	srv.Shutdown(context.Background())
}

func TestServe_UnauthorizedApiRoute(t *testing.T) {
	// /api/ routes that aren't auth/ or oidc/ require a valid token
	businessHandler := new(MockStrictServer)
	wsHandler := new(MockWebSocketHandler)
	userSvc := new(MockUserService)
	oidcSvc := new(MockOidcService)

	srv := NewServer()
	params := models.ServerParams{
		Port:     18095,
		FrontDir: t.TempDir(),
	}

	go func() {
		_ = srv.Serve(params, businessHandler, wsHandler, userSvc, oidcSvc)
	}()

	client := &http.Client{Timeout: 5 * time.Second}
	var resp *http.Response
	for i := 0; i < 10; i++ {
		resp, _ = client.Get("http://127.0.0.1:18095/api/features")
		if resp != nil {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	assert.NotNil(t, resp)
	// Without auth cookies, should get 401
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	resp.Body.Close()

	srv.Shutdown(context.Background())
}

func TestShutdown_NoPanic(t *testing.T) {
	srv := NewServer().(*HTTPServer)
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	// Shutdown on an uninitialized server should not panic
	assert.NotPanics(t, func() {
		srv.Shutdown(ctx)
	})
}

// Ensure mock types satisfy their interfaces at compile time
var _ api.StrictServerInterface = (*MockStrictServer)(nil)
var _ interface {
	Handle(http.ResponseWriter, *http.Request)
	BroadcastEvent(context.Context, models.Event)
} = (*MockWebSocketHandler)(nil)
var _ users.Service = (*MockUserService)(nil)
var _ users.OidcService = (*MockOidcService)(nil)
