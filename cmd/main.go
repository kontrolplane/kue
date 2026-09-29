package cmd

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/kontrolplane/kue/pkg/client"
	tui "github.com/kontrolplane/kue/pkg/tui"
	"github.com/kontrolplane/kue/pkg/tui/styles"
)

var (
	projectName = "kontrolplane"
	programName = "kue"
)

func fail(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}

// usage prints the help, with the flags spelled with two dashes.
func usage(w io.Writer, fs *flag.FlagSet) {
	_, _ = fmt.Fprintf(w, `%[1]s is a terminal user interface for managing AWS SQS queues and messages.

usage: %[1]s [flags]

the aws profile, region and endpoint are taken from the environment and the shared aws config,
e.g. AWS_PROFILE, AWS_REGION and AWS_ENDPOINT_URL.

flags:
`, programName)
	fs.VisitAll(func(f *flag.Flag) {
		name := "--" + f.Name
		if kind, _ := flag.UnquoteUsage(f); kind != "" {
			name += " " + kind
		}
		line := "  " + name
		if f.Name == "theme" {
			_, _ = fmt.Fprintf(w, "%-28s %s (default %q)\n", line, f.Usage, f.DefValue)
			return
		}
		_, _ = fmt.Fprintf(w, "%-28s %s\n", line, f.Usage)
	})
}

func Execute(version, commit, date string) {
	fs := flag.NewFlagSet(programName, flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	theme := fs.String("theme", "auto", "colour theme: auto follows the terminal, dark or light also paint the background")
	debug := fs.Bool("debug", false, "write debug logs to debug.log")
	showVersion := fs.Bool("version", false, "print the version and exit")
	if err := fs.Parse(os.Args[1:]); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			usage(os.Stdout, fs)
			return
		}
		fmt.Fprintf(os.Stderr, "%s (see %s --help)\n", styles.Clean(err.Error()), programName)
		os.Exit(2)
	}

	if fs.NArg() > 0 {
		fmt.Fprintf(os.Stderr, "unexpected argument %q, %s only takes flags (see %s --help)\n", fs.Arg(0), programName, programName)
		os.Exit(2)
	}

	if *showVersion {
		fmt.Printf("%s %s (commit %s, built %s)\n", programName, version, commit, date)
		return
	}

	switch *theme {
	case "dark":
		styles.Use(true)
		styles.Paint = true
	case "light":
		styles.Use(false)
		styles.Paint = true
	case "auto":
		styles.Use(lipgloss.HasDarkBackground(os.Stdin, os.Stdout))
	default:
		fmt.Fprintf(os.Stderr, "unknown theme %q, use auto, dark or light\n", styles.Clean(*theme))
		os.Exit(2)
	}

	log.SetOutput(io.Discard)
	if *debug {
		f, err := tea.LogToFile("debug.log", "debug")
		if err != nil {
			fail("could not open debug.log: %s", styles.Clean(err.Error()))
		}
		defer func() { _ = f.Close() }()
	}

	sqsClient, awsInfo, err := client.CreateSqsClient(context.Background())
	if err != nil {
		fail("could not create the sqs client: %s", styles.Clean(err.Error()))
	}

	model := tui.NewModel(projectName, programName, sqsClient, awsInfo)
	if _, err := tea.NewProgram(model).Run(); err != nil {
		fail("error running program: %s", styles.Clean(err.Error()))
	}
}
