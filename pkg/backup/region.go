package backup

import (
	"context"
	"database/sql"
	"net/http"
	"strings"
	"sync"
)

// RegionAwareRouter routes reads to the nearest healthy replica and writes to primary.
type RegionAwareRouter struct {
	primary      *sql.DB
	replicas     []*ReplicaConfig
	regionHeader string // e.g. X-Region
	mu           sync.RWMutex
}

// RefreshHealth pings all replicas and updates Healthy status.
func (r *RegionAwareRouter) RefreshHealth(ctx context.Context) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, rep := range r.replicas {
		rep.Healthy = (rep.DB.PingContext(ctx) == nil)
	}
}

// NewRegionAwareRouter creates a router with primary and replicas.
func NewRegionAwareRouter(primary *sql.DB, replicas []*ReplicaConfig) *RegionAwareRouter {
	return &RegionAwareRouter{
		primary:      primary,
		replicas:     replicas,
		regionHeader: "X-Region",
	}
}

// SetRegionHeader sets the HTTP header used for region-aware routing.
func (r *RegionAwareRouter) SetRegionHeader(h string) {
	r.regionHeader = h
}

// Primary returns the primary database for writes.
func (r *RegionAwareRouter) Primary() *sql.DB {
	return r.primary
}

// Reader returns a DB suitable for reads. Prefers a replica in the request's region.
func (r *RegionAwareRouter) Reader(ctx context.Context, req *http.Request) *sql.DB {
	preferredRegion := ""
	if req != nil && r.regionHeader != "" {
		preferredRegion = strings.TrimSpace(req.Header.Get(r.regionHeader))
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	// Prefer replica in the same region
	if preferredRegion != "" {
		for _, rep := range r.replicas {
			if rep.Region == preferredRegion && rep.Healthy {
				return rep.DB
			}
		}
	}

	// Fall back to any healthy replica
	for _, rep := range r.replicas {
		if rep.Healthy {
			return rep.DB
		}
	}

	// No healthy replica: use primary (read-your-writes consistency)
	return r.primary
}

// ReaderForRegion returns a read DB for a given region.
func (r *RegionAwareRouter) ReaderForRegion(ctx context.Context, region string) *sql.DB {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, rep := range r.replicas {
		if rep.Region == region && rep.Healthy {
			return rep.DB
		}
	}
	for _, rep := range r.replicas {
		if rep.Healthy {
			return rep.DB
		}
	}
	return r.primary
}
