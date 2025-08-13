package httpserver

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httprate"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sunzhqr/phonebook/internal/config"
	"github.com/sunzhqr/phonebook/internal/handler"
	"github.com/sunzhqr/phonebook/internal/logger"
	"github.com/sunzhqr/phonebook/internal/service"
	"net/http"
	"runtime/debug"
	"time"
)

type Server struct {
	http *http.Server
	lg   *logger.Logger
}

func New(lg *logger.Logger, cfg config.Config, svc service.ContactsService) *Server {
	r := chi.NewRouter()
	r.Use(requestID())
	r.Use(recoverer(lg))
	r.Use(serverHeader())
	r.Use(httprate.LimitByIP(200, time.Minute))

	h := handler.New(lg, svc)

	r.Get("/metrics", promhttp.Handler().ServeHTTP)
	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/contacts", h.ListContacts)
		r.Get("/contacts/search", h.Search)
		r.Get("/contacts/{id}", h.GetContact)
		r.Post("/contacts", h.CreateContact)
		r.Put("/contacts/{id}", h.UpdateContact)
		r.Delete("/contacts/{id}", h.DeleteContact)
	})

	srv := &http.Server{
		Addr:              cfg.HTTP.Addr,
		Handler:           r,
		ReadTimeout:       cfg.HTTP.ReadTimeout,
		ReadHeaderTimeout: cfg.HTTP.ReadHeaderTimeout,
		WriteTimeout:      cfg.HTTP.WriteTimeout,
		IdleTimeout:       cfg.HTTP.IdleTimeout,
		MaxHeaderBytes:    cfg.HTTP.MaxHeaderBytes,
	}
	return &Server{http: srv, lg: lg}
}

func (s *Server) Start() error                   { return s.http.ListenAndServe() }
func (s *Server) Stop(ctx context.Context) error { return s.http.Shutdown(ctx) }

const reqIDKey = "req_id"

func requestID() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var b [16]byte
			_, _ = rand.Read(b[:])
			id := make([]byte, hex.EncodedLen(len(b)))
			hex.Encode(id, b[:])
			r = r.WithContext(context.WithValue(r.Context(), reqIDKey, string(id)))
			w.Header().Set("X-Request-ID", string(id))
			next.ServeHTTP(w, r)
		})
	}
}

func serverHeader() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Server", "phonebook/1.0")
			next.ServeHTTP(w, r)
		})
	}
}

func recoverer(lg *logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					lg.Error("panic", logger.KV("panic", rec), logger.KV("stack", string(debug.Stack())))
					http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
