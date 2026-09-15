package main

import (
	"context"
	"encoding/xml"
	"fmt"
	"html"
	"io"
	"net/http"
)

// RSSFeed represents the root RSS document.
type RSSFeed struct {
	Channel RSSChannel `xml:"channel"`
}

// RSSChannel represents the channel element inside an RSS feed.
type RSSChannel struct {
	Title       string    `xml:"title"`
	Link        string    `xml:"link"`
	Description string    `xml:"description"`
	Item        []RSSItem `xml:"item"`
}

// RSSItem represents an individual item in an RSS channel.
type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

// fetchFeed downloads an RSS feed and converts the XML response
// into an RSSFeed struct.
func fetchFeed(ctx context.Context, feedURL string) (*RSSFeed, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		feedURL,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("create feed request: %w", err)
	}

	// Identify this application to the RSS server.
	req.Header.Set("User-Agent", "gator")

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch feed: %w", err)
	}
	defer resp.Body.Close()

	// Client.Do does not treat HTTP error status codes such as
	// 404 or 500 as Go errors, so check the status explicitly.
	if resp.StatusCode < http.StatusOK ||
		resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf(
			"fetch feed: server returned %s",
			resp.Status,
		)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read feed response: %w", err)
	}

	var feed RSSFeed

	err = xml.Unmarshal(data, &feed)
	if err != nil {
		return nil, fmt.Errorf("parse feed XML: %w", err)
	}

	unescapeFeed(&feed)

	return &feed, nil
}

// unescapeFeed converts HTML entities such as &amp; and &ldquo;
// into their corresponding characters.
func unescapeFeed(feed *RSSFeed) {
	feed.Channel.Title =
		html.UnescapeString(feed.Channel.Title)

	feed.Channel.Description =
		html.UnescapeString(feed.Channel.Description)

	for i := range feed.Channel.Item {
		feed.Channel.Item[i].Title =
			html.UnescapeString(feed.Channel.Item[i].Title)

		feed.Channel.Item[i].Description =
			html.UnescapeString(feed.Channel.Item[i].Description)
	}
}
