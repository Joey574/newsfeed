package main

import (
	"fmt"
	"newsfeed/v2/internal/cli"
	"newsfeed/v2/internal/log"
	"newsfeed/v2/internal/rss"
	"newsfeed/v2/internal/runner"
	"os"
)

const (
	DefaultConfigDir = "/etc/newsfeed"
)

func getConfigPath(name string) string {
	localConfDir, err := os.UserConfigDir()
	if err != nil {
		return ""
	}

	if _, err := os.Stat(localConfDir + name); err == nil {
		return localConfDir + name
	}

	if _, err := os.Stat(DefaultConfigDir + name); err == nil {
		return DefaultConfigDir + name
	}

	return ""
}

type Config struct {
	RSSPath         string
	DmenuPath       string
	BrowserPath     string
	ElementsPerFeed int
	MaxTitleLength  int
	OpenMenu        bool
	Names           bool
	Quiet           bool
}

func main() {
	args := cli.NewArgs()

	urls, err := args.Parse()
	if err != nil {
		log.Fatalln(err)
	}

	log.SetQuiet(args.Quiet)
	log.SetVerbose(args.Verbose)

	var feeds []rss.Feed
	for i := range urls {
		feeds = append(feeds, rss.Feed{
			Name: "",
			Url:  urls[i],
		})
	}

	r := runner.NewRunner()
	articles, err := r.Run(feeds, true)
	if err != nil {
		log.Fatalf("%v\n", err)
	}

	if args.Disaplay {
		for k, v := range articles {
			for _, a := range v {
				if args.WriteNames {
					fmt.Printf("%s: %s\n", k, a.Title)
				} else {
					fmt.Println(a.Title)
				}
			}
		}
	}
}
