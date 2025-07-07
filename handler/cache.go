package handler

import (
	"cache/core"
	"cache/interface"
	"encoding/json"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func GetCacheHandler(service *core.CacheService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		topic := chi.URLParam(r, "topic")
		key := chi.URLParam(r, "key")
		val, err := service.Get(topic, key)
		if err != nil {
			http.Error(w, "failed to get cache", http.StatusInternalServerError)
			return
		}
		w.Write([]byte(val))
	}
}

func SetCacheHandler(service *core.CacheService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		topic := chi.URLParam(r, "topic")
		key := chi.URLParam(r, "key")
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "invalid body", http.StatusBadRequest)
			return
		}
		var payload struct {
			Value string `json:"value"`
			TTL   int    `json:"ttl"`
		}
		if err := json.Unmarshal(body, &payload); err != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}
		if err := service.Set(topic, key, payload.Value, payload.TTL); err != nil {
			http.Error(w, "failed to set cache", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

func InvalidateHandler(service *core.CacheService, broker _interface.IEventBroker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		topic := chi.URLParam(r, "topic")
		key := chi.URLParam(r, "key")

		// 무효화 처리
		if err := service.Invalidate(topic, key); err != nil {
			http.Error(w, "failed to invalidate", http.StatusInternalServerError)
			return
		}

		// Kafka 브로드캐스트
		if err := broker.Publish(topic, key); err != nil {
			http.Error(w, "failed to publish", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
