package main

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gatorcli/internal/database"
)

// handlerAddFeed creates a feed owned by the currently configured user.
// Usage: addfeed <name> <url>
func handlerAddFeed(s *state, cmd command,user database.User,) error {
	if len(cmd.args) != 2 {
		return fmt.Errorf("usage: gator addfeed <name> <url>")
	}

	feedName := cmd.args[0]
	feedURL := cmd.args[1]
	ctx := context.Background()

	currentUser := user
	feed, err := s.db.CreateFeed(
		ctx,
		database.CreateFeedParams{
			ID:        uuid.New(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Name:      feedName,
			Url:       feedURL,
			UserID:    currentUser.ID,
		},
	)
	if err != nil {
		return fmt.Errorf("couldn't create feed: %w", err)
	}

	feedFollow, err := s.db.CreateFeedFollow(
		ctx,
		database.CreateFeedFollowParams{
			ID:        uuid.New(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			UserID:    currentUser.ID,
			FeedID:    feed.ID,
		},
	)
	if err != nil {
		return fmt.Errorf(
			"feed %q was created, but couldn't follow it: %w",
			feed.Name,
			err,
		)
	}

	fmt.Println("Feed created successfully:")
	fmt.Printf("ID:      %s\n", feed.ID)
	fmt.Printf("Name:    %s\n", feed.Name)
	fmt.Printf("URL:     %s\n", feed.Url)
	fmt.Printf("User ID: %s\n", feed.UserID)

	fmt.Printf(
		"%s is now following %s\n",
		feedFollow.UserName,
		feedFollow.FeedName,
	)

	return nil
}

//Get feeds
func handlerFeeds(s *state, cmd command, user database.User,) error {

        fds, err := s.db.GetFeeds(context.Background())
        if err != nil {
                return fmt.Errorf("couldn's load feeds", err)
        }

        for _,fd := range fds {
		fmt.Printf("-------\n")
		fmt.Printf("Feed Name:  %s\n", fd.Name)
		fmt.Printf("Feed URL:   %s\n", fd.Url)
		fmt.Printf("User Name:  %s\n", fd.UserName)
		fmt.Printf("========\n")
        }
        return nil
}

//Follow Feeds
func handlerFollow(s *state, cmd command, user database.User,) error {
	if len(cmd.args) != 1 {
		return fmt.Errorf("usage: gator follow <url>")
	}

	feedURL := cmd.args[0]
	ctx := context.Background()

	currentUser:=user

//	currentUser, err := s.db.GetUser(ctx, s.cfg.CurrentUserName)
//	if err != nil {
//		return fmt.Errorf(
//			"couldn't get current user %q: %w",
//			s.cfg.CurrentUserName,
//			err,
//		)
//	}

	feed, err := s.db.GetFeedByURL(ctx, feedURL)
	if err != nil {
		return fmt.Errorf(
			"couldn't find feed with URL %q: %w",
			feedURL,
			err,
		)
	}

	feedFollow, err := s.db.CreateFeedFollow(
		ctx,
		database.CreateFeedFollowParams{
			ID:        uuid.New(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			UserID:    currentUser.ID,
			FeedID:    feed.ID,
		},
	)
	if err != nil {
		return fmt.Errorf("couldn't follow feed: %w", err)
	}

	fmt.Printf(
		"%s is now following %s\n",
		feedFollow.UserName,
		feedFollow.FeedName,
	)

	return nil
}

//following
func handlerFollowing(s *state, cmd command,user database.User,) error {
	if len(cmd.args) != 0 {
		return fmt.Errorf("usage: gator following")
	}

	ctx := context.Background()

	currentUser:=user

//	currentUser, err := s.db.GetUser(ctx, s.cfg.CurrentUserName)
//	if err != nil {
//		return fmt.Errorf(
//			"couldn't get current user %q: %w",
//			s.cfg.CurrentUserName,
//			err,
//		)
//	}

	feedFollows, err := s.db.GetFeedFollowsForUser(
		ctx,
		currentUser.ID,
	)
	if err != nil {
		return fmt.Errorf(
			"couldn't get feed follows for %q: %w",
			currentUser.Name,
			err,
		)
	}

	if len(feedFollows) == 0 {
		fmt.Printf("%s is not following any feeds\n", currentUser.Name)
		return nil
	}

	fmt.Printf("Feeds followed by %s:\n", currentUser.Name)

	for _, feedFollow := range feedFollows {
		fmt.Printf("* %s\n", feedFollow.FeedName)
	}

	return nil
}

//Delete a follow
func handleDeleteFeedFollow(s *state, cmd command, user database.User) error {
	if len(cmd.args) != 1 {
		return fmt.Errorf("usage: gator follow <url>")
	}

	feedURL := cmd.args[0]
	ctx := context.Background()

	currentUser:=user
	feed, err := s.db.GetFeedByURL(ctx, feedURL)
	if err != nil {
		return fmt.Errorf(
			"couldn't find feed with URL %q: %w",
			feedURL,
			err,
		)
	}


	err2 := s.db.DeleteFeedFollow(ctx, database.DeleteFeedFollowParams{
                        UserID:    currentUser.ID,
                        FeedID:    feed.ID,
                }, )
        return err2
}

func printFeed(feed database.Feed, user database.User) {
	fmt.Printf("* ID:            %s\n", feed.ID)
	fmt.Printf("* Created:       %v\n", feed.CreatedAt)
	fmt.Printf("* Updated:       %v\n", feed.UpdatedAt)
	fmt.Printf("* Name:          %s\n", feed.Name)
	fmt.Printf("* URL:           %s\n", feed.Url)
	fmt.Printf("* User:          %s\n", user.Name)
	fmt.Printf("* LastFetchedAt: %v\n", feed.LastFetchedAt.Time)
}
