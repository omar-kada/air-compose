package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAuthorizationMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		url            string
		displayConfig  string
		editConfig     string
		editSettings   string
		expectedStatus int
	}{
		{
			name:           "GET config with all features enabled",
			method:         "GET",
			url:            "/api/config",
			displayConfig:  "true",
			editConfig:     "true",
			editSettings:   "true",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "POST config with all features enabled",
			method:         "POST",
			url:            "/api/config",
			displayConfig:  "true",
			editConfig:     "true",
			editSettings:   "true",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "GET config with display config disabled",
			method:         "GET",
			url:            "/api/config",
			displayConfig:  "false",
			editConfig:     "true",
			editSettings:   "true",
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "POST config with edit config disabled",
			method:         "POST",
			url:            "/api/config",
			displayConfig:  "true",
			editConfig:     "false",
			editSettings:   "true",
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "POST settings with edit settings enabled",
			method:         "POST",
			url:            "/api/settings",
			displayConfig:  "true",
			editConfig:     "true",
			editSettings:   "true",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "POST settings with edit settings disabled",
			method:         "POST",
			url:            "/api/settings",
			displayConfig:  "true",
			editConfig:     "true",
			editSettings:   "false",
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "POST route that is not feature based",
			method:         "POST",
			url:            "/api/another-route",
			displayConfig:  "false",
			editConfig:     "false",
			editSettings:   "false",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock handler to test the middleware
			mockHandler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			// Create a new request
			req, err := http.NewRequest(tt.method, tt.url, http.NoBody)
			if err != nil {
				t.Fatal(err)
			}

			// Create a response recorder
			rr := httptest.NewRecorder()

			// Set up environment variables for testing
			t.Setenv("AIR_COMPOSE_DISPLAY_CONFIG", tt.displayConfig)
			t.Setenv("AIR_COMPOSE_EDIT_CONFIG", tt.editConfig)
			t.Setenv("AIR_COMPOSE_EDIT_SETTINGS", tt.editSettings)

			// Call the middleware with the mock handler
			handler := AuthorizationMiddleware(mockHandler)
			handler.ServeHTTP(rr, req)

			// Check the status code
			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, tt.expectedStatus)
			}
		})
	}
}
