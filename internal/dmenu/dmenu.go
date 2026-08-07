package dmenu

import (
	"bytes"
	"encoding/json"
	"newsfeed/v2/internal/log"
	"newsfeed/v2/internal/rss"
	"os"
	"os/exec"
	"slices"
	"strings"
)

const (
	ConfigPath = "/newsfeed/dmenu.conf"
)

type Config struct {
	Font               string `json:"Font"`
	Prompt             string `json:"Prompt"`
	NormalBackground   string `json:"NormalBackground"`
	NormalForeground   string `json:"NormalForeground"`
	SelectedBackground string `json:"SelectedBackground"`
	SelectedForeground string `json:"SelectedForeground"`
	Lines              string `json:"Lines"`
}

func NewConfig(path string) *Config {
	c := &Config{
		Font:               "DejaVu Sans Mono:size=10",
		Prompt:             "NEWS FEED",
		NormalBackground:   "#111111",
		NormalForeground:   "#bbbbbb",
		SelectedBackground: "#77d424",
		SelectedForeground: "#000000",
		Lines:              "20",
	}

	if path == "" {
		return c
	}

	bytes, err := os.ReadFile(path)
	if err != nil {
		log.Fatalln(err)
	}

	err = json.Unmarshal(bytes, c)
	if err != nil {
		log.Fatalln(err)
	}
	return c
}

// Returns the url of the selected article
func OpenMenu(articles []rss.Article, c *Config) (string, error) {
	cmd := exec.Command(
		"dmenu",
		"-i",
		"-p", c.Prompt,
		"-fn", c.Font,
		"-nb", c.NormalBackground,
		"-nf", c.NormalForeground,
		"-sb", c.SelectedBackground,
		"-sf", c.SelectedForeground,
		"-l", c.Lines,
	)
	if cmd == nil {
		log.Fatalln("command is nil")
	}

	options := make([]string, 0, len(articles))
	for _, v := range articles {
		options = append(options, v.Title)
	}

	input := strings.Join(options, "\n")
	cmd.Stdin = strings.NewReader(input)

	out := bytes.NewBuffer(make([]byte, 0, 32))
	cmd.Stdout = out

	err := cmd.Run()
	if err != nil {
		return "", err
	}

	// get output and remove trailing newline
	choice := out.String()
	idx := slices.Index(options, choice)
	if idx == -1 {
		panic("this is really bad")
	}

	return articles[idx].Url, nil
}
