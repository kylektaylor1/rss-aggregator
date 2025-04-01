package config

import (
	"context"
	"errors"
	"fmt"
	"os"
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
	default:
		return errors.New("error - command not found")
	}
}
