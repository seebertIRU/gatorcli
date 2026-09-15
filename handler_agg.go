package main

import (
	"context"
	"fmt"
	"time"
	
	"database/sql"
	
	"log"
	"strings"
	

	"gatorcli/internal/database"
	"github.com/google/uuid"
)

const minimumFeedInterval = 15 * time.Minute

func handlerAgg(s *state, cmd command) error {
	if len(cmd.args) != 1 {
		return fmt.Errorf("usage: agg <time_between_reqs>")
	}

	timeBetweenRequests, err := time.ParseDuration(cmd.args[0])
	if err != nil {
		return fmt.Errorf(
			"invalid duration %q: %w",
			cmd.args[0],
			err,
		)
	}

	if timeBetweenRequests < minimumFeedInterval {
		return fmt.Errorf(
			"time_between_reqs must be at least %s",
			minimumFeedInterval,
		)
	}

	fmt.Printf("Collecting feeds every %s\n", timeBetweenRequests)

	// Create the ticker before starting the loop.
	ticker := time.NewTicker(timeBetweenRequests)
	defer ticker.Stop()

	// Run immediately and then wait for each ticker event.
	for ; ; <-ticker.C {
		if err := scrapeFeeds(s); err != nil {
			fmt.Printf("Error scraping feed: %v\n", err)
		}
	}
}

func scrapeFeeds(s *state) error {
	ctx := context.Background()

	feed, err := s.db.GetNextFeedToFetch(ctx)
	if err != nil {
		return fmt.Errorf("couldn't get next feed to fetch: %w", err)
	}

	fmt.Printf("Fetching feed: %s\n", feed.Name)
	fmt.Printf("URL: %s\n", feed.Url)

	if err := s.db.MarkFeedFetched(ctx, feed.ID); err != nil {
		return fmt.Errorf(
			"couldn't mark feed %q as fetched: %w",
			feed.Name,
			err,
		)
	}

	rssFeed, err := fetchFeed(ctx, feed.Url)
	if err != nil {
		return fmt.Errorf(
			"couldn't fetch feed %q: %w",
			feed.Name,
			err,
		)
	}

	fmt.Printf("Posts from %s:\n", feed.Name)

	for _, item := range rssFeed.Channel.Item {
		fmt.Printf(" - %s\n", item.Title)
	}
	
	scrapeFeed(s.db, feed)

	return nil
}

func scrapeFeed(db *database.Queries, feed database.Feed) {
	err := db.MarkFeedFetched(context.Background(), feed.ID)
	if err != nil {
		log.Printf("Couldn't mark feed %s fetched: %v", feed.Name, err)
		return
	}

	feedData, err := fetchFeed(context.Background(), feed.Url)
	if err != nil {
		log.Printf("Couldn't collect feed %s: %v", feed.Name, err)
		return
	}
	for _, item := range feedData.Channel.Item {
		publishedAt := sql.NullTime{}
		if t, err := time.Parse(time.RFC1123Z, item.PubDate); err == nil {
			publishedAt = sql.NullTime{
				Time:  t,
				Valid: true,
			}
		}

		_, err = db.CreatePost(context.Background(), database.CreatePostParams{
			ID:        uuid.New(),
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
			FeedID:    feed.ID,
			Title:     item.Title,
			Description: sql.NullString{
				String: item.Description,
				Valid:  true,
			},
			Url:         item.Link,
			PublishedAt: publishedAt,
		})
		if err != nil {
			if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
				continue
			}
			log.Printf("Couldn't create post: %v", err)
			continue
		}
	}
	log.Printf("Feed %s collected, %v posts found", feed.Name, len(feedData.Channel.Item))
}
