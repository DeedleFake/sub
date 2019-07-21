package sub

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// A Commander controls a set of subcommands.
type Commander struct {
	name     string
	help     string
	flags    func(*flag.FlagSet)
	commands []Command
}

// Register registers a command with the Commander. If a command with
// the same name as cmd already exists, it is replaced with cmd.
func (c *Commander) Register(cmd Command) {
	i := sort.Search(len(c.commands), func(i int) bool {
		return cmd.Name() < c.commands[i].Name()
	})
	if (i < len(c.commands)) && (cmd.Name() == c.commands[i].Name()) {
		c.commands[i] = cmd
		return
	}

	c.commands = append(c.commands[:i], append([]Command{cmd}, c.commands[i:]...)...)
}

func (c *Commander) get(name string) Command {
	for _, cmd := range c.commands {
		if cmd.Name() == name {
			return cmd
		}
	}

	return nil
}

// Help sets the help message. This should be a longer help message
// that explains what the program.
func (c *Commander) Help(help string) {
	c.help = help
}

// Flags sets flags as the function to use to fill the global FlagSet.
// If any function is set at all, it is assumed that there are global
// flags, which changes the structure of the global help message.
func (c *Commander) Flags(flags func(*flag.FlagSet)) {
	c.flags = flags
}

// Run runs the commander against the given arguments. The first
// argument should be the name of the executable. In many cases, this
// should be filepath.Base(os.Args[0]).
//
// If there is a problem with args, such as an attempt to call a
// non-existent command, flag.ErrHelp is returned. Otherwise, any
// errors returned from subcommand's Run method are returned directly.
func (c *Commander) Run(args []string) error {
	c.name = args[0]

	fset := flag.NewFlagSet(args[0], flag.ContinueOnError)
	fset.Usage = func() {
		_ = (*helpCmd)(c).Run(nil)
	}
	if c.flags != nil {
		c.flags(fset)
	}
	err := fset.Parse(args[1:])
	if err != nil {
		return err
	}

	if fset.NArg() == 0 {
		fset.Usage()
		return flag.ErrHelp
	}

	cmd := c.get(fset.Arg(0))
	if cmd == nil {
		fmt.Fprintf(os.Stderr, "Error: No such command: %q\n\n", fset.Arg(0))
		fset.Usage()
		return flag.ErrHelp
	}

	sub := flag.NewFlagSet(cmd.Name(), flag.ContinueOnError)
	sub.Usage = func() {
		_ = (*helpCmd)(c).Run([]string{cmd.Name()})
	}
	cmd.Flags(sub)
	err = sub.Parse(fset.Args()[1:])
	if err != nil {
		return err
	}

	return cmd.Run(sub.Args())
}

// Command is a subcommand.
type Command interface {
	// Name is the name of the command. This is what the user is
	// expected to enter in order to call this specific command.
	Name() string

	// Desc is a short description of the command.
	Desc() string

	// Help is a longer help message. It should ideally start with a
	// usage line. It does not need any particular whitespace around it.
	Help() string

	// Flags fills the given FlagSet. If the command has any flags, they
	// should be filled here. In most cases, the client will probably
	// want to use the Var variants of the flag declaration functions
	// with fields in the command's underlying type so that their values
	// can be accessed when the command is run.
	Flags(fset *flag.FlagSet)

	// Run actually runs the command. It is passed any leftover
	// arguments after the flags have been parsed.
	Run(args []string) error
}

type helpCmd Commander

// HelpCmd returns a "help" Command that provides help for c. If
// clients want an explicit "help" command to be available, this must
// be manually registered.
func (c *Commander) HelpCmd() Command {
	return (*helpCmd)(c)
}

func (h *helpCmd) Name() string {
	return "help"
}

func (h *helpCmd) Desc() string {
	return "show help for commands"
}

func (h *helpCmd) Help() string {
	return `Usage: help [command]

help displays a help summary for the entire set of commands or it
shows more detailed help for a specific named subcommand.`
}

func (h *helpCmd) Flags(*flag.FlagSet) {
}

func (h *helpCmd) Run(args []string) error {
	c := (*Commander)(h)
	if len(args) == 0 {
		name := c.name
		if name == "" {
			name = filepath.Base(os.Args[0])
		}

		globalOptions := ""
		if c.flags != nil {
			globalOptions = " [global options]"
		}

		fmt.Fprintf(os.Stderr, "Usage: %v%v <subcommand> [subcommand arguments]\n", name, globalOptions)
		if c.help != "" {
			fmt.Fprintf(os.Stderr, "\n%v\n", strings.TrimSpace(c.help))
		}
		if c.flags != nil {
			fmt.Fprintf(os.Stderr, "\nGlobal Options:\n")
			fset := flag.NewFlagSet(name, flag.ContinueOnError)
			c.flags(fset)
			fset.PrintDefaults()
		}
		fmt.Fprintf(os.Stderr, "\nCommands:\n")
		for _, cmd := range c.commands {
			fmt.Fprintf(os.Stderr, "\t%v\t\t%v\n", cmd.Name(), cmd.Desc())
		}

		return nil
	}

	cmd := c.get(args[0])
	if cmd == nil {
		fmt.Fprintf(os.Stderr, "Error: No such command: %q\n\n", args[0])
		_ = h.Run(nil)
		return flag.ErrHelp
	}

	if cmd.Help() != "" {
		fmt.Fprintf(os.Stderr, "%v\n", strings.TrimSpace(cmd.Help()))
	}

	var optionBuf bytes.Buffer
	fset := flag.NewFlagSet(cmd.Name(), flag.ContinueOnError)
	fset.SetOutput(&optionBuf)
	cmd.Flags(fset)
	fset.PrintDefaults()
	if optionBuf.Len() > 0 {
		fmt.Fprintf(os.Stderr, "\nOptions:\n")
		_, _ = io.Copy(os.Stderr, &optionBuf)
	}

	return nil
}
