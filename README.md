sub
===

sub is a simple subcommand package for Go. After getting annoyed with the options that were available, as they were either _way_ too complicated (Cobra) or simply didn't seem to work very well (subcommands), I decided to just make my own. So I did.

Example
-------

```go
var commander sub.Commander
commander.Help(`This is an example program. It doesn't really do anything particularly
interesting. In fact, it's not even really a whole program. It's just
a small piece of the entrypoint.`)

commander.Register(commander.HelpCmd())
commander.Register(&exampleCmd{})
err := commander.Run(append([]string{filepath.Base(os.Args[0])}, os.Args[1:]...))
if err != nil {
	if err == flag.ErrHelp {
		os.Exit(2)
	}

	fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	os.Exit(1)
}
```
