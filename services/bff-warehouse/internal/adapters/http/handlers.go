package httpadapter
import ("encoding/json";"net/http";"time";"github.com/nexora/bff-warehouse/internal/app")
func NewServer(addr string, d *app.Deps) *http.Server {
  mux:=http.NewServeMux()
  mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request){ json.NewEncoder(w).Encode(map[string]string{"status":"ok"}) })
  mux.HandleFunc("GET /ready", func(w http.ResponseWriter, r *http.Request){ json.NewEncoder(w).Encode(map[string]string{"status":"ready"}) })
  mux.HandleFunc("POST /v1/warehouse/tasks/{id}/pick", func(w http.ResponseWriter, r *http.Request){ res,_:=d.Pick(r.Context(), r.Header.Get("X-Tenant-Id"), r.PathValue("id")); json.NewEncoder(w).Encode(res) })
  mux.HandleFunc("POST /v1/warehouse/tasks/{id}/pack", func(w http.ResponseWriter, r *http.Request){ res,_:=d.Pack(r.Context(), r.Header.Get("X-Tenant-Id"), r.PathValue("id")); json.NewEncoder(w).Encode(res) })
  mux.HandleFunc("POST /v1/warehouse/tasks/{id}/ready", func(w http.ResponseWriter, r *http.Request){ res,_:=d.DispatchReady(r.Context(), r.Header.Get("X-Tenant-Id"), r.PathValue("id")); json.NewEncoder(w).Encode(res) })
  return &http.Server{Addr:addr, Handler:mux, ReadHeaderTimeout:5*time.Second}
}
