package main

import (
	"context"
	"database/sql"
	_ "errors"
	"fmt"
	_ "log"
	"os"

	"gatorcli/internal/config"
	"gatorcli/internal/database"

	_ "github.com/lib/pq"
)

// state contains shared application dependencies.
// Database access can be added here later.
type state struct {
	db  *database.Queries
	cfg *config.Config
}

// command represents a parsed command-line command.
type command struct {
	name string
	args []string
}


// commands stores the handlers registered with the CLI.
type commands struct {
	handlers map[string]func(*state, command) error
}

// run finds and executes the handler registered for the command.
func (c *commands) run(s *state, cmd command) error {
	handler, exists := c.handlers[cmd.name]
	if !exists {
		return fmt.Errorf("unknown command: %s", cmd.name)
	}

	return handler(s, cmd)
}

// register associates a command name with a handler function.
func (c *commands) register(
	name string,
	handler func(*state, command) error,
) {
	c.handlers[name] = handler
}

func middlewareLoggedIn(
	handler func(s *state, cmd command, user database.User) error,
	) func(*state, command) error {

	return func(s *state, cmd command) error {

		user, err := s.db.GetUser(
			context.Background(),
			s.cfg.CurrentUserName,
		)
		if err != nil {
			return err
		}

		return handler(s, cmd, user)
	}
}

func main() {
	cfg, err := config.Read()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading config: %v\n", err)
		os.Exit(1)
	}

	db, err := sql.Open("postgres", cfg.DBURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error opening database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	dbQueries := database.New(db)

	programState := &state{
		db:  dbQueries,
		cfg: &cfg,
	}

	cmds := commands{
		handlers: make(map[string]func(*state, command) error),
	}

	cmds.register("login", handlerLogin)
	cmds.register("register", handlerRegister)
	cmds.register("reset", handlerReset)
	cmds.register("users", handlerUsers)
	cmds.register("agg", handlerAgg)
	cmds.register("addfeed", middlewareLoggedIn(handlerAddFeed))
	cmds.register("feeds", middlewareLoggedIn(handlerFeeds))
	cmds.register("follow",middlewareLoggedIn(handlerFollow))
	cmds.register("following",middlewareLoggedIn(handlerFollowing))
	cmds.register("unfollow",middlewareLoggedIn(handleDeleteFeedFollow))
	cmds.register("agg", handlerAgg)
	cmds.register("browse", middlewareLoggedIn(handlerBrowse))

	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: gator <command> [arguments...]")
		os.Exit(1)
	}

	cmd := command{
		name: os.Args[1],
		args: os.Args[2:],
	}

	if err := cmds.run(programState, cmd); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
