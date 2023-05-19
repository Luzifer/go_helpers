package cli

import (
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"
)

type (
	// Registry contains a collection of commands to be executed
	Registry struct {
		cmds map[string]RegistryEntry
		sync.Mutex
	}

	// Entry defines a sub-command with its parameters, description and
	// run function to be called when this command is executed
	RegistryEntry struct {
		Description string
		Name        string
		Params      []string
		Run         func([]string) error
	}
)

// ErrHelpCalled is returned from the Call function if the given
// command is not found and the help function was executed
var ErrHelpCalled = errors.New("help called")

// New creates a new Registry
func New() *Registry {
	return &Registry{
		cmds: make(map[string]RegistryEntry),
	}
}

// Add adds a new command to the Registry
func (c *Registry) Add(e RegistryEntry) {
	c.Lock()
	defer c.Unlock()

	c.cmds[e.Name] = e
}

// Call executes the matchign command from the given arguments
func (c *Registry) Call(args []string) error {
	c.Lock()
	defer c.Unlock()

	cmd := "help"
	if len(args) > 0 {
		cmd = args[0]
	}

	cmdEntry := c.cmds[cmd]
	if cmdEntry.Name != cmd {
		c.help()
		return ErrHelpCalled
	}

	return cmdEntry.Run(args)
}

func (c *Registry) help() {
	// Called from Call, does not need lock

	var (
		maxCmdLen int
		cmds      []RegistryEntry
	)

	for name := range c.cmds {
		entry := c.cmds[name]
		if l := len(entry.commandDisplay()); l > maxCmdLen {
			maxCmdLen = l
		}
		cmds = append(cmds, entry)
	}

	sort.Slice(cmds, func(i, j int) bool { return cmds[i].Name < cmds[j].Name })

	tpl := fmt.Sprintf("  %%-%ds  %%s\n", maxCmdLen)
	fmt.Fprintln(os.Stdout, "Supported sub-commands are:")
	for _, cmd := range cmds {
		fmt.Fprintf(os.Stdout, tpl, cmd.commandDisplay(), cmd.Description)
	}
}

func (c RegistryEntry) commandDisplay() string {
	return strings.Join(append([]string{c.Name}, c.Params...), " ")
}
