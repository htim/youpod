package server

import (
	"github.com/htim/youpod/server/handler"
	"net/http"
)

type Server struct {
	Handler handler.Handler
}

func (s *Server) Run(addr string) error {
	return http.ListenAndServe(addr, s.Handler.Routes())
}
