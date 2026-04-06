package cli

import (
	"io"
	"os"

	"github.com/vyper-tooling/pi/internal/cache"
)

// PackageFetcher abstracts GitHub package cache lookup and fetching.
// Implementations handle the cache-check → fetch → return-path lifecycle.
// The interface decouples CLI commands from the cache package, enabling
// unit testing without filesystem or network access.
type PackageFetcher interface {
	// Fetch ensures the package at org/repo@version is available locally.
	// Returns the local path, whether it was already cached, and any error.
	Fetch(org, repo, version string) (path string, wasCached bool, err error)
}

// CachePackageFetcher wraps cache.Cache to implement PackageFetcher.
type CachePackageFetcher struct {
	cache *cache.Cache
}

// NewCachePackageFetcher creates a PackageFetcher backed by the default
// ~/.pi/cache directory.
func NewCachePackageFetcher(stderr io.Writer) (*CachePackageFetcher, error) {
	root, err := cache.DefaultCacheRoot()
	if err != nil {
		return nil, err
	}
	return &CachePackageFetcher{
		cache: &cache.Cache{
			Root:       root,
			WarnWriter: stderr,
			PIVersion:  version,
		},
	}, nil
}

func (f *CachePackageFetcher) Fetch(org, repo, version string) (string, bool, error) {
	cachePath := f.cache.PackagePath(org, repo, version)
	if info, err := os.Stat(cachePath); err == nil && info.IsDir() {
		return cachePath, true, nil
	}

	path, err := f.cache.Fetch(org, repo, version)
	if err != nil {
		return "", false, err
	}
	return path, false, nil
}
