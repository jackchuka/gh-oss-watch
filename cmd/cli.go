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
	formatter     services.Formatter
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
	case "fans":
		err = c.handleFansCommand(cmdArgs, globalFlags)
	default:
		c.output.Printf("Unknown command: %s\n", command)
		c.printUsage()
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

type GlobalFlags struct {
	MaxConcurrent int
	Timeout       int
	Format        string
}

func (c *CLI) parseGlobalFlags(args []string) (GlobalFlags, string, []string) {
	flags := GlobalFlags{
		MaxConcurrent: 10,
		Timeout:       30,
		Format:        "plain",
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
		} else if after, ok := strings.CutPrefix(arg, "--format="); ok {
			flags.Format = after
		} else if after, ok := strings.CutPrefix(arg, "-f="); ok {
			flags.Format = after
		} else if (arg == "--format" || arg == "-f") && i+1 < len(args) {
			flags.Format = args[i+1]
			i++
		} else if command == "" && !strings.HasPrefix(arg, "-") {
			command = arg
		} else if command != "" {
			cmdArgs = append(cmdArgs, arg)
		}
	}

	if flags.Format != "plain" && flags.Format != "json" {
		fmt.Fprintf(os.Stderr, "Unknown format %q, supported: plain, json\n", flags.Format)
		os.Exit(1)
	}

	return flags, command, cmdArgs
}

func (c *CLI) handleStatusCommand(_ []string, flags GlobalFlags) error {
	c.githubService.SetMaxConcurrent(flags.MaxConcurrent)
	c.githubService.SetTimeout(time.Duration(flags.Timeout) * time.Second)
	c.formatter = services.NewFormatter(flags.Format)

	return c.handleStatus()
}

func (c *CLI) handleDashboardCommand(_ []string, flags GlobalFlags) error {
	c.githubService.SetMaxConcurrent(flags.MaxConcurrent)
	c.githubService.SetTimeout(time.Duration(flags.Timeout) * time.Second)
	c.formatter = services.NewFormatter(flags.Format)

	return c.handleDashboard()
}

func (c *CLI) handleReleasesCommand(args []string, flags GlobalFlags) error {
	c.githubService.SetMaxConcurrent(flags.MaxConcurrent)
	c.githubService.SetTimeout(time.Duration(flags.Timeout) * time.Second)
	c.formatter = services.NewFormatter(flags.Format)

	onlyUnreleased := false
	for _, arg := range args {
		if arg == "--only-unreleased" || arg == "-u" {
			onlyUnreleased = true
		}
	}

	return c.handleReleases(onlyUnreleased)
}

func (c *CLI) handleFansCommand(args []string, flags GlobalFlags) error {
	c.githubService.SetMaxConcurrent(flags.MaxConcurrent)
	c.githubService.SetTimeout(time.Duration(flags.Timeout) * time.Second)
	c.formatter = services.NewFormatter(flags.Format)

	top := 0
	for i, arg := range args {
		if after, ok := strings.CutPrefix(arg, "--top="); ok {
			if val, err := strconv.Atoi(after); err == nil {
				top = val
			}
		} else if after, ok := strings.CutPrefix(arg, "-t="); ok {
			if val, err := strconv.Atoi(after); err == nil {
				top = val
			}
		} else if (arg == "--top" || arg == "-t") && i+1 < len(args) {
			if val, err := strconv.Atoi(args[i+1]); err == nil {
				top = val
			}
		}
	}

	return c.handleFans(top)
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
	c.output.Println("  fans [--top N]          Show who starred your repos")
	c.output.Println("  dashboard               Show summary across all repos")
	c.output.Println("")
	c.output.Println("Flags:")
	c.output.Println("  -f, --format <format>   Output format: plain, json (default: plain)")
	c.output.Println("  --max-concurrent <n>    Max concurrent API requests (default: 10)")
	c.output.Println("  --timeout <seconds>     Request timeout in seconds (default: 30)")
	c.output.Println("")
	c.output.Println("Examples:")
	c.output.Println("  gh oss-watch status --max-concurrent 20")
	c.output.Println("  gh oss-watch dashboard --timeout 60")
}
