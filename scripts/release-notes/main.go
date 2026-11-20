// release-notes 命令从 changelog 目录生成正式发行正文。
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/AlonaBot/AlonaBot/internal/releasenotes"
)

func main() {
	version := flag.String("version", "", "release version (required)")
	flag.Parse()
	notes, err := releasenotes.Render("changelog", "v"+strings.TrimPrefix(*version, "v"))
	if err == nil {
		_, err = os.Stdout.Write(notes)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
