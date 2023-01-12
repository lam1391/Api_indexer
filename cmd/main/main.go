package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/docgen"
	"github.com/go-chi/render"
)

var routes = flag.Bool("routes", false, "Generate router documentation")
var ErrNotFound = &ErrResponse{HTTPStatusCode: 404, StatusText: "Resource not found."}

type Server struct {
	Router *chi.Mux
	// Db, config can be added here
}

type Mail struct {
	Date    string `json:"Date"`
	From    string `json:"From"`
	Subject string `json:"Subject"`
	To      string `json:"To"` // the author
	Body    string `json:"Body"`
}

type MailResponse struct {
	*Mail
}

type ResponseMails struct {
	Took     int64  `json:"took"`
	Time_out bool   `json:"time_out"`
	ErrorM   string `json:"error"`
	Hits     struct {
		Total struct {
			Value int64 `json:"value"`
		} `json:"total"`
		Hits_2 []struct {
			Index    string `json:"_index"`
			Dtype    string `json:"_type"`
			IdM      string `json:"_id"`
			Score    int64  `json:"_score"`
			Timestap string `json:"@timestamp"`
			Source   Mail   `json:"_source"`
		} `json:"hits"`
	} `json:"hits"`
}

type ErrResponse struct {
	Err            error `json:"-"` // low-level runtime error
	HTTPStatusCode int   `json:"-"` // http response status code

	StatusText string `json:"status"`          // user-level status message
	AppCode    int64  `json:"code,omitempty"`  // application-specific error code
	ErrorText  string `json:"error,omitempty"` // application-level error message, for debugging
}

func main() {

	s := CreateNewServer()
	s.MountHandlers()
	http.ListenAndServe(":3033", s.Router)

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
		// AllowedOrigins:   []string{"https://foo.com"}, // Use this to allow specific origin hosts
		AllowedOrigins: []string{"https://*", "http://*"},
		// AllowOriginFunc:  func(r *http.Request, origin string) bool { return true },
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))

	// Mount all Middleware here
	s.Router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hi"))
	})

	s.Router.Get("/panic", func(w http.ResponseWriter, r *http.Request) {
		panic("test")
	})

	// RESTy routes for "mail" resource
	// s.Router.Get("/mails/{max:^[0-9]{1,2}}", ListMails)
	s.Router.Get("/mails/", ListMails)
	s.Router.Get("/mails/filter/", GetMail)

	if *routes {
		// fmt.Println(docgen.JSONRoutesDoc(r))
		fmt.Println(docgen.MarkdownRoutesDoc(s.Router, docgen.MarkdownOpts{
			ProjectPath: "github.com/go-chi/chi/v5",
			Intro:       "Welcome to the chi/_examples/rest generated docs.",
		}))
		return
	}

}

func GetMail(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("content-Type", "Application/json")
	w.Header().Set("done-by", "Luis Martinez")

	from := r.URL.Query().Get("from")
	max_items := r.URL.Query().Get("max")
	filter := r.URL.Query().Get("filterID")

	resp, err := my_search_filter(from, max_items, filter)

	if err != nil {
		w.WriteHeader(404)
	}

	json.NewEncoder(w).Encode(resp)
}

func ListMails(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("content-Type", "Application/json")
	w.Header().Set("done-by", "Luis Martinez")

	from := r.URL.Query().Get("from")
	max_items := r.URL.Query().Get("max")

	resp := my_search_all(from, max_items)

	json.NewEncoder(w).Encode(resp)

}

func (rd *MailResponse) Render(w http.ResponseWriter, r *http.Request) error {
	// Pre-processing before a response is marshalled and sent across the wire
	return nil
}

func NewMailListResponse(mails []*Mail) []render.Renderer {
	list := []render.Renderer{}
	for _, mail := range mails {
		list = append(list, NewMailResponse(mail))
	}
	return list
}

func NewMailResponse(mail *Mail) *MailResponse {
	resp := &MailResponse{Mail: mail}
	return resp
}

func (e *ErrResponse) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, e.HTTPStatusCode)
	return nil
}

func ErrRender(err error) render.Renderer {
	return &ErrResponse{
		Err:            err,
		HTTPStatusCode: 422,
		StatusText:     "Error rendering response.",
		ErrorText:      err.Error(),
	}
}

func ErrInvalidRequest(err error) render.Renderer {
	return &ErrResponse{
		Err:            err,
		HTTPStatusCode: 400,
		StatusText:     "Invalid request.",
		ErrorText:      err.Error(),
	}
}

func my_search_filter(from string, maxCount string, filter string) (ResponseMails, error) {

	query := `{
        "search_type": "match",
		"query":
        {
            "term":"` + filter + `"
        },
        "from":` + from + `,
        "max_results":` + maxCount + ` ,
        "_source": ["From","To","Date","Subject","body"]
    }`
	req, err := http.NewRequest("POST", "http://localhost:4080/api/maildir/_search", strings.NewReader(query))
	if err != nil {
		log.Fatal(err)
	}
	req.SetBasicAuth("admin", "Complexpass#123")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_4) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/81.0.4044.138 Safari/537.36")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	log.Println(resp.StatusCode)
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	// bodyString := string(body)
	var data ResponseMails
	json.Unmarshal(body, &data)

	if data.Hits.Total.Value == 0 {
		return data, errors.New("mail not found")
	}

	return data, nil

}

func my_search_all(from string, max_item string) ResponseMails {
	query := `{
        "search_type": "matchall",
        "from":` + from + `,
		"max_results":` + max_item + `,
		"_source": ["From","To","Date","Subject","body"]     }`

	req, err := http.NewRequest("POST", "http://localhost:4080/api/maildir/_search", strings.NewReader(query))
	if err != nil {
		log.Fatal(err)
	}
	req.SetBasicAuth("admin", "Complexpass#123")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_4) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/81.0.4044.138 Safari/537.36")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	log.Println(resp.StatusCode)
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	// bodyString := string(body)
	var data ResponseMails
	json.Unmarshal(body, &data)

	return data
}
