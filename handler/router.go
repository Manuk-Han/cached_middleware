package handler

import (
	"cache/core"
	"cache/interface"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func NewRouter(cacheService *core.CacheService, broker _interface.IEventBroker) http.Handler {
	r := chi.NewRouter()

	r.Get("/cache/{topic}/{key}", GetCacheHandler(cacheService))
	r.Post("/cache/{topic}/{key}", SetCacheHandler(cacheService))
	r.Post("/invalidate/{topic}/{key}", InvalidateHandler(cacheService, broker))

	return r
}
