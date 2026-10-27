package server

import (
	"log"
	"net/http"

	"go_final_project/internal/api"
)

type Server struct {
	Logger     *log.Logger
	httpServer *http.Server
}

func NewServer(logger *log.Logger, port string) *Server {
	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir("web")))
	api.Init(mux)

	httpServer := &http.Server{
		Addr:     ":" + port,
		Handler:  mux,
		ErrorLog: logger,
	}

	return &Server{
		Logger:     logger,
		httpServer: httpServer,
	}
}

func (s *Server) Run() error {
	s.Logger.Println("Starting server on " + s.httpServer.Addr)
	return s.httpServer.ListenAndServe()
}
