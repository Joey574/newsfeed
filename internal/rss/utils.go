package rss

import (
	"fmt"
	"os"
)

var cacheDir = establishCacheDir()

func establishCacheDir() string {
	cacheDir, err := os.UserCacheDir()
	if err == nil {
		cacheDir += "/newsfeed"
		err = os.MkdirAll(cacheDir, 0755)
		if err == nil {
			return cacheDir
		}
	}

	return ""
}

func truncate(s string, max int) string {
	if len(s) < max {
		return s
	}

	return fmt.Sprintf("%.*s...", max, s)
}
