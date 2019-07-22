package sub_test

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"testing"

	"github.com/DeedleFake/sub"
)

type testCmd struct {
	w    io.Writer
	flag string
}

func (cmd *testCmd) Name() string {
	return "test"
}

func (cmd *testCmd) Desc() string {
	return "a simple test"
}

func (cmd *testCmd) Help() string {
	return `
This is just a simple test.
No, really. That's it.
Probably.
`
}

func (cmd *testCmd) Flags(fset *flag.FlagSet) {
	fset.StringVar(&cmd.flag, "flag", "test", "a flag test")
}

func (cmd *testCmd) Run(args []string) error {
	fmt.Fprintf(cmd.w, "%q", args[0])
	return nil
}

func TestSimpleCmd(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		cout    string
		testout string
		ret     error
	}{
		{
			name: "Simple Help",
			args: []string{"subtest", "--help"},
			cout: `Usage: subtest <subcommand> [subcommand arguments]

Even more help text.

Commands:
	help		show help for commands
	test		a simple test
`,
			testout: ``,
			ret:     flag.ErrHelp,
		},
		{
			name: "Subcommand Help",
			args: []string{"subtest", "test", "--help"},
			cout: `This is just a simple test.
No, really. That's it.
Probably.

Options:
  -flag string
    	a flag test (default "test")
`,
			testout: ``,
			ret:     flag.ErrHelp,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			var cout bytes.Buffer
			var testout bytes.Buffer

			c := &sub.Commander{
				Output: &cout,
				Help: `
Even more help text.
`,
			}

			c.Register(c.HelpCmd())
			c.Register(&testCmd{w: &testout})

			err := c.Run(test.args)
			if err != test.ret {
				t.Errorf("Expected:\t%v", test.ret)
				t.Errorf("Got:\t\t%v", err)
			}

			if out := cout.String(); out != test.cout {
				t.Errorf("Expected:\t%q", test.cout)
				t.Errorf("Got:\t\t%q", out)
			}

			if out := testout.String(); out != test.testout {
				t.Errorf("Expected:\t%q", test.testout)
				t.Errorf("Got:\t\t%q", out)
			}
		})
	}
}
