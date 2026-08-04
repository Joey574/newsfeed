package cli

import (
	"math"

	"github.com/jessevdk/go-flags"
)

type Args struct {
	ConfigPath  string `short:"c" long:"config" description:"specify a path to a config file"`
	UseDmenu    bool   `long:"dmenu" description:"open dmenu to browse all articles"`
	WriteNames  bool   `long:"names" description:"write source names along with titles"`
	Disaplay    bool   `long:"display" description:"write articles to stdout"`
	Quiet       bool   `short:"q" long:"quiet" description:"disable logging"`
	Verbose     bool   `short:"v" long:"verbose" description:"verbose logging"`
	PerFeed     int    `long:"per-feed" default:"-1" description:"number of articles to display per feed, -1 disables limit"`
	TitleLength int    `long:"title-length" default:"-1" description:"truncates titles longer than this, -1 disables truncation"`
}

func NewArgs() *Args {
	return &Args{}
}

func (a *Args) Parse() ([]string, error) {
	pos, err := flags.Parse(a)
	if flags.WroteHelp(err) {
		return nil, nil
	}

	if a.PerFeed == -1 {
		a.PerFeed = math.MaxInt
	}

	if a.TitleLength == -1 {
		a.TitleLength = math.MaxInt
	}

	return pos, err
}
