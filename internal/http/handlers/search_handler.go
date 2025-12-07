package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/gin-gonic/gin"
)

type SearchHandler struct {
	es    *elasticsearch.Client
	index string
}

func NewSearchHandler(es *elasticsearch.Client, index string) *SearchHandler {
	return &SearchHandler{es: es, index: index}
}

// Search GET /api/v1/search?query=xxx
func (h *SearchHandler) Search(c *gin.Context) {
	q := strings.TrimSpace(c.Query("query"))
	if q == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query is required"})
		return
	}

	body := map[string]any{
		"query": map[string]any{
			"multi_match": map[string]any{
				"query":  q,
				"fields": []string{"title^2", "content"},
			},
		},
	}
	b, err := json.Marshal(body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}

	res, err := h.es.Search(
		h.es.Search.WithContext(context.Background()),
		h.es.Search.WithIndex(h.index),
		h.es.Search.WithBody(strings.NewReader(string(b))),
		h.es.Search.WithTrackTotalHits(true),
		// simple size limit
		h.es.Search.WithSize(20),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "search_error"})
		return
	}
	defer res.Body.Close()

	if res.IsError() {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "search_es_error"})
		return
	}

	var parsed map[string]any
	if err := json.NewDecoder(res.Body).Decode(&parsed); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "decode_error"})
		return
	}

	c.JSON(http.StatusOK, parsed)
}
