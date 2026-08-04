package runner

import "os"

var (
	cacheDir    = establishCacheDir()
	tmpCacheDir = establishTmpCacheDir()
)

// Instantiate directory for caching rss results
func establishCacheDir() string {
	dir, err := os.UserCacheDir()
	if err == nil {
		dir += "/newsfeed"
		err = os.MkdirAll(dir, 0o700)
		if err == nil {
			return dir
		}
	}

	return ""
}

func establishTmpCacheDir() string {
	dir := os.TempDir() + "/newsfeed"
	err := os.MkdirAll(dir, 0o700)
	if err != nil {
		return ""
	}

	return dir
}
