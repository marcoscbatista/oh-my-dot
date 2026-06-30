package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/marcoscbatista/oh-my-dot/internal/dotfiles"
)

type Logger struct {
	verbose bool
}

func (l *Logger) Verbosef(format string, args ...any) {
	if l.verbose {
		fmt.Fprintf(os.Stderr, format+"\n", args...)
	}
}

func (l *Logger) Errorf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
}

func main() {
	isVerbose := flag.Bool("v", false, "Activate verbose mode")
	flag.Parse()

	log := Logger{verbose: *isVerbose}

	args := flag.Args()

	home, err := os.UserHomeDir()
	if err != nil {
		log.Errorf("Error: %s", err)
		os.Exit(1)
	}

	log.Verbosef("Using home directory: %s", home)

	store, err := dotfiles.NewDotFileStore("oh-my-dot/db.json")
	if err != nil {
		log.Errorf("Error: %s", err)
		os.Exit(1)
	}

	log.Verbosef("Loaded dotfiles store")

	service := dotfiles.DotFilesService{
		Store: store,
	}

	handler := dotfiles.DotFilesHandler{
		Service:     &service,
		DotfilesDir: filepath.Join(home, "oh-my-dot"),
		ConfigPath:  filepath.Join(home, ".fake-config"),
	}

	log.Verbosef("Using dotfiles directory: %s", handler.DotfilesDir)
	log.Verbosef("Using config path: %s", handler.ConfigPath)

	if len(args) < 1 {
		printUsage()
		os.Exit(1)
	}

	command := args[0]

	switch command {
	case "list":
		log.Verbosef("Listing dotfiles")

		dots, err := handler.GetAll()
		if err != nil {
			log.Errorf("Error: %s", err)
			os.Exit(1)
		}

		if len(dots) == 0 {
			log.Errorf("Error: No dotfiles found.")
			os.Exit(1)
		}

		for _, dot := range dots {
			log.Verbosef("%v", dot)
			if dot.InUse {
				fmt.Fprintf(os.Stdout, "%d - %s activated\n", dot.ID, dot.Name)
			} else {
				fmt.Fprintf(os.Stdout, "%d - %s\n", dot.ID, dot.Name)
			}
		}

	case "switch":
		switchFlags := flag.NewFlagSet("switch", flag.ExitOnError)

		force := switchFlags.Bool("force", false, "Replace current ~/.config even if it is not managed by oh-my-dot")
		forceShort := switchFlags.Bool("f", false, "Replace current ~/.config even if it is not managed by oh-my-dot")

		switchFlags.Parse(args[1:])

		if switchFlags.NArg() < 1 {
			log.Errorf("Error: You need to send the dotfiles name you want to activate.")
			os.Exit(1)
		}

		id, err := strconv.Atoi(switchFlags.Arg(0))
		if err != nil {
			log.Errorf("Error: id must be a number.")
			os.Exit(1)
		}

		log.Verbosef("Switching to dotfiles %q", id)

		if err := handler.Switch(id, *force || *forceShort); err != nil {
			log.Errorf("Error: %s", err)
			os.Exit(1)
		}

		fmt.Fprintf(os.Stdout, "Dotfiles %d activated successfully.\n", id)

	case "create":
		if len(args) < 3 {
			fmt.Fprintln(os.Stderr, "Usage: oh-my-dot create <name> <remote-address>")
			os.Exit(1)
		}

		name := args[1]
		remoteAddr := args[2]

		log.Verbosef("Creating dotfiles %q", name)
		log.Verbosef("Remote address: %s", remoteAddr)
		log.Verbosef("Destination: %s", filepath.Join(handler.DotfilesDir, name))

		if err := handler.Create(name, remoteAddr, *isVerbose); err != nil {
			log.Errorf("Error: %s", err)
			os.Exit(1)
		}

		fmt.Fprintf(os.Stdout, "Dotfiles %q created successfully.\n", name)

	default:
		printUsage()
		os.Exit(1)
	}
}
func printUsage() {
	fmt.Fprintln(os.Stderr, `Commands:
  list                          List all dotfiles
  switch <id>                 Switch the active dotfiles
  create <name> <remote-url>    Add your dotfiles`)
}
