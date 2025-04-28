package main

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/muszkin/blog-rss-feed/internal/config"
	"github.com/muszkin/blog-rss-feed/internal/database"
	rss_feed "github.com/muszkin/blog-rss-feed/internal/rss-feed"
	"os"
	"strconv"
	"time"
)

type state struct {
	db     *database.Queries
	config *config.Config
}
type command struct {
	name      string
	arguments []string
}
type commands struct {
	handlers map[string]func(*state, command) error
}

func (c *commands) register(name string, f func(*state, command) error) {
	c.handlers[name] = f
}
func (c *commands) run(s *state, cmd command) error {
	commandToRun, ok := c.handlers[cmd.name]
	if !ok {
		return fmt.Errorf("command %v not registered", cmd.name)
	}
	return commandToRun(s, cmd)
}

func main() {
	var s state
	c, err := config.Read()
	if err != nil {
		fmt.Printf("Cannot read config file: %v\n", err)
	}
	s.config = &c
	db, err := sql.Open("postgres", s.config.DbURL)
	dbQueries := database.New(db)
	s.db = dbQueries
	var cmds commands
	cmds.handlers = make(map[string]func(*state, command) error)
	cmds.register("login", handlerLogin)
	cmds.register("register", handlerRegister)
	cmds.register("reset", handleReset)
	cmds.register("users", handleUsers)
	cmds.register("agg", handleAgg)
	cmds.register("addfeed", middlewareLoggedIn(handleAddFeed))
	cmds.register("feeds", handleFeeds)
	cmds.register("follow", middlewareLoggedIn(handleFollow))
	cmds.register("following", middlewareLoggedIn(handleFollowing))
	cmds.register("unfollow", middlewareLoggedIn(handleUnfollow))
	cmds.register("browse", middlewareLoggedIn(handleBrowse))
	args := os.Args
	if len(args) < 2 {
		fmt.Printf("too few arguments\n")
		os.Exit(1)
	}
	cmd := command{
		name:      args[1],
		arguments: args[2:],
	}
	if err := cmds.run(&s, cmd); err != nil {
		fmt.Printf("something goes wrong: %v\n", err)
		os.Exit(1)
	}
	os.Exit(0)
}

func handlerLogin(s *state, cmd command) error {
	if len(cmd.arguments) == 0 {
		return fmt.Errorf("you need to provide login\n")
	}
	_, err := s.db.GetUser(context.Background(), cmd.arguments[0])
	if err != nil {
		return fmt.Errorf("user does not exists")
	}
	err = s.config.SetUser(cmd.arguments[0])
	if err != nil {
		return err
	}
	fmt.Println("User has been set")
	fmt.Printf("Welcome %v!\n", cmd.arguments[0])
	return nil
}

func handlerRegister(s *state, cmd command) error {
	if len(cmd.arguments) == 0 {
		return fmt.Errorf("you need to provide name\n")
	}
	user := database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      cmd.arguments[0],
	}
	if _, err := s.db.CreateUser(context.Background(), user); err != nil {
		return fmt.Errorf("user already exists")
	}
	err := s.config.SetUser(user.Name)
	if err != nil {
		return err
	}
	fmt.Printf("User %v registered", user.Name)
	return nil
}

func handleReset(s *state, _ command) error {
	if err := s.db.Truncate(context.Background()); err != nil {
		return err
	}
	fmt.Println("Table users has been reset")
	return nil
}

func handleUsers(s *state, _ command) error {
	users, err := s.db.GetUsers(context.Background())
	if err != nil {
		return err
	}
	for _, user := range users {
		fmt.Printf(" * %s", user.Name)
		if user.Name == s.config.CurrentUserName {
			fmt.Printf(" (current)")
		}
		fmt.Printf("\n")
	}
	return nil
}

func handleAgg(s *state, cmd command) error {
	if len(cmd.arguments) == 0 {
		return fmt.Errorf("you need to provide duration in form of eg. 5s, 1m, 1h")
	}
	duration, err := time.ParseDuration(cmd.arguments[0])
	if err != nil {
		return err
	}
	fmt.Printf("Collecting feeds every %v\n", cmd.arguments[0])
	ticker := time.NewTicker(duration)
	for ; ; <-ticker.C {
		scrapeFeeds(s)
	}
}

