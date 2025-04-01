package main

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/kylektaylor1/rss-aggregator/internal/config"
	"github.com/kylektaylor1/rss-aggregator/internal/database"

	_ "github.com/lib/pq"
)

func main() {
	origCfg, err := config.Read()
	if err != nil {
		fmt.Printf("error reading file: %v\n", err)
	}

	db, err := sql.Open("postgres", origCfg.DbUrl)
	if err != nil {
		fmt.Println("Error opening db connection")
		os.Exit(1)
	}
	dbQueries := database.New(db)

	state := config.State{
		Cfg: &origCfg,
		Db:  dbQueries,
	}

	cmds := make(map[string]func(*config.State, config.Command) error)
	commands := config.Commands{
		Commands: cmds,
	}

	commands.Register("login", config.HandleLogin)
	commands.Register("register", config.HandleRegister)
	commands.Register("reset", config.HandleReset)
	commands.Register("users", config.HandleUsers)
	args := os.Args

	if len(args) < 2 {
		fmt.Printf("not enough args \n")
		os.Exit(1)
	}

	command := args[1]
	commandArgs := args[2:]

	rErr := commands.Run(&state, config.Command{
		Name: command,
		Args: commandArgs,
	})

	if rErr != nil {
		fmt.Printf("Error: %v\n", rErr)
		os.Exit(1)
	}
}
