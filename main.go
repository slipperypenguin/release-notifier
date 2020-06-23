package main

import (
	"context"
	"fmt"
	"flag"
	"github.com/joho/godotenv"
	githubql "github.com/shurcooL/githubql"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
	"log"
	"time"
)

// Config of env and args
type Config struct {
	GithubToken     string        `arg:"env:GITHUB_TOKEN"`
	Interval        time.Duration `arg:"env:INTERVAL"`
	LogLevel        string        `arg:"env:LOG_LEVEL"`
	Repositories    []string      `arg:"-r,separate"`
	SlackHook       string        `arg:"env:SLACK_HOOK"`
	IgnoreNonstable bool          `arg:"env:IGNORE_NONSTABLE"`
}

// Token returns an oauth2 token or an error.
func (c Config) Token() *oauth2.Token {
	return &oauth2.Token{AccessToken: c.GithubToken}
}

// init is invoked before main()
func init() {
	// loads values from .env into the system
	if err := godotenv.Load(); err != nil {
		fmt.Println("No .env file found...")
	}
}

func main() {
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	// default is to watch kubernetes/kubernetes repo
	pflag.StringP("repo", "r", "kubernetes/kubernetes", "repository to check")
	pflag.Parse()
	if err := viper.BindPFlags(pflag.CommandLine); err != nil {
		panic(err)
	}

	c := Config{
		Interval: time.Hour,
		LogLevel: "info",
	}

	logger, _ := zap.NewProduction()
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
			// log.Printf("not notifying about non-stable version: %v", repo.Release.Name)
			sugar.Infow("not notifying about non-stable version.",
				"version", repo.Release.Name,
			)
			continue
		}
		if err := slack.Send(repo); err != nil {
			// log.Printf("failed to send release to messenger. %+v", err)
			sugar.Errorf("failed to send release to messenger. %+v", err)
			sugar.Warnw(
				"failed to send release to messenger.",
				"err", err,
			)
			continue
		}
	}
}
