package service

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
)

func (s *Service) RegisterRoutes(mux *http.ServeMux) {
	slog.Info("registering service routes")

	mux.HandleFunc("/api/search", makeHandler(s.Search))
	mux.HandleFunc("/api/process_document", makeHandler(s.ProcessDocument))
	mux.HandleFunc("/api/batch_process_documents", makeHandler(s.BatchProcessDocuments))
}

func makeHandler[Req, Res any](handler func(context.Context, *Req) (Res, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req *Req
		if r.Method != "GET" && r.Method != "DELETE" {
			req = new(Req)
			if err := json.NewDecoder(r.Body).Decode(req); err != nil {
				slog.Error("failed to decode request", slog.String("error", err.Error()))
				http.Error(w, "invalid request body", http.StatusBadRequest)
				return
			}
		}
		res, err := handler(r.Context(), req)
		if err != nil {
			slog.Error("handler error", slog.String("error", err.Error()))
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(res); err != nil {
			slog.Error("failed to encode response", slog.String("error", err.Error()))
			http.Error(w, "failed to encode response", http.StatusInternalServerError)
			return
		}
	}
}
