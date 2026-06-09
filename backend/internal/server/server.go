// Package server provides implementations of http and ws handlers
package server

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"omar-kada/air-compose/api"
	"omar-kada/air-compose/internal/models"
	"omar-kada/air-compose/internal/server/middlewares"
	"omar-kada/air-compose/internal/users"

	"github.com/rs/cors"
)

// Server will listen to requests on a port
type Server interface {
	Serve(
		params models.ServerParams,
		businessHandler api.StrictServerInterface,
		userSvc users.Service,
		oidcSvc users.OidcService,
	) error
	Shutdown(ctx context.Context)
}

// HTTPServer is responsible for listening and mapping http requests
type HTTPServer struct {
	server *http.Server
}

// NewServer creates a new http server
func NewServer() Server {
	return &HTTPServer{}
}

// applyMiddlewares applies a list of middlewares to a handler
func applyMiddlewares(h http.Handler, mws ...func(http.Handler) http.Handler) http.Handler {
	for i := len(mws) - 1; i >= 0; i-- {
		h = mws[i](h)
	}
	return h
}

// Serve initializes routes from generated api and serves on the given port
func (s *HTTPServer) Serve(
	params models.ServerParams,
	businessHandler api.StrictServerInterface,
	userSvc users.Service,
	oidcSvc users.OidcService,
	// add clientEventsService
) error {
	// Create a new serve mux
	mux := http.NewServeMux()

	strict := api.NewStrictHandler(businessHandler, []api.StrictMiddlewareFunc{})
	mux.HandleFunc("/api/ws", webSocketHandler)
	mux.Handle("/api/", api.Handler(strict))
	mux.HandleFunc("/", spaHandler)

	// Set up the CORS filter
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"localhost:*", "127.0.0.1:*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	})

	s.server = &http.Server{
		Handler: applyMiddlewares(mux,
			middlewares.LoggingMiddleware,
			c.Handler,
			middlewares.OidcMiddlewareFunc(oidcSvc),
			middlewares.AuthnMiddlewareFunc(userSvc),
			middlewares.AuthorizationMiddleware,
		),
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
