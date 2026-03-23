package server

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/davidklassen/amort/explore"
	"github.com/davidklassen/amort/store"
	"github.com/davidklassen/amort/web"
)

type Server struct {
	store *store.Store
	loop  *explore.Loop
	mux   *http.ServeMux
}

func New(s *store.Store, loop *explore.Loop) *Server {
	srv := &Server{store: s, loop: loop, mux: http.NewServeMux()}
	srv.mux.Handle("GET /", http.FileServerFS(web.Static))
	srv.mux.HandleFunc("GET /api/proposals", srv.handleList)
	srv.mux.HandleFunc("GET /api/proposals/{id}", srv.handleGet)
	srv.mux.HandleFunc("POST /api/proposals", srv.handleCreate)
	srv.mux.HandleFunc("POST /api/proposals/{id}/approve", srv.handleApprove)
	srv.mux.HandleFunc("POST /api/proposals/{id}/reject", srv.handleReject)
	return srv
}

func (s *Server) Handler() http.Handler {
	return s.mux
}

func (s *Server) handleList(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	proposals, err := s.store.List(status)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	writeJSON(w, proposals)
}

func (s *Server) handleCreate(w http.ResponseWriter, r *http.Request) {
	var p store.Proposal
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	if p.ID == "" || p.Title == "" || p.Summary == "" {
		http.Error(w, "id, title, and summary are required", 400)
		return
	}
	if err := s.store.Insert(&p); err != nil {
		slog.Error("failed to create proposal", "id", p.ID, "error", err)
		http.Error(w, err.Error(), 500)
		return
	}
	slog.Info("proposal created", "id", p.ID, "title", p.Title)
	w.WriteHeader(201)
}

func (s *Server) handleGet(w http.ResponseWriter, r *http.Request) {
	p, err := s.store.Get(r.PathValue("id"))
	if err != nil {
		http.Error(w, "not found", 404)
		return
	}
	writeJSON(w, p)
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("failed to write json response", "error", err)
	}
}

func (s *Server) handleApprove(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := s.store.Approve(id); err != nil {
		http.Error(w, "not found", 404)
		return
	}
	slog.Info("proposal approved", "id", id)
	s.loop.Signal()
	w.WriteHeader(200)
}

func (s *Server) handleReject(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := s.store.Reject(id); err != nil {
		http.Error(w, "not found", 404)
		return
	}
	slog.Info("proposal rejected", "id", id)
	s.loop.Signal()
	w.WriteHeader(200)
}
