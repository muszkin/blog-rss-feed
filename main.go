package main

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/muszkin/blog-rss-feed/internal/config"
	"github.com/muszkin/blog-rss-feed/internal/database"
	"os"
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
