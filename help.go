package main

import (
	"fmt"
	"os"

	"github.com/hoisie/mustache"
)

const (
	name    = "gwatch"
	version = "0.1.0"
	usage   = "A cross-platform version of the linux tool watch."
	author  = "Austin Poor"
	license = "MIT"
	repo    = "https://github.com/a-poor/gwatch"
)

const customUsage = `NAME:
    {{ name }} - {{ usage }}

USAGE:
    $ {{ name }} [FLAGS...] COMMAND

FLAGS:
    -i  Interval; How often should the command be run? (default 2s)
    -h  Display the help message (and exit)

ABOUT:
    Author:  {{ author }}
    License: {{ license }}
    Source:  {{ repo }}
    Bugs:    {{ repo }}/issues

`

func CustomUsage() {
	usage := mustache.Render(customUsage, map[string]interface{}{
		"name":    name,
		"version": version,
		"usage":   usage,
		"author":  author,
		"license": license,
		"repo":    repo,
	})
	_, err := fmt.Fprint(os.Stderr, usage)
	if err != nil {
		panic(err)
	}
}
