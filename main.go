package main

import (
	"fmt"
	"maps"
	"net/url"
	"newsfeed/v2/internal/browser"
	"newsfeed/v2/internal/cli"
	"newsfeed/v2/internal/dmenu"
	"newsfeed/v2/internal/log"
	"newsfeed/v2/internal/rss"
	"newsfeed/v2/internal/runner"
	"os"
	"slices"
	"strings"
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
		parts := strings.SplitN(urls[i], "=", 2)
		if len(parts) != 2 {
			u, err := url.Parse(urls[i])
			if err != nil {
				log.Fatalln(err)
			}

			feeds = append(feeds, rss.Feed{
				Name: u.Hostname(),
				Url:  urls[i],
			})
		} else {
			feeds = append(feeds, rss.Feed{
				Name: parts[0],
				Url:  parts[1],
			})
		}
	}

	r := runner.NewRunner()
	articles, err := r.Run(feeds, true)
	if err != nil {
		log.Fatalf("%v\n", err)
	}

	if args.Disaplay {
		var out strings.Builder

		for k, feed := range articles {
			size := len(feed) * min(args.TitleLength, 256) // limit to 256 cap for prealloc
			if args.WriteNames {
				size += (len(k) + 2) * len(feed)
			}
			out.Grow(size)

			for i := range min(len(feed), args.PerFeed) {
				if args.WriteNames {
					fmt.Fprintf(&out, "%s: %s\n", k, feed[i].Title)
				} else {
					fmt.Fprintln(&out, feed[i].Title)
				}
			}
		}

		fmt.Print(strings.TrimSpace(out.String()))
	}

	if args.UseDmenu {
		tmp := slices.Collect(maps.Values(articles))
		a := make([]rss.Article, len(tmp[0])*len(tmp))

		for _, feed := range tmp {
			a = append(a, feed...)
		}

		dconf := dmenu.NewConfig("")
		url, err := dmenu.OpenMenu(a, dconf)
		if err != nil {
			log.Fatalln(err)
		}

		bconf := browser.NewConfig("")
		err = browser.StartBrowser(url, bconf)
		if err != nil {
			log.Fatalln(err)
		}
	}
}
