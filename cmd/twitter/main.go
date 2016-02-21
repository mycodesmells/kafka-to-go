package main

import (
	"fmt"
	"time"

	"github.com/mycodesmells/kafka-to-go/twitter"
)

func main() {
	var (
		lastTweetID uint64
		handle      = "mycodesmells"
	)

	client := twitter.Connect()

	for {
		select {
		case <-time.After(time.Second):
			tl, err := twitter.Tweets(client, handle, lastTweetID)
			if err != nil {
				fmt.Printf("Failed to load Tweets from %s: %v", handle, err)
			}

			twitter.PrintTweets(tl)
			if len(tl) > 0 {
				lastTweetID = tl[0].Id()
			}
		}
	}

}
