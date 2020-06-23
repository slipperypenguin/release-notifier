package main

import (
	"context"
	"fmt"
	githubql "github.com/shurcooL/githubql"
	"go.uber.org/zap"
	"strings"
	"time"
)

// Checker has a githubql client to run queries and also knows about
// the current repositories releases to compare against.
type Checker struct {
	logger   zap.SugaredLogger
	client   *githubql.Client
	releases map[string]Repository
}

// Run the queries and comparisons for the given repositories in a given interval
func (c *Checker) Run(interval time.Duration, repositories []string, releases chan<- Repository) {
	if c.releases == nil {
		c.releases = make(map[string]Repository)
	}

	for {
		for _, repoName := range repositories {
			s := strings.Split(repoName, "/")
			owner, name := s[0], s[1]

			nextRepo, err := c.query(owner, name)
			if err != nil {
				// log.Printf("failed to query the repository's releases. owner: %s, name: %s, err: %v", owner, name, err)
				c.logger.Warnw("failed to query the repository's releases.",
					"owner", owner,
					"name", name,
					"err", err,
				)
				continue
			}

			currRepo, ok := c.releases[repoName]

			// we've queried the repo for the first time. save the current state to compare with next iteration
			if !ok {
				c.releases[repoName] = nextRepo
				continue
			}

			if nextRepo.Release.PublishedAt.After(currRepo.Release.PublishedAt) {
				releases <- nextRepo
				c.releases[repoName] = nextRepo
			} else {
				//log.Printf("no new releases for this repo. %s/%s", owner, name)
				c.logger.Infow("no new releases for this repo.",
					"owner", owner,
					"name", name,
				)
			}
		}
		time.Sleep(interval)
	}
}

func (c *Checker) query(owner, name string) (Repository, error) {
	var query struct {
		Repository struct {
			ID          githubql.ID
			Name        githubql.String
			Description githubql.String
			URL         githubql.URI

			Releases struct {
				Edges []struct {
					Node struct {
						ID          githubql.ID
						Name        githubql.String
						Description githubql.String
						URL         githubql.URI
						PublishedAt githubql.DateTime
					}
				}
			} `graphql:"releases(last: 1)"`
		} `graphql:"repository(owner: $owner, name: $name)"`
	}

	vars := map[string]interface{}{
		"owner": githubql.String(owner),
		"name":  githubql.String(name),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := c.client.Query(ctx, &query, vars); err != nil {
		return Repository{}, err
	}

	repoID, ok := query.Repository.ID.(string)
	if !ok {
		return Repository{}, fmt.Errorf("can't convert repo id to string: %v", query.Repository.ID)
	}

	if len(query.Repository.Releases.Edges) == 0 {
		return Repository{}, fmt.Errorf("can't find any releases for %s/%s", owner, name)
	}
	latestRelease := query.Repository.Releases.Edges[0].Node

	releaseID, ok := latestRelease.ID.(string)
	if !ok {
		return Repository{}, fmt.Errorf("can't convert release id to string: %v", query.Repository.ID)
	}

	return Repository{
		ID:          repoID,
		Name:        string(query.Repository.Name),
		Owner:       owner,
		Description: string(query.Repository.Description),
		URL:         *query.Repository.URL.URL,

		Release: Release{
			ID:          releaseID,
			Name:        string(latestRelease.Name),
			Description: string(latestRelease.Description),
			URL:         *latestRelease.URL.URL,
			PublishedAt: latestRelease.PublishedAt.Time,
		},
	}, nil
}
