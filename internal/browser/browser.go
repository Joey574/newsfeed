package browser

import (
	"os/exec"
)

const (
	ConfigPath = "/newsfeed/browser.conf"
)

type Config struct {
}

func NewConfig(path string) *Config {
	return &Config{}
}

func StartBrowser(url string, c *Config) error {
	cmd := exec.Command("firefox", "--new-window", url)
	return cmd.Start()
}
