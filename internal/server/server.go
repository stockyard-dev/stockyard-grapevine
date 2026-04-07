package server

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/stockyard-dev/stockyard-grapevine/internal/store"
)

const resourceName = "articles"

type Server struct {
	db      *store.DB
	mux     *http.ServeMux
	limits  Limits
	dataDir string
	pCfg    map[string]json.RawMessage
}

func New(db *store.DB, limits Limits, dataDir string) *Server {
	s := &Server{
		db:      db,
		mux:     http.NewServeMux(),
		limits:  limits,
		dataDir: dataDir,
	}
	s.loadPersonalConfig()

	// Articles CRUD
	s.mux.HandleFunc("GET /api/articles", s.list)
	s.mux.HandleFunc("POST /api/articles", s.create)
	s.mux.HandleFunc("GET /api/articles/{id}", s.get)
	s.mux.HandleFunc("PUT /api/articles/{id}", s.update)
	s.mux.HandleFunc("DELETE /api/articles/{id}", s.del)

	// Lookup by slug (for public help center)
	s.mux.HandleFunc("GET /api/articles/by-slug/{slug}", s.getBySlug)

	// Voting (atomic)
	s.mux.HandleFunc("POST /api/articles/{id}/helpful", s.markHelpful)
	s.mux.HandleFunc("POST /api/articles/{id}/not-helpful", s.markNotHelpful)

	// Stats / health
	s.mux.HandleFunc("GET /api/stats", s.stats)
	s.mux.HandleFunc("GET /api/health", s.health)

	// Personalization
	s.mux.HandleFunc("GET /api/config", s.configHandler)

	// Extras
	s.mux.HandleFunc("GET /api/extras/{resource}", s.listExtras)
	s.mux.HandleFunc("GET /api/extras/{resource}/{id}", s.getExtras)
	s.mux.HandleFunc("PUT /api/extras/{resource}/{id}", s.putExtras)

	// Tier
	s.mux.HandleFunc("GET /api/tier", func(w http.ResponseWriter, r *http.Request) {
		wj(w, 200, map[string]any{
			"tier":        s.limits.Tier,
			"upgrade_url": "https://stockyard.dev/grapevine/",
		})
	})

	// Dashboard
	s.mux.HandleFunc("GET /ui", s.dashboard)
	s.mux.HandleFunc("GET /ui/", s.dashboard)
	s.mux.HandleFunc("GET /", s.root)

	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

// ─── helpers ──────────────────────────────────────────────────────

func wj(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(v)
}

func we(w http.ResponseWriter, code int, msg string) {
	wj(w, code, map[string]string{"error": msg})
}

func oe[T any](s []T) []T {
	if s == nil {
		return []T{}
	}
	return s
}

func (s *Server) root(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	http.Redirect(w, r, "/ui", 302)
}

// ─── personalization ──────────────────────────────────────────────

func (s *Server) loadPersonalConfig() {
	path := filepath.Join(s.dataDir, "config.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	var cfg map[string]json.RawMessage
	if err := json.Unmarshal(data, &cfg); err != nil {
		log.Printf("grapevine: warning: could not parse config.json: %v", err)
		return
	}
	s.pCfg = cfg
	log.Printf("grapevine: loaded personalization from %s", path)
}

func (s *Server) configHandler(w http.ResponseWriter, r *http.Request) {
	if s.pCfg == nil {
		wj(w, 200, map[string]any{})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s.pCfg)
}

// ─── extras ───────────────────────────────────────────────────────

func (s *Server) listExtras(w http.ResponseWriter, r *http.Request) {
	resource := r.PathValue("resource")
	all := s.db.AllExtras(resource)
	out := make(map[string]json.RawMessage, len(all))
	for id, data := range all {
		out[id] = json.RawMessage(data)
	}
	wj(w, 200, out)
}

func (s *Server) getExtras(w http.ResponseWriter, r *http.Request) {
	resource := r.PathValue("resource")
	id := r.PathValue("id")
	data := s.db.GetExtras(resource, id)
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(data))
}

func (s *Server) putExtras(w http.ResponseWriter, r *http.Request) {
	resource := r.PathValue("resource")
	id := r.PathValue("id")
	body, err := io.ReadAll(r.Body)
	if err != nil {
		we(w, 400, "read body")
		return
	}
	var probe map[string]any
	if err := json.Unmarshal(body, &probe); err != nil {
		we(w, 400, "invalid json")
		return
	}
	if err := s.db.SetExtras(resource, id, string(body)); err != nil {
		we(w, 500, "save failed")
		return
	}
	wj(w, 200, map[string]string{"ok": "saved"})
}

// ─── articles ─────────────────────────────────────────────────────

func (s *Server) list(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	filters := map[string]string{}
	if v := r.URL.Query().Get("category"); v != "" {
		filters["category"] = v
	}
	if v := r.URL.Query().Get("status"); v != "" {
		filters["status"] = v
	}
	if q != "" || len(filters) > 0 {
		wj(w, 200, map[string]any{"articles": oe(s.db.Search(q, filters))})
		return
	}
	wj(w, 200, map[string]any{"articles": oe(s.db.List())})
}

func (s *Server) create(w http.ResponseWriter, r *http.Request) {
	if s.limits.MaxItems > 0 && s.db.Count() >= s.limits.MaxItems {
		we(w, 402, "Free tier limit reached. Upgrade at https://stockyard.dev/grapevine/")
		return
	}
	var e store.Article
	if err := json.NewDecoder(r.Body).Decode(&e); err != nil {
		we(w, 400, "invalid json")
		return
	}
	if e.Title == "" {
		we(w, 400, "title required")
		return
	}
	if err := s.db.Create(&e); err != nil {
		we(w, 500, "create failed")
		return
	}
	wj(w, 201, s.db.Get(e.ID))
}

func (s *Server) get(w http.ResponseWriter, r *http.Request) {
	e := s.db.Get(r.PathValue("id"))
	if e == nil {
		we(w, 404, "not found")
		return
	}
	wj(w, 200, e)
}

func (s *Server) getBySlug(w http.ResponseWriter, r *http.Request) {
	e := s.db.GetBySlug(r.PathValue("slug"))
	if e == nil {
		we(w, 404, "not found")
		return
	}
	wj(w, 200, e)
}

// update accepts a partial article. Original handler had 'if x == 0
// preserve' for both helpful and not_helpful counters, preventing
// admin reset of disputed vote counts.
func (s *Server) update(w http.ResponseWriter, r *http.Request) {
	existing := s.db.Get(r.PathValue("id"))
	if existing == nil {
		we(w, 404, "not found")
		return
	}

	var raw map[string]json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
		we(w, 400, "invalid json")
		return
	}

	patch := *existing
	if v, ok := raw["title"]; ok {
		var s string
		json.Unmarshal(v, &s)
		if s != "" {
			patch.Title = s
		}
	}
	if v, ok := raw["body"]; ok {
		json.Unmarshal(v, &patch.Body)
	}
	if v, ok := raw["category"]; ok {
		json.Unmarshal(v, &patch.Category)
	}
	if v, ok := raw["tags"]; ok {
		json.Unmarshal(v, &patch.Tags)
	}
	if v, ok := raw["slug"]; ok {
		json.Unmarshal(v, &patch.Slug)
	}
	if v, ok := raw["helpful"]; ok {
		json.Unmarshal(v, &patch.Helpful)
	}
	if v, ok := raw["not_helpful"]; ok {
		json.Unmarshal(v, &patch.NotHelpful)
	}
	if v, ok := raw["status"]; ok {
		var s string
		json.Unmarshal(v, &s)
		if s != "" {
			patch.Status = s
		}
	}

	if err := s.db.Update(&patch); err != nil {
		we(w, 500, "update failed")
		return
	}
	wj(w, 200, s.db.Get(patch.ID))
}

func (s *Server) del(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	s.db.Delete(id)
	s.db.DeleteExtras(resourceName, id)
	wj(w, 200, map[string]string{"deleted": "ok"})
}

// markHelpful atomically increments the helpful counter.
func (s *Server) markHelpful(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if s.db.Get(id) == nil {
		we(w, 404, "not found")
		return
	}
	v, err := s.db.MarkHelpful(id)
	if err != nil {
		we(w, 500, "vote failed")
		return
	}
	wj(w, 200, map[string]any{"helpful": v})
}

func (s *Server) markNotHelpful(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if s.db.Get(id) == nil {
		we(w, 404, "not found")
		return
	}
	v, err := s.db.MarkNotHelpful(id)
	if err != nil {
		we(w, 500, "vote failed")
		return
	}
	wj(w, 200, map[string]any{"not_helpful": v})
}

func (s *Server) stats(w http.ResponseWriter, r *http.Request) {
	wj(w, 200, s.db.Stats())
}

func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	wj(w, 200, map[string]any{
		"status":   "ok",
		"service":  "grapevine",
		"articles": s.db.Count(),
	})
}

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}
