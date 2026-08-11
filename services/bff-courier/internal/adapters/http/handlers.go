package httpadapter
import ("encoding/json";"net/http";"time";"github.com/nexora/bff-courier/internal/app")
func NewServer(addr string, d *app.Deps) *http.Server {
  mux := http.NewServeMux()
  mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request){ json.NewEncoder(w).Encode(map[string]string{"status":"ok"}) })
  mux.HandleFunc("GET /ready", func(w http.ResponseWriter, r *http.Request){ json.NewEncoder(w).Encode(map[string]string{"status":"ready"}) })
  mux.HandleFunc("POST /v1/courier/duty", func(w http.ResponseWriter, r *http.Request){
    var b struct{ CourierID string `json:"courierId"`; On bool `json:"on"` }
    _ = json.NewDecoder(r.Body).Decode(&b)
    res,_ := d.Duty(r.Context(), r.Header.Get("X-Tenant-Id"), b.CourierID, b.On)
    json.NewEncoder(w).Encode(res)
  })
  mux.HandleFunc("POST /v1/courier/offers/{id}", func(w http.ResponseWriter, r *http.Request){
    var b struct{ CourierID string `json:"courierId"`; Accept bool `json:"accept"` }
    _ = json.NewDecoder(r.Body).Decode(&b)
    res,_ := d.Offer(r.Context(), r.Header.Get("X-Tenant-Id"), b.CourierID, r.PathValue("id"), b.Accept)
    json.NewEncoder(w).Encode(res)
  })
  return &http.Server{Addr: addr, Handler: mux, ReadHeaderTimeout: 5*time.Second}
}
