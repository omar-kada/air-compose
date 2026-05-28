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
	websocketHandler *WebsocketHandler

	server *http.Server
}

// NewServer creates a new http server
func NewServer() Server {
	return &HTTPServer{
		websocketHandler: newWebsocketHandler(),
	}
}

// Serve initializes routes from generated api and serves on the given port
func (s *HTTPServer) Serve(
	params models.ServerParams,
	businessHandler api.StrictServerInterface,
	userSvc users.Service,
	oidcSvc users.OidcService,
) error {
	// Create a new serve mux
	mux := http.NewServeMux()

	// Add frontend file server
	mux.HandleFunc("/ws", s.websocketHandler.handle)
	mux.HandleFunc("/", spaHandler)

	// create a type that satisfies the `api.ServerInterface`, which contains an implementation of every operation from the generated code
	strict := api.NewStrictHandler(businessHandler, []api.StrictMiddlewareFunc{})

	// get an `http.Handler` that we can use
	h := api.HandlerFromMux(strict, mux)
	h = middlewares.AuthorizationMiddleware(h)
	h = middlewares.AuthnMiddleware(h, userSvc)
	h = middlewares.OidcMiddleware(h, oidcSvc)
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
