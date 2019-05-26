package bridge

import (
	"net/http"
)

type dummies map[string]struct{}

func (dummies) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func (s *Server) createDummies() dummies {
	return map[string]struct{}{}
}

func (s *Server) getDummies(w http.ResponseWriter, r *http.Request) {
	g := s.createDummies()
	renderOK(w, r, g)
}
