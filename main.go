package main

import (
	"fmt"
	config "github.com/muszkin/blog-rss-feed/internal/config"
	"os"
)

type state struct {
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
		return fmt.Errorf("Command %v not registered", cmd.name)
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
	var cmds commands
	cmds.handlers = make(map[string]func(*state, command) error)
	cmds.register("login", handlerLogin)
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
	err := s.config.SetUser(cmd.arguments[0])
	if err != nil {
		return err
	}
	fmt.Println("User has been set")
	fmt.Printf("Welcome %v!\n", cmd.arguments[0])
	return nil
}
