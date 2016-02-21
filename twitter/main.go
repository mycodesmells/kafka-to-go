package main

import (
	"fmt"
	"net/http"

	"github.com/mycodesmells/kafka-to-go/twitter/twitter"

	"github.com/kurrik/oauth1a"
	"github.com/kurrik/twittergo"
	"time"
)

func main() {
	var (
		lastTweetID uint64
		handle      = "mycodesmells"
	)

	client := connect()

	for {
		select {
		case <-time.After(time.Second):
			tl, err := getTweets(client, handle, lastTweetID)
			if err != nil {
				fmt.Printf("Failed to load Tweets from %s: %v", handle, err)
			}

			lastID := printTweets(tl)
			if lastID != 0 {
				lastTweetID = lastID
			}
		}
	}

}

func connect() *twittergo.Client {
	config := &oauth1a.ClientConfig{
		ConsumerKey:    twitter.Key,
		ConsumerSecret: twitter.Secret,
	}
	return twittergo.NewClient(config, nil)
}

func getTweets(client *twittergo.Client, handle string, lastTweetID uint64) (twittergo.Timeline, error) {
	url := getURL(handle, lastTweetID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to build service request: %v", err)
	}

	resp, err := client.SendRequest(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send API request: %v", err)
	}

	timeline := &twittergo.Timeline{}

	if err = resp.Parse(timeline); err != nil {
		return nil, fmt.Errorf("failed to parse API response: %v", err)
	}

	return *timeline, nil
}

func getURL(handle string, lastTweetID uint64) string {
	baseURL := fmt.Sprintf("https://api.twitter.com/1.1/statuses/user_timeline.json?screen_name=%s", handle)
	if lastTweetID == 0 {
		return baseURL
	}
	return fmt.Sprintf("%s&since_id=%d", baseURL, lastTweetID)
}

func printTweets(tl twittergo.Timeline) uint64 {
	fmt.Printf("Tweets found: %d\n", len(tl))
	for _, t := range tl {
		fmt.Printf("#%d: %s\n", t.Id(), t.Text())
	}

	if len(tl) > 0 {
		return tl[0].Id()
	}

	return 0
}
