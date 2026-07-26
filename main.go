package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"newsfeed/v2/internal/browser"
	"newsfeed/v2/internal/dmenu"
	"newsfeed/v2/internal/log"
	"newsfeed/v2/internal/rss"
	"os"
	"strings"
	"sync"
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
	var c Config

	flag.StringVar(&c.RSSPath, "rss", getConfigPath(rss.ConfigPath), "json file with rss config data")
	flag.StringVar(&c.DmenuPath, "dmenu", getConfigPath(dmenu.ConfigPath), "json file with dmenu config data")
	flag.StringVar(&c.BrowserPath, "browser", getConfigPath(browser.ConfigPath), "json file with browser config data")
	flag.BoolVar(&c.OpenMenu, "menu", false, "open dmenu to browse articles")
	flag.BoolVar(&c.Names, "names", false, "write source names along with titles")
	flag.BoolVar(&c.Quiet, "q", false, "disable logging")
	flag.IntVar(&c.ElementsPerFeed, "l", -1, "elements per feed to display, -1 disables limit")
	flag.IntVar(&c.MaxTitleLength, "c", -1, "max title length, -1 disables limit")
	flag.Parse()
	log.SetQuiet(c.Quiet)

	if c.ElementsPerFeed == -1 {
		c.ElementsPerFeed = math.MaxInt
	}

	if c.MaxTitleLength == -1 {
		c.MaxTitleLength = math.MaxInt
	}

	var feeds []rss.Feed
	urls := flag.Args()
	if len(urls) == 0 {
		// if we haven't been given any urls, try and parse them from the rss config
		bytes, err := os.ReadFile(c.RSSPath)
		if err != nil {
			log.Fatalln(err)
		}

		if err := json.Unmarshal(bytes, &feeds); err != nil {
			log.Fatalln(err)
		}
	} else {
		for i := range urls {
			feeds = append(feeds, rss.Feed{
				Name: "",
				Url:  urls[i],
			})
		}
	}

	var wg sync.WaitGroup
	for i := range feeds {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			err := feeds[i].GetArticles()
			if err != nil {
				log.Println(err)
			}
		}(i)
	}
	wg.Wait()

	totalArticles := 0
	for i := range feeds {
		totalArticles += len(feeds[i].Articles)
	}

	choices := make([]string, 0, totalArticles)
	if c.OpenMenu {
		for i := range feeds {
			choices = append(choices, feeds[i].FormatArticles(c.Names, c.ElementsPerFeed, c.MaxTitleLength)...)
		}

		// dmenu and browser both have default configs if one isn't provided
		dmenuConf := dmenu.NewConfig(c.DmenuPath)
		browserConf := browser.NewConfig(c.BrowserPath)

		choice, err := dmenu.OpenMenu(choices, dmenuConf)
		if err != nil {
			log.Fatalln(err)
		}

		choice = choice[strings.Index(choice, "] ")+2:]
		for i := range feeds {
			for j := range feeds[i].Articles {
				if feeds[i].Articles[j].Title == choice {
					browser.StartBrowser(feeds[i].Articles[j].Url, browserConf)
					return
				}
			}
		}
	} else {
		for i := range feeds {
			choices = append(choices, feeds[i].FormatArticles(c.Names, c.ElementsPerFeed, c.MaxTitleLength)...)
		}

		fmt.Println(strings.Join(choices, "\n"))
	}
}
