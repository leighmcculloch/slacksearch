package main // import "4d63.com/slacksearch"

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

func main() {
	slackToken := os.Getenv("SLACK_TOKEN")
	if slackToken == "" {
		fmt.Fprintf(os.Stderr, "Environment variable SLACK_TOKEN must be set.\n")
		os.Exit(1)
	}

	if len(os.Args) <= 1 {
		fmt.Fprintf(os.Stderr, "Usage: SLACK_TOKEN=xoxp-... slacksearch <search query> \n")
		return
	}

	words := os.Args[1:]
	query := strings.Join(words, " ")

	results, err := search(slackToken, query)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error searching with query %q.\n", query)
		fmt.Fprintf(os.Stderr, "%#v\n", err)
		os.Exit(1)
	}

	if len(results) == 0 {
		fmt.Fprintf(os.Stderr, "No results.\n")
		return
	}

	for _, r := range results {
		fmt.Fprintf(os.Stdout, "%s #%s @%s %s %s\n", r.Timestamp.Format("2006-01-02"), r.Channel, r.Username, r.Text, r.Permalink)
	}
}

type result struct {
	Timestamp time.Time
	Channel   string
	Username  string
	Text      string
	Permalink string
}

func search(slackToken, query string) ([]result, error) {
	q := url.Values{}
	q.Set("token", slackToken)
	q.Set("query", query)
	q.Set("sort", "timestamp")
	q.Set("sort_dir", "desc")

	searchURL := "https://slack.com/api/search.all?" + q.Encode()

	res, err := http.Get(searchURL)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var resDecoded struct {
		Messages struct {
			Matches []struct {
				Timestamp string `json:"ts"`
				Channel   struct {
					Name string `json:"name"`
				} `json:"channel"`
				Username  string `json:"username"`
				Text      string `json:"text"`
				Permalink string `json:"permalink"`
			} `json:"matches"`
		} `json:"messages"`
	}

	err = json.NewDecoder(res.Body).Decode(&resDecoded)
	if err != nil {
		return nil, err
	}

	resultCount := len(resDecoded.Messages.Matches)
	results := make([]result, resultCount)
	for i := 0; i < resultCount; i++ {
		results[i] = result{
			Timestamp: time.Now(),
			Channel:   resDecoded.Messages.Matches[i].Channel.Name,
			Username:  resDecoded.Messages.Matches[i].Username,
			Text:      strings.Replace(resDecoded.Messages.Matches[i].Text, "\n", " ", -1),
			Permalink: resDecoded.Messages.Matches[i].Permalink,
		}
	}

	return results, nil
}
