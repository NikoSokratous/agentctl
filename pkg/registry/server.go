package registry

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

// Server is the plugin marketplace HTTP server.
type Server struct {
	registry  *Registry
	validator *SecurityValidator
	router    *mux.Router
}

// NewServer creates a new marketplace server.
func NewServer(registry *Registry, validator *SecurityValidator) *Server {
	s := &Server{
		registry:  registry,
		validator: validator,
		router:    mux.NewRouter(),
	}

	s.setupRoutes()
	return s
}

// setupRoutes configures HTTP routes.
func (s *Server) setupRoutes() {
	api := s.router.PathPrefix("/v1").Subrouter()

	// Plugin operations
	api.HandleFunc("/plugins", s.handleListPlugins).Methods("GET")
	api.HandleFunc("/plugins", s.handlePublishPlugin).Methods("POST")
	api.HandleFunc("/plugins/{id}", s.handleGetPlugin).Methods("GET")
	api.HandleFunc("/plugins/{id}/download", s.handleDownloadPlugin).Methods("GET")
	api.HandleFunc("/plugins/{id}/stats", s.handleGetStats).Methods("GET")

	// Reviews
	api.HandleFunc("/plugins/{id}/reviews", s.handleGetReviews).Methods("GET")
	api.HandleFunc("/plugins/{id}/reviews", s.handleSubmitReview).Methods("POST")

	// Discovery
	api.HandleFunc("/plugins/search", s.handleSearchPlugins).Methods("GET")
	api.HandleFunc("/plugins/trending", s.handleGetTrending).Methods("GET")
}

// ServeHTTP implements http.Handler.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

// handleListPlugins lists all plugins.
func (s *Server) handleListPlugins(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse query parameters
	filter := &SearchFilter{
		Type:      PluginType(r.URL.Query().Get("type")),
		Runtime:   RuntimeType(r.URL.Query().Get("runtime")),
		Author:    r.URL.Query().Get("author"),
		SortBy:    r.URL.Query().Get("sort_by"),
		SortOrder: r.URL.Query().Get("sort_order"),
	}

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil {
			filter.Limit = limit
		}
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil {
			filter.Offset = offset
		}
	}

	// Search plugins
	result, err := s.registry.Search(ctx, filter)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	s.writeJSON(w, http.StatusOK, result)
}

// handlePublishPlugin publishes a new plugin.
func (s *Server) handlePublishPlugin(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse multipart form
	if err := r.ParseMultipartForm(50 << 20); err != nil { // 50MB max
		s.writeError(w, http.StatusBadRequest, "invalid form data")
		return
	}

	// Read manifest
	manifestFile, _, err := r.FormFile("manifest")
	if err != nil {
		s.writeError(w, http.StatusBadRequest, "manifest file required")
		return
	}
	defer manifestFile.Close()

	manifestData, err := io.ReadAll(manifestFile)
	if err != nil {
		s.writeError(w, http.StatusBadRequest, "failed to read manifest")
		return
	}

	var manifest PluginManifest
	if err := json.Unmarshal(manifestData, &manifest); err != nil {
		s.writeError(w, http.StatusBadRequest, "invalid manifest JSON")
		return
	}

	// Validate metadata
	validationResult, err := s.validator.Validate(&manifest.Metadata)
	if err != nil || !validationResult.Valid {
		s.writeJSON(w, http.StatusBadRequest, map[string]interface{}{
			"error":      "validation failed",
			"validation": validationResult,
		})
		return
	}

	// Read artifact
	artifactFile, _, err := r.FormFile("artifact")
	if err != nil {
		s.writeError(w, http.StatusBadRequest, "artifact file required")
		return
	}
	defer artifactFile.Close()

	artifactData, err := io.ReadAll(artifactFile)
	if err != nil {
		s.writeError(w, http.StatusBadRequest, "failed to read artifact")
		return
	}

	// Register plugin
	if err := s.registry.Register(ctx, &manifest, artifactData); err != nil {
		s.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	s.writeJSON(w, http.StatusCreated, map[string]interface{}{
		"message":    "plugin published successfully",
		"plugin_id":  manifest.Metadata.ID,
		"validation": validationResult,
	})
}

// handleGetPlugin retrieves plugin details.
func (s *Server) handleGetPlugin(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	pluginID := vars["id"]

	plugin, err := s.registry.Get(ctx, pluginID)
	if err != nil {
		s.writeError(w, http.StatusNotFound, "plugin not found")
		return
	}

	s.writeJSON(w, http.StatusOK, plugin)
}

// handleDownloadPlugin handles plugin downloads.
func (s *Server) handleDownloadPlugin(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	pluginID := vars["id"]

	// Get plugin metadata
	plugin, err := s.registry.Get(ctx, pluginID)
	if err != nil {
		s.writeError(w, http.StatusNotFound, "plugin not found")
		return
	}

	// Increment download counter
	s.registry.IncrementDownloads(ctx, pluginID)

	// Placeholder: would serve actual artifact file
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s-%s.tar.gz", plugin.Name, plugin.Version))
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("plugin artifact placeholder"))
}

// handleGetStats retrieves plugin statistics.
func (s *Server) handleGetStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	pluginID := vars["id"]

	stats, err := s.registry.GetStats(ctx, pluginID)
	if err != nil {
		s.writeError(w, http.StatusNotFound, "plugin not found")
		return
	}

	s.writeJSON(w, http.StatusOK, stats)
}

// handleGetReviews retrieves plugin reviews.
func (s *Server) handleGetReviews(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	pluginID := vars["id"]

	limit := 50
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	reviews, err := s.registry.GetReviews(ctx, pluginID, limit)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	s.writeJSON(w, http.StatusOK, reviews)
}

// handleSubmitReview submits a plugin review.
func (s *Server) handleSubmitReview(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	pluginID := vars["id"]

	var review PluginReview
	if err := json.NewDecoder(r.Body).Decode(&review); err != nil {
		s.writeError(w, http.StatusBadRequest, "invalid review data")
		return
	}

	review.PluginID = pluginID

	if err := s.registry.AddReview(ctx, &review); err != nil {
		s.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	s.writeJSON(w, http.StatusCreated, review)
}

// handleSearchPlugins searches plugins.
func (s *Server) handleSearchPlugins(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	filter := &SearchFilter{
		Query:     r.URL.Query().Get("q"),
		Type:      PluginType(r.URL.Query().Get("type")),
		Runtime:   RuntimeType(r.URL.Query().Get("runtime")),
		Author:    r.URL.Query().Get("author"),
		SortBy:    r.URL.Query().Get("sort_by"),
		SortOrder: r.URL.Query().Get("sort_order"),
	}

	if minRatingStr := r.URL.Query().Get("min_rating"); minRatingStr != "" {
		if rating, err := strconv.ParseFloat(minRatingStr, 64); err == nil {
			filter.MinRating = rating
		}
	}

	result, err := s.registry.Search(ctx, filter)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	s.writeJSON(w, http.StatusOK, result)
}

// handleGetTrending retrieves trending plugins.
func (s *Server) handleGetTrending(w http.ResponseWriter, r *http.Request) {
	// Placeholder: would implement trending algorithm
	s.writeJSON(w, http.StatusOK, []PluginMetadata{})
}

// Helper methods

func (s *Server) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (s *Server) writeError(w http.ResponseWriter, status int, message string) {
	s.writeJSON(w, status, map[string]string{"error": message})
}