func handleAddFeed(s *state, cmd command, user database.User) error {
	if len(cmd.arguments) < 2 {
		return fmt.Errorf("you need to provide name and url for feed")
	}
	existingFeed, err := s.db.GetFeedByUrl(context.Background(), cmd.arguments[1])
	var feedId uuid.UUID
	if err != nil && err.Error() == "sql: no rows in result set" {
		feed := database.CreateFeedParams{
			ID:        uuid.New(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Name:      cmd.arguments[0],
			Url:       cmd.arguments[1],
		}
		feedRecord, err := s.db.CreateFeed(context.Background(), feed)
		if err != nil {
			return err
		}
		fmt.Println(feedRecord)
		feedId = feed.ID
	} else if err != nil {
		return err
	} else {
		feedId = existingFeed.ID
	}
	createFeedFollowParams := database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    user.ID,
		FeedID:    feedId,
	}
	_, err = s.db.CreateFeedFollow(context.Background(), createFeedFollowParams)
	if err != nil {
		return err
	}
	return nil
}

func handleFeeds(s *state, _ command) error {
	userWithFeeds, err := s.db.GetUsersWithFeeds(context.Background())
	if err != nil {
		return err
	}
	for _, userWithFeeds := range userWithFeeds {
		fmt.Printf("Name: %s, Url: %s, User: %s\n", userWithFeeds.Name_2, userWithFeeds.Url, userWithFeeds.Name)
	}
	return nil
}

func handleFollow(s *state, cmd command, user database.User) error {
	if len(cmd.arguments) == 0 {
		return fmt.Errorf("you should provide url for feed")
	}
	feed, err := s.db.GetFeedByUrl(context.Background(), cmd.arguments[0])
	if err != nil {
		return err
	}
	createFeedFollowParams := database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    user.ID,
		FeedID:    feed.ID,
	}
	feedFollow, err := s.db.CreateFeedFollow(context.Background(), createFeedFollowParams)
	if err != nil {
		return err
	}
	fmt.Printf("Added: Feed '%s' to '%s'\n", feedFollow.FeedName, feedFollow.UserName)
	return nil
}

func handleFollowing(s *state, _ command, user database.User) error {
	feedFollows, err := s.db.GetFeedFollowsForUser(context.Background(), user.ID)
	if err != nil {
		return err
	}
	for _, feedFollow := range feedFollows {
		fmt.Printf("Feed '%s'\n", feedFollow.Name)
	}
	return nil
}

func handleUnfollow(s *state, cmd command, user database.User) error {
	if len(cmd.arguments) == 0 {
		return fmt.Errorf("you should provide feed url to unfollow")
	}
	feedUrl := cmd.arguments[0]
	feed, err := s.db.GetFeedByUrl(context.Background(), feedUrl)
	if err != nil {
		return err
	}
	err = s.db.DeleteFeedFollow(context.Background(), database.DeleteFeedFollowParams{
		FeedID: feed.ID,
		UserID: user.ID,
	})
	if err != nil {
		return err
	}
	return nil
}

func handleBrowse(s *state, cmd command, user database.User) error {
	var limit int
	if len(cmd.arguments) == 0 {
		limit = 2
	} else {
		limit, _ = strconv.Atoi(cmd.arguments[0])
	}
	posts, err := s.db.GetPostsForUser(context.Background(), database.GetPostsForUserParams{
		ID:    user.ID,
		Limit: int32(limit),
	})
	if err != nil {
		return err
	}
	for _, post := range posts {
		fmt.Println(post.Title)
	}
	return nil
}

func scrapeFeeds(s *state) {
	nextFeed, err := s.db.GetNextFeedToFetch(context.Background())
	if err != nil {
		fmt.Printf("can't get next feed to check: %v\n", err)
	}
	err = s.db.MarkFeedFetched(context.Background(), nextFeed.ID)
	if err != nil {
		fmt.Printf("can't set feed as fetched: %v\n", err)
	}
	rssFeedData, err := rss_feed.FetchFeed(context.Background(), nextFeed.Url)
	if err != nil {
		fmt.Printf("can't download feed: %v\n", err)
	}
	fmt.Printf("Feed downloaded: %v\n", rssFeedData.Channel.Title)
	for _, rssItem := range rssFeedData.Channel.Item {
		publishedAt, _ := time.Parse("Y-m-d H:i:s", rssItem.PubDate)
		createPostParams := database.CreatePostParams{
			ID:          uuid.New(),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			Title:       rssItem.Title,
			Url:         rssItem.Link,
			Description: sql.NullString{String: rssItem.Description},
			PublishedAt: publishedAt,
			FeedID:      nextFeed.ID,
		}
		_, err := s.db.CreatePost(context.Background(), createPostParams)
		if err != nil {
			fmt.Printf("something goes wrong when saving post: %v\n", err)
		}
	}
}

func middlewareLoggedIn(handler func(s *state, cmd command, user database.User) error) func(*state, command) error {
	return func(s *state, cmd command) error {
		user, err := s.db.GetUser(context.Background(), s.config.CurrentUserName)
		if err != nil {
			return err
		}

		return handler(s, cmd, user)
	}
}
