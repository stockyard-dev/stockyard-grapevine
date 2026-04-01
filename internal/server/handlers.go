package server
import("encoding/json";"net/http";"strconv";"github.com/stockyard-dev/stockyard-grapevine/internal/store")
func(s *Server)handleList(w http.ResponseWriter,r *http.Request){q:=r.URL.Query().Get("q");cat:=r.URL.Query().Get("category");list,_:=s.db.List(q,cat);if list==nil{list=[]store.Article{}};writeJSON(w,200,list)}
func(s *Server)handleCreate(w http.ResponseWriter,r *http.Request){var a store.Article;json.NewDecoder(r.Body).Decode(&a);if a.Title==""{writeError(w,400,"title required");return};if a.Category==""{a.Category="general"};s.db.Create(&a);writeJSON(w,201,a)}
func(s *Server)handleView(w http.ResponseWriter,r *http.Request){id,_:=strconv.ParseInt(r.PathValue("id"),10,64);s.db.View(id);writeJSON(w,200,map[string]string{"status":"viewed"})}
func(s *Server)handleUpdate(w http.ResponseWriter,r *http.Request){id,_:=strconv.ParseInt(r.PathValue("id"),10,64);var a store.Article;json.NewDecoder(r.Body).Decode(&a);s.db.Update(id,a.Title,a.Body,a.Category);writeJSON(w,200,map[string]string{"status":"updated"})}
func(s *Server)handleDelete(w http.ResponseWriter,r *http.Request){id,_:=strconv.ParseInt(r.PathValue("id"),10,64);s.db.Delete(id);writeJSON(w,200,map[string]string{"status":"deleted"})}
func(s *Server)handleOverview(w http.ResponseWriter,r *http.Request){m,_:=s.db.Stats();writeJSON(w,200,m)}
