package config

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/kylektaylor1/rss-aggregator/internal/database"
)

type Command struct {
	Name string
	Args []string
}

type Commands struct {
	Commands map[string]func(*State, Command) error
}

func HandleLogin(s *State, cmd Command) error {
	if len(cmd.Args) == 0 {
		return errors.New("expect non-zero args")
	}

	if len(cmd.Args) > 1 {
		return errors.New("too many args")
	}

	name := cmd.Args[0]

	user, err := s.Db.GetUser(context.Background(), name)
	if err != nil {
		fmt.Println("Error getting user")
		return err
	}

	setErr := s.Cfg.SetUser(user.Name)
	if setErr != nil {
		return setErr
	}

	fmt.Printf("Username set: %v\n", s.Cfg.CurrentUserName)
	return nil
}

func HandleReset(s *State, cmd Command) error {
	if len(cmd.Args) != 0 {
		return errors.New("expect zero args")
	}

	err := s.Db.DeleteUsers(context.Background())
	if err != nil {
		return fmt.Errorf("error deleting users: %w", err)
	}

	fmt.Println("Users deleted.")

	return nil
}

func HandleRegister(s *State, cmd Command) error {
	if len(cmd.Args) == 0 {
		return errors.New("expect non-zero args")
	}

	if len(cmd.Args) > 1 {
		return errors.New("too many args")
	}

	name := cmd.Args[0]
	id := uuid.New()

	_, err := s.Db.GetUser(context.Background(), name)
	if err == nil {
		fmt.Println("User found")
		os.Exit(1)
	}

	dbUser, err := s.Db.CreateUser(context.Background(), database.CreateUserParams{
		ID:        id,
		Name:      name,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})
	if err != nil {
		return err
	}

	s.Cfg.SetUser(dbUser.Name)

	fmt.Println("User Created!")
	fmt.Printf("User info: %v\n", dbUser)

	return nil
}

func HandleUsers(s *State, command Command) error {
	if len(command.Args) > 0 {
		return errors.New("too many args")
	}

	users, err := s.Db.GetUsers(context.Background())
	if err != nil {
		return fmt.Errorf("error getting users: %w", err)
	}

	for _, u := range users {
		if u.Name == s.Cfg.CurrentUserName {
			fmt.Printf("* %v (current)\n", u.Name)
			continue
		}
		fmt.Printf("* %v\n", u.Name)
	}

	return nil
}

func HandleAgg(s *State, cmd Command) error {
	if len(cmd.Args) != 1 {
		return errors.New("expect only one arg")
	}

	duration, err := time.ParseDuration(cmd.Args[0])
	if err != nil {
		return err
	}

	fmt.Printf("Collecting feeds every %v\n", duration)

	ticker := time.NewTicker(duration)
	for ; ; <-ticker.C {
		fmt.Println("queue scrape")
		ScrapeFeeds(s)
	}
}

func HandleAddFeed(s *State, cmd Command, user database.User) error {
	if len(cmd.Args) != 2 {
		return errors.New("you need two args")
	}

	name := cmd.Args[0]
	url := cmd.Args[1]

	newFeedRecord, err := s.Db.CreateFeed(context.Background(), database.CreateFeedParams{
		ID:            uuid.New(),
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
		Name:          name,
		Url:           url,
		UserID:        user.ID,
		LastFetchedAt: sql.NullTime{},
	})
	if err != nil {
		return err
	}

	_, ffErr := s.Db.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		UserID:    user.ID,
		FeedID:    newFeedRecord.ID,
	})
	if ffErr != nil {
		return ffErr
	}

	fmt.Println(newFeedRecord)

	return nil
}

func HandleFeeds(s *State, cmd Command) error {
	if len(cmd.Args) != 0 {
		return errors.New("no args needed")
	}

	feeds, err := s.Db.GetFeeds(context.Background())
	if err != nil {
		return err
	}

	for _, f := range feeds {
		user, err := s.Db.GetUserById(context.Background(), f.UserID)
		if err != nil {
			return err
		}

		fmt.Printf("Feed name: %v\n", f.Name)
		fmt.Printf("Feed URL: %v\n", f.Url)
		fmt.Printf("Feed owner name: %v\n", user.Name)
	}

	return nil
}

func HandleFollow(s *State, cmd Command, user database.User) error {
	if len(cmd.Args) != 1 {
		return errors.New("url arg needed")
	}

	feed, err := s.Db.GetFeedByUrl(context.Background(), cmd.Args[0])
	if err != nil {
		return err
	}

	newFeed, err := s.Db.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		UserID:    user.ID,
		FeedID:    feed.ID,
	})
	if err != nil {
		return err
	}

	fmt.Printf("Feed set. Feed name: %v. User name: %v\n", newFeed.FeedName, newFeed.UserName)

	return nil
}

func HandleFollowing(s *State, cmd Command, user database.User) error {
	if len(cmd.Args) != 0 {
		return errors.New("no args needed")
	}

	feeds, err := s.Db.GetFeedFollowsForUser(context.Background(), user.ID)
	if err != nil {
		return err
	}

	fmt.Printf("User: %v is following these feeds:\n", user.Name)

	for _, f := range feeds {
		fmt.Printf("Feed: %v\n", f.FeedName)
	}

	return nil
}

func HandleBrowse(s *State, cmd Command, user database.User) error {
	var limit int32
	limit = 2
	if len(cmd.Args) > 1 {
		return errors.New("at most one arg")
	}

	if len(cmd.Args) == 1 {
		newLimit, err := strconv.Atoi(cmd.Args[0])
		if err != nil {
			return err
		}
		limit = int32(newLimit)
	}

	posts, err := s.Db.GetPostsForUser(context.Background(), database.GetPostsForUserParams{
		UserID: user.ID,
		Limit:  limit,
	})
	if err != nil {
		return err
	}

	for _, p := range posts {
		fmt.Printf("Post title: %v\n", p.Title)
	}

	return nil
}

func HandleUnfollow(s *State, cmd Command, user database.User) error {
	if len(cmd.Args) != 1 {
		return errors.New("one arg expected")
	}

	err := s.Db.DeleteFeedFollows(context.Background(), database.DeleteFeedFollowsParams{
		UserID: user.ID,
		Url:    cmd.Args[0],
	})
	if err != nil {
		return err
	}

	fmt.Printf("Deleted feed url: %v", cmd.Args[0])

	return nil
}

func (c *Commands) Register(name string, f func(*State, Command) error) {
	if _, ok := c.Commands[name]; ok {
		fmt.Printf("This command name has already been registerd. Name: %v\n", name)
	}
	c.Commands[name] = f
}

func (c *Commands) Run(s *State, cmd Command) error {
	if _, ok := c.Commands[cmd.Name]; !ok {
		return errors.New("command does not exist")
	}
	switch cmd.Name {
	case "login":
		return c.Commands["login"](s, cmd)
	case "register":
		return c.Commands["register"](s, cmd)
	case "reset":
		return c.Commands["reset"](s, cmd)
	case "users":
		return c.Commands["users"](s, cmd)
	case "agg":
		return c.Commands["agg"](s, cmd)
	case "addfeed":
		return c.Commands["addfeed"](s, cmd)
	case "feeds":
		return c.Commands["feeds"](s, cmd)
	case "follow":
		return c.Commands["follow"](s, cmd)
	case "following":
		return c.Commands["following"](s, cmd)
	case "unfollow":
		return c.Commands["unfollow"](s, cmd)
	case "browse":
		return c.Commands["browse"](s, cmd)
	default:
		return errors.New("error - command not found")
	}
}
