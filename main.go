package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/joho/godotenv"
	githubql "github.com/shurcooL/githubql"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
	"log"
	"os"
	"time"
)

// Config of env and args
type Config struct {
	GithubToken     string        `arg:"env:GITHUB_TOKEN"`
	Interval        time.Duration `arg:"env:INTERVAL"`
	Repositories    []string      `arg:"-r,separate"`
	SlackHook       string        `arg:"env:SLACK_HOOK"`
	IgnoreNonstable bool          `arg:"env:IGNORE_NONSTABLE"`
}

// Token returns an oauth2 token or an error.
func (c Config) Token() *oauth2.Token {
	return &oauth2.Token{AccessToken: c.GithubToken}
}

func main() {
	_ = godotenv.Load()

	c := Config{
		Interval:        time.Hour,
		IgnoreNonstable: true,
	}

	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.StringSliceP("repo", "r", []string{""}, "repository to check")
	pflag.Parse()
	if err := viper.BindPFlags(pflag.CommandLine); err != nil {
		panic(err)
	}
	c.Repositories = viper.GetStringSlice("repo")
	if len(c.Repositories) == 0 {
		log.Fatalf("no repositories to watch")
	}

	ghtok, exists := os.LookupEnv("GITHUB_TOKEN")
	if exists {
		c.GithubToken = ghtok
		// c.Interval = time.Minute
		fmt.Println("github credential loaded")
	}
	shook, exists := os.LookupEnv("SLACK_HOOK")
	if exists {
		c.SlackHook = shook
		fmt.Println("slack credential loaded")
	}

	logger, _ := zap.NewDevelopment()
	defer logger.Sync()
	sugar := logger.Sugar()

	if len(c.Repositories) == 0 {
		log.Fatalf("no repositories to watch")
	}

	tokenSource := oauth2.StaticTokenSource(c.Token())
	client := oauth2.NewClient(context.Background(), tokenSource)
	checker := &Checker{
		logger: *sugar,
		client: githubql.NewClient(client),
	}

	releases := make(chan Repository)
	go checker.Run(c.Interval, c.Repositories, releases)

	slack := SlackSender{Hook: c.SlackHook}

	sugar.Infow("waiting for new releases")
	for repo := range releases {
		if c.IgnoreNonstable && repo.Release.IsNonstable() {
			sugar.Infow("not notifying about non-stable version.",
				"version", repo.Release.Name,
			)
			continue
		}
		if err := slack.Send(repo); err != nil {
			sugar.Errorf("failed to send release to messenger. %+v", err)
			sugar.Warnw(
				"failed to send release to messenger.",
				"err", err,
			)
			continue
		}
	}
}
