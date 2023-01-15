package apiServer

import (
	apiMethod "api_indexer/cmd/main/pkg/apiMethods"
	"flag"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/go-chi/docgen"
	"github.com/go-chi/render"
)

var routes = flag.Bool("routes", false, "Generate router documentation")

type Server struct {
	Router *chi.Mux
	// Db, config can be added here
}

func CreateNewServer() *Server {
	s := &Server{}
	s.Router = chi.NewRouter()
	return s
}

func (s *Server) MountHandlers() {

	// Mount all Middleware here ,A good base middleware stack
	s.Router.Use(middleware.RequestID)
	s.Router.Use(middleware.RealIP)
	s.Router.Use(middleware.Logger)
	s.Router.Use(middleware.Recoverer)
	s.Router.Use(render.SetContentType(render.ContentTypeJSON))
	s.Router.Use(middleware.AllowContentType("application/json"))

	// processing should be stopped.
	s.Router.Use(middleware.Timeout(60 * time.Second))

	// Basic CORS
	s.Router.Use(cors.Handler(cors.Options{

		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))

	// Mount all Middleware here

	if *routes {
		// fmt.Println(docgen.JSONRoutesDoc(r))
		fmt.Println(docgen.MarkdownRoutesDoc(s.Router, docgen.MarkdownOpts{
			ProjectPath: "github.com/go-chi/chi/v5",
			Intro:       "Welcome to the chi/_examples/rest generated docs.",
		}))
		return
	}

}

func (s *Server) ApiMethods() {

	s.Router.Get("/", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("Api indexer by Luis Martinez")) })
	s.Router.Get("/panic", func(w http.ResponseWriter, r *http.Request) { panic("test") })
	// REST y routes for "mail" resource
	s.Router.Get("/mails/", apiMethod.AllMails)
	s.Router.Get("/mails/filter/", apiMethod.FilterMails)
}
