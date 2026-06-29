package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/marcoscbatista/oh-my-dot/internal/dotfiles"
)

func main() {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal("could not find home folder")
	}

	store, err := dotfiles.NewDotFileStore("oh-my-dot/db.json")
	if err != nil {
		log.Fatal(err)
	}

	service := dotfiles.DotFilesService{
		Store: store,
	}

	handler := dotfiles.DotFilesHandler{
		Service:     &service,
		DotfilesDir: filepath.Join(home, "oh-my-dot"),
		ConfigPath:  filepath.Join(home, ".config"),
	}

	if len(os.Args) < 2 {
		printUsage()
		return
	}

	command := os.Args[1]

	switch command {
	case "list":
		dots, err := handler.GetAll()
		if err != nil {
			fmt.Printf("Error: %s\n", err)
			return
		}

		if len(dots) == 0 {
			fmt.Println("No dotfiles found.")
			return
		}

		for i, dot := range dots {
			fmt.Printf("%d - %s\n", i+1, dot.Name)
		}

	case "switch":
		if len(os.Args) < 3 {
			fmt.Println("You need to send the dotfiles name you want to activate.")
			return
		}

		name := os.Args[2]

		if err := handler.Switch(name); err != nil {
			fmt.Printf("Error: %s\n", err)
			return
		}

		fmt.Printf("Dotfiles %q activated successfully.\n", name)

	case "create":
		if len(os.Args) < 4 {
			fmt.Println("Usage: oh-my-dot create <name> <remote-address>")
			return
		}

		name := os.Args[2]
		remoteAddr := os.Args[3]

		if err := handler.Create(name, remoteAddr); err != nil {
			fmt.Printf("Error: %s\n", err)
			return
		}

		fmt.Printf("Dotfiles %q created successfully.\n", name)

	default:
		printUsage()
	}
}

func printUsage() {
	fmt.Println(`Commands:
  list                          List all dotfiles
  switch <name>                 Switch the active dotfiles
  create <name> <remote-url>    Add your dotfiles

Warning:
  If your current .config is not managed by oh-my-dot, it will be replaced.
  A backup will be created in ~/oh-my-dot.`)
}
