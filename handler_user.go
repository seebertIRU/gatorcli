package main

import (
	"context"
	"fmt"
	"time"

	"gatorcli/internal/database"
	"github.com/google/uuid"
)

func handlerReset(s *state, cmd command) error {
	err := s.db.DELETEALLUSERS(context.Background())
	return err
}

func handlerRegister(s *state, cmd command) error {
	if len(cmd.args) != 1 {
		return fmt.Errorf("usage: %v <name>", cmd.name)
	}

	name := cmd.args[0]

	user, err := s.db.CreateUser(context.Background(), database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		Name:      name,
	})
	if err != nil {
		return fmt.Errorf("couldn't create user: %w", err)
	}

	err = s.cfg.SetUser(user.Name)
	if err != nil {
		return fmt.Errorf("couldn't set current user: %w", err)
	}

	fmt.Println("User created successfully:")
	printUser(user)
	return nil
}

func handlerLogin(s *state, cmd command) error {
	if len(cmd.args) != 1 {
		return fmt.Errorf("usage: %v <name>", cmd.name)
	}
	name := cmd.args[0]

	_, err := s.db.GetUser(context.Background(), name)
	if err != nil {
		return fmt.Errorf("couldn't find user: %w", err)
	}

	err = s.cfg.SetUser(name)
	if err != nil {
		return fmt.Errorf("couldn't set current user: %w", err)
	}

	fmt.Println("User switched successfully!")
	return nil
}

func handlerUsers(s *state, cmd command) error {


	usrs, err := s.db.GetUsers(context.Background())
	if err != nil {
		return fmt.Errorf("couldn's load users", err)
	}

	cu:=""
	for _,usr := range usrs {
		cu=""
		if usr.Name==s.cfg.CurrentUserName {
			cu=" (current)"
		}
		fmt.Printf(" * %v%v\n",usr.Name, cu)
	}
	return nil
}


func printUser(user database.User) {
	fmt.Printf(" * ID:      %v\n", user.ID)
	fmt.Printf(" * Name:    %v\n", user.Name)
}


