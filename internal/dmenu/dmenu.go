package dmenu

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
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

	bytes, err := os.ReadFile(path)
	if err != nil {
		return c
	}

	_ = json.Unmarshal(bytes, c)
	return c
}

func OpenMenu(options []string, c *Config) (string, error) {
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
	return choice[:len(choice)-1], nil
}
