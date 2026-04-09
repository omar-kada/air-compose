// Package server provides implementations of http and ws handlers
package server

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"omar-kada/air-compose/api"
	"omar-kada/air-compose/internal/process"
	"omar-kada/air-compose/internal/server/middlewares"
	"omar-kada/air-compose/internal/storage"
	"omar-kada/air-compose/internal/users"
	"omar-kada/air-compose/models"

	"github.com/rs/cors"
)

// Server will listen to requests on a port
type Server interface {
	Serve(params models.ServerParams) error
	Shutdown(ctx context.Context)
}

// HTTPServer is responsible for listening and mapping http requests
type HTTPServer struct {
	configStore      storage.ConfigStore
	processSvc       process.Service
	userSvc          users.Service
	websocketHandler *WebsocketHandler
	server           *http.Server
}

// NewServer creates a new http server
func NewServer(configStore storage.ConfigStore, service process.Service, userService users.Service) Server {
	return &HTTPServer{
		configStore:      configStore,
		processSvc:       service,
		userSvc:          userService,
		websocketHandler: newWebsocketHandler(),
	}
}

// Serve initializes routes from generated api and serves on the given port
func (s *HTTPServer) Serve(params models.ServerParams) error {
	// Create a new serve mux
	mux := http.NewServeMux()

	// Add frontend file server
	mux.HandleFunc("/ws", s.websocketHandler.handle)
	mux.HandleFunc("/", spaHandler)

	// create a type that satisfies the `api.ServerInterface`, which contains an implementation of every operation from the generated code
	myHandler := NewHandler(s.configStore, s.processSvc, s.userSvc)
	strict := api.NewStrictHandler(myHandler, []api.StrictMiddlewareFunc{})

	// get an `http.Handler` that we can use
	h := api.HandlerFromMux(strict, mux)
	h = middlewares.AuthorizationMiddleware(h)
	h = middlewares.AuthnMiddleware(h, s.userSvc, params.IsServerSecure())
	// Set up the CORS filter
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"localhost:*", "127.0.0.1:*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	})

	// Use the CORS filter as a middleware
	h = c.Handler(h)

	// api.HandlerWithOptions(strict, api.StdHTTPServerOptions{
	// 	BaseRouter: mux,
	// 	Middlewares: []api.MiddlewareFunc{
	// 		s.checkUsersMiddleware,
	// 		c.Handler,
	// 		loggingMiddleware,
	// 	},
	// })
	s.server = &http.Server{
		Handler:           middlewares.LoggingMiddleware(h),
		Addr:              ":" + strconv.Itoa(params.Port),
		ReadHeaderTimeout: 3 * time.Second,
	}
	slog.Info("server starting", "port", params.Port)

	// And we serve HTTP until the world ends.
	return s.server.ListenAndServe()
}

// Shutdown closes the server
func (s *HTTPServer) Shutdown(ctx context.Context) {
	s.server.Shutdown(ctx)
}
