package server

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

func (s *Server) handlerSearch(resp http.ResponseWriter, req *http.Request) {
	v, e := req.URL.Query()["q"]
	if !e || len(v) == 0 {
		http.Error(resp, "Invalid \"q\" param", 400)
		return
	}
	ret := make([]map[string]any, 0)
	for _, v := range s.es.Search(v[0]) {
		tmp := map[string]any{}
		err := json.Unmarshal([]byte(v), &tmp)
		if err != nil {
			http.Error(resp, "Err", 500)
			return

		}
		ret = append(ret, tmp)
	}

	resp.Header().Set("Content-Type", "application/json")
	json.NewEncoder(resp).Encode(ret)
}

func (s *Server) setupHTTPAPI() {
	mux := http.NewServeMux()
	mux.HandleFunc("/search", s.handlerSearch)
	srv := &http.Server{Addr: "0.0.0.0:80", Handler: mux}

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			slog.Error("Http server closed: %v", "error", err)
		}
	}()

	go func() {
		<-s.ctx.Done()
		srv.Shutdown(s.ctx)
	}()
}
