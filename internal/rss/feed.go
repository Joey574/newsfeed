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

type Feed struct {
	Name     string    `json:"Name,omitempty"`
	Url      string    `json:"Url,omitempty"`
	Articles []Article `json:"Articles,omitempty"`
}

func (f *Feed) GetArticles() error {
	err := f.loadFromCache(cacheDir)
	if err != nil {
		err = f.fetchLiveFeed()
		if err != nil {
			return err
		}

		err = f.save(cacheDir)
	}

	return err
}

// Returns a base32 hash of the feed url
func (f *Feed) hash() string {
	hash := sha256.Sum224([]byte(f.Url))
	return strings.TrimRight(base32.HexEncoding.EncodeToString(hash[:]), "=")
}

// Fetches the current rss feed
func (f *Feed) fetchLiveFeed() error {
	fp := gofeed.NewParser()
	fp.UserAgent = userAgent

	gf, err := fp.ParseURL(f.Url)
	if err != nil {
		return err
	}

	f.Articles = make([]Article, len(gf.Items))
	for i := range gf.Items {
		f.Articles[i] = Article{
			Title: gf.Items[i].Title,
			Url:   gf.Items[i].Link,
		}
	}

	return nil
}

// Returns nil if and only if a cache file was succesfully loaded
func (f *Feed) loadFromCache(dir string) error {
	if dir == "" {
		return fmt.Errorf("invalid directory")
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		path := dir + "/" + entry.Name()
		if f.isValidCache(path) {
			return f.load(path)
		}
	}

	return fmt.Errorf("no cache found")
}

// Returns true if the cache is for our url and not stale
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

	// delete stale cache
	os.Remove(path)
	return false
}

func (f *Feed) load(path string) error {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	return json.Unmarshal(bytes, f)
}

func (f *Feed) save(dir string) error {
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

func (f *Feed) FormatArticles(names bool, list, cap int) []string {
	choices := make([]string, 0, len(f.Articles))
	if names {
		for i := 0; i < len(f.Articles) && i < list; i++ {
			choices = append(choices, fmt.Sprintf("[%s] %s", f.Name, truncate(f.Articles[i].Title, cap)))
		}
	} else {
		for i := 0; i < len(f.Articles) && i < list; i++ {
			choices = append(choices, fmt.Sprintf("%d. %s", i+1, truncate(f.Articles[i].Title, cap)))
		}
	}

	return choices
}
