package handler

import (
	"cache/core"
	_interface "cache/interface"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

func RegisterCacheRoutes(r *mux.Router, service core.CacheService, broker _interface.IEventBroker) {
	r.HandleFunc("/cache/get", func(w http.ResponseWriter, r *http.Request) {
		topic := r.URL.Query().Get("topic")
		key := r.URL.Query().Get("key")
		if key == "" || topic == "" {
			http.Error(w, "missing topic or key", http.StatusBadRequest)
			return
		}
		val, err := service.Get(topic, key)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		err = json.NewEncoder(w).Encode(map[string]string{"key": key, "value": val})
		if err != nil {
			return
		}
	}).Methods("GET")

	r.HandleFunc("/cache/set", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Key   string `json:"key"`
			Value string `json:"value"`
			Topic string `json:"topic"`
			TTL   int    `json:"ttl_seconds"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}
		if err := service.Set(req.Topic, req.Key, req.Value, req.TTL); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if p, ok := broker.(interface{ PublishTo(topic, key string) error }); ok {
			if err := p.PublishTo(req.Topic, req.Key); err != nil {
				http.Error(w, "publish failed", http.StatusInternalServerError)
				return
			}
		}
		err := json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
		if err != nil {
			return
		}
	}).Methods("POST")

	r.HandleFunc("/cache/delete", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Key   string `json:"key"`
			Topic string `json:"topic"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}
		if err := service.Invalidate(req.Topic, req.Key); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if p, ok := broker.(interface{ PublishTo(topic, key string) error }); ok {
			if err := p.PublishTo(req.Topic, req.Key); err != nil {
				http.Error(w, "publish failed", http.StatusInternalServerError)
				return
			}
		}
		err := json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
		if err != nil {
			return
		}
	}).Methods("POST")
}
