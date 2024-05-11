package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/ViBiOh/flags"
	"github.com/ViBiOh/ketchup/pkg/provider/github"
)

func main() {
	fs := flag.NewFlagSet("checker", flag.ExitOnError)
	fs.Usage = flags.Usage(fs)

	githubConfig := github.Flags(fs, "github")

	_ = fs.Parse(os.Args[1:])

	ctx := context.Background()

	githubApp := github.New(githubConfig, nil, nil, nil)

	patterns, err := githubApp.LatestVersions(ctx, "k3s-io/k3s", []string{"stable", "latest"})
	fmt.Println(patterns, err)
}
