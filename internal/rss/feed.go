package rss

import (
	"crypto/sha256"
	"encoding/base32"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mmcdole/gofeed"
)

const (
	ConfigPath = "/newsfeed/rss.conf"
	TimeFormat = "2006-01-02_15:04:05"
)

var UserAgents = []string{
	"curl/8.21.0", // appear curl-like

	// immitate windows
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/128.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:130.0) Gecko/20100101 Firefox/130.0",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/127.0.0.0 Safari/537.36 Edg/127.0.0.0",

	// immitate maxOS
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 14_6_1) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.6 Safari/605.1.15",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 14_6_1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/128.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/127.0.0.0 Safari/537.36",

	// immitate ios
	"Mozilla/5.0 (iPhone; CPU iPhone OS 17_6_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.6 Mobile/15E148 Safari/604.1",
	"Mozilla/5.0 (iPad; CPU OS 17_6_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.6 Mobile/15E148 Safari/604.1",
	"Mozilla/5.0 (iPhone; CPU iPhone OS 16_7_10 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.6 Mobile/15E148 Safari/604.1",

	// immitate android
	"Mozilla/5.0 (Linux; Android 14; SM-S928B) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/128.0.6613.88 Mobile Safari/537.36",
	"Mozilla/5.0 (Linux; Android 14; Pixel 8 Pro) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/127.0.6533.103 Mobile Safari/537.36",
	"Mozilla/5.0 (Linux; Android 13; SM-G991B) AppleWebKit/537.36 (KHTML, like Gecko) SamsungBrowser/26.0 Chrome/122.0.0.0 Mobile Safari/537.36",
}

type Feed struct {
	Name string `json:"Name,omitempty"`
	Url  string `json:"Url,omitempty"`
}

// Returns a base32 hash of the feed url
func (f *Feed) hash() string {
	hash := sha256.Sum224([]byte(f.Url))
	return strings.TrimRight(base32.HexEncoding.EncodeToString(hash[:]), "=")
}

// Fetches the current rss feed
func (f *Feed) Fetch() ([]Article, error) {
	fp := gofeed.NewParser()

	var err error
	var gf *gofeed.Feed
	for _, ua := range UserAgents {
		fp.UserAgent = ua
		gf, err = fp.ParseURL(f.Url)
		if err == nil {
			articles := make([]Article, len(gf.Items))
			for i := range gf.Items {
				articles[i] = Article{
					Title: gf.Items[i].Title,
					Url:   gf.Items[i].Link,
				}
			}

			return articles, nil
		}
	}

	return nil, err
}

// Returns nil if and only if a valid cache file was found and loaded
func (f *Feed) LoadFromDir(dir string) ([]Article, error) {
	if dir == "" {
		return nil, fmt.Errorf("invalid directory")
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		path := dir + "/" + entry.Name()
		if f.isValidCache(path) {
			return f.load(path)
		}
	}

	return nil, fmt.Errorf("no cache found in '%s'", dir)
}

// Returns true if the cache is for our url and is not stale
func (f *Feed) isValidCache(path string) bool {
	name := filepath.Base(path)
	parts := strings.Split(name, ".")
	if len(parts) != 4 || parts[1] != f.hash() {
		return false
	}

	cacheTime, err := time.Parse(TimeFormat, parts[0])
	if err != nil {
		return false
	}

	if time.Since(cacheTime) < RefreshRate {
		return true
	}

	return false
}

func (f *Feed) load(path string) ([]Article, error) {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var articles []Article
	err = json.Unmarshal(bytes, &articles)
	if err != nil {
		return nil, err
	}

	return articles, nil
}

func (f *Feed) SaveToDir(dir string) error {
	if dir == "" {
		return nil
	}

	path := fmt.Sprintf("%s/%s.%s.cache.json", dir, time.Now().UTC().Format(TimeFormat), f.hash())
	bytes, err := json.Marshal(f)
	if err != nil {
		return err
	}

	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0755)
	if err != nil {
		return err
	}

	_, err = file.Write(bytes)
	return err
}
