package runner

import (
	"crypto/sha256"
	"encoding/base32"
	"encoding/json"
	"fmt"
	"newsfeed/v2/internal/log"
	"newsfeed/v2/internal/rss"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	timeFormat  = "2006-01-02_15:04:05"
	refreshRate = time.Minute * 60
)

type Runner struct{}

func NewRunner() *Runner {
	return &Runner{}
}

func (r *Runner) removeStale(dir string) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}

	for _, entry := range entries {
		parts := strings.Split(entry.Name(), ".")
		if len(parts) != 4 { // cache is in format time.hash.cache.json
			continue
		}

		cacheTime, err := time.Parse(timeFormat, parts[0])
		if err != nil {
			continue
		}

		if time.Since(cacheTime) > refreshRate {
			// casche is expired
			path := dir + "/" + entry.Name()
			if err = os.Remove(path); err != nil {
				log.Logf(log.DEBUG, "failed to remove stale cache '%s': %v\n", path, err)
			} else {
				log.Logf(log.DEBUG, "removing stale cache '%s'\n", path)
			}
		}
	}
}

func (r *Runner) Run(feeds []rss.Feed, isOneShot bool) (map[string][]rss.Article, error) {
	articles := make(map[string][]rss.Article, len(feeds))

	var mx sync.Mutex
	var wg sync.WaitGroup

	if tmpCacheDir != "" {
		r.removeStale(tmpCacheDir)
	}

	if cacheDir != "" {
		r.removeStale(cacheDir)
	}

	for idx := range feeds {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			log.Logf(log.DEBUG, "fetching articles from '%s'\n", feeds[i].Url)

			var err error
			var localArticles []rss.Article
			cacheMiss := false

			// first check tmp dir for ephemeral caching
			if tmpCacheDir != "" {
				localArticles, err = feeds[i].LoadFromDir(tmpCacheDir)
				if err != nil {
					log.Logf(log.DEBUG, "%v\n", err)
					cacheMiss = true
				} else {
					log.Logf(log.DEBUG, "loaded '%s' from cache\n", feeds[i].Name)
					cacheMiss = false
				}
			}

			// if we've identified a cache dir check that
			if cacheDir != "" && cacheMiss {
				localArticles, err = feeds[i].LoadFromDir(cacheDir)
				if err != nil {
					log.Logf(log.DEBUG, "%v\n", err)
					cacheMiss = true
				} else {
					log.Logf(log.DEBUG, "loaded '%s' from cache\n", feeds[i].Name)
					cacheMiss = false
				}
			}

			// failed to load from cache, have to get live feed
			if cacheMiss {
				log.Logf(log.DEBUG, "'%s' missed cache, fetching live feed\n", feeds[i].Name)

				localArticles, err = feeds[i].Fetch()
				if err != nil {
					log.Fatalf("[%d] %v\n", i, err)
				}

				dir := cacheDir
				if isOneShot {
					dir = tmpCacheDir
				}

				// update cache
				if dir != "" {
					err = r.saveToDir(feeds[i], localArticles, dir)
					if err != nil {
						log.Logf(log.ERROR, "%v\n", err)
					}
				}
			}

			mx.Lock()
			articles[feeds[i].Name] = localArticles
			mx.Unlock()
		}(idx)
	}
	wg.Wait()

	return articles, nil
}

func (r *Runner) saveToDir(feed rss.Feed, articles []rss.Article, dir string) error {
	if dir == "" {
		return fmt.Errorf("dir cannot be empty")
	}

	hash := sha256.Sum224([]byte(feed.Url))
	strHash := strings.TrimRight(base32.HexEncoding.EncodeToString(hash[:]), "=")

	path := fmt.Sprintf("%s/%s.%s.cache.json", dir, time.Now().UTC().Format(timeFormat), strHash)
	bytes, err := json.Marshal(articles)
	if err != nil {
		return err
	}

	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0o755)
	if err != nil {
		return err
	}

	_, err = file.Write(bytes)
	return err
}
