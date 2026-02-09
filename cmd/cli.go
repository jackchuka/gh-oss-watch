package cmd

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/jackchuka/gh-oss-watch/services"
)

type CLI struct {
	configService services.ConfigService
	cacheService  services.CacheService
	githubService services.GitHubService
	output        services.Output
}

func NewCLI(configService services.ConfigService, cacheService services.CacheService, githubService services.GitHubService, output services.Output) *CLI {
	return &CLI{
		configService: configService,
		cacheService:  cacheService,
		githubService: githubService,
		output:        output,
	}
}

func (c *CLI) Run(args []string) {
	if len(args) < 2 {
		c.printUsage()
		return
	}

	// Parse global flags and command
	globalFlags, command, cmdArgs := c.parseGlobalFlags(args[1:])

	var err error

	switch command {
	case "init":
		err = c.handleInit()
	case "add":
		err = c.handleAddCommand(cmdArgs)
	case "set":
		err = c.handleSetCommand(cmdArgs)
	case "remove":
		err = c.handleRemoveCommand(cmdArgs)
	case "status":
		err = c.handleStatusCommand(cmdArgs, globalFlags)
	case "dashboard":
		err = c.handleDashboardCommand(cmdArgs, globalFlags)
	case "releases":
		err = c.handleReleasesCommand(cmdArgs, globalFlags)
	default:
		c.output.Printf("Unknown command: %s\n", command)
		c.printUsage()
		os.Exit(1)
	}

	if err != nil {
		c.output.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

type GlobalFlags struct {
	MaxConcurrent int
	Timeout       int
}

func (c *CLI) parseGlobalFlags(args []string) (GlobalFlags, string, []string) {
	flags := GlobalFlags{
		MaxConcurrent: 10,
		Timeout:       30,
	}

	var command string
	var cmdArgs []string

	for i := 0; i < len(args); i++ {
		arg := args[i]

		if after, ok := strings.CutPrefix(arg, "--max-concurrent="); ok {
			if val, err := strconv.Atoi(after); err == nil {
				flags.MaxConcurrent = val
			}
		} else if after, ok := strings.CutPrefix(arg, "--timeout="); ok {
			if val, err := strconv.Atoi(after); err == nil {
				flags.Timeout = val
			}
		} else if arg == "--max-concurrent" && i+1 < len(args) {
			if val, err := strconv.Atoi(args[i+1]); err == nil {
				flags.MaxConcurrent = val
				i++ // Skip next arg
			}
		} else if arg == "--timeout" && i+1 < len(args) {
			if val, err := strconv.Atoi(args[i+1]); err == nil {
				flags.Timeout = val
				i++ // Skip next arg
			}
		} else if command == "" && !strings.HasPrefix(arg, "-") {
			command = arg
		} else if command != "" {
			cmdArgs = append(cmdArgs, arg)
		}
	}

	return flags, command, cmdArgs
}

func (c *CLI) handleStatusCommand(_ []string, flags GlobalFlags) error {
	c.githubService.SetMaxConcurrent(flags.MaxConcurrent)
	c.githubService.SetTimeout(time.Duration(flags.Timeout) * time.Second)

	return c.handleStatus()
}

func (c *CLI) handleDashboardCommand(_ []string, flags GlobalFlags) error {
	c.githubService.SetMaxConcurrent(flags.MaxConcurrent)
	c.githubService.SetTimeout(time.Duration(flags.Timeout) * time.Second)

	return c.handleDashboard()
}

func (c *CLI) handleReleasesCommand(args []string, flags GlobalFlags) error {
	c.githubService.SetMaxConcurrent(flags.MaxConcurrent)
	c.githubService.SetTimeout(time.Duration(flags.Timeout) * time.Second)

	onlyUnreleased := false
	for _, arg := range args {
		if arg == "--only-unreleased" || arg == "-u" {
			onlyUnreleased = true
		}
	}

	return c.handleReleases(onlyUnreleased)
}

func (c *CLI) handleAddCommand(args []string) error {
	if len(args) < 1 {
		c.output.Println("Usage: gh oss-watch add <repo> [events...]")
		return fmt.Errorf("repository required")
	}
	return c.handleConfigAdd(args[0], args[1:])
}

func (c *CLI) handleSetCommand(args []string) error {
	if len(args) < 2 {
		c.output.Println("Usage: gh oss-watch set <repo> <events...>")
		return fmt.Errorf("repository and events required")
	}
	return c.handleConfigSet(args[0], args[1:])
}

func (c *CLI) handleRemoveCommand(args []string) error {
	if len(args) < 1 {
		c.output.Println("Usage: gh oss-watch remove <repo>")
		return fmt.Errorf("repository required")
	}
	return c.handleConfigRemove(args[0])
}

func (c *CLI) printUsage() {
	c.output.Println("gh-oss-watch - GitHub CLI plugin for OSS maintainers")
	c.output.Println("")
	c.output.Println("Usage:")
	c.output.Println("  gh oss-watch [flags] <command> [args...]")
	c.output.Println("")
	c.output.Println("Commands:")
	c.output.Println("  init                    Initialize config file")
	c.output.Println("  add <repo> [events...]  Add repo to watch list")
	c.output.Println("  set <repo> <events...>  Configure events for repo")
	c.output.Println("  remove <repo>           Remove repo from watch list")
	c.output.Println("  status                  Show new activity")
	c.output.Println("  releases                Show release status across all repos")
	c.output.Println("  dashboard               Show summary across all repos")
	c.output.Println("")
	c.output.Println("Performance Flags:")
	c.output.Println("  --max-concurrent <n>    Max concurrent API requests (default: 10)")
	c.output.Println("  --timeout <seconds>     Request timeout in seconds (default: 30)")
	c.output.Println("")
	c.output.Println("Examples:")
	c.output.Println("  gh oss-watch status --max-concurrent 20")
	c.output.Println("  gh oss-watch dashboard --timeout 60")
}
