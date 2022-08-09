package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/stv0g/gont/internal"
	"github.com/stv0g/gont/internal/utils"
	g "github.com/stv0g/gont/pkg"
	"golang.org/x/exp/slices"
)

var GitCommit string
var GitTag string

func usage() {
	w := flag.CommandLine.Output() // may be os.Stderr - but not necessarily

	argv0 := filepath.Base(os.Args[0])

	fmt.Fprintf(w, "Usage: %s [flags] <command>\n", argv0)
	fmt.Fprintln(w)
	fmt.Fprintln(w, " Supported <commands> are:")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "   identify                                     return the network and node name if gontc is executed within a network namespace")
	fmt.Fprintln(w, "   shell [<network>]/<node>                     get an interactive shell inside <node>")
	fmt.Fprintln(w, "   exec  [<network>]/<node> <command> [args]    executes a <command> in the namespace of <node> with optional [args]")
	fmt.Fprintln(w, "   list  [<network>]                            list all active Gont networks or nodes of a given network")
	fmt.Fprintln(w, "   clean [<network>]                            removes the all or just the specified Gont network")
	fmt.Fprintln(w, "   help                                         show this usage information")
	fmt.Fprintln(w, "   version                                      shows the version of Gont")
	// fmt.Fprintln(w)
	// fmt.Fprintln(w, " Supported [flags] are:")
	// flag.PrintDefaults()
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Example:")
	fmt.Fprintln(w)
	fmt.Fprintf(w, "   %s exec zorn/h1 ping h2\n", argv0)
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Gont - The Go network tester")
	fmt.Fprintln(w, "   Author Steffen Vogel <post@steffenvogel>")
}

func main() {
	var err error
	var network, node string

	internal.SetupRand()
	logger := internal.SetupLogging()
	defer logger.Sync()

	if err := g.CheckCaps(); err != nil {
		fmt.Printf("error: %s\n", err)
		os.Exit(-1)
	}

	flag.Usage = usage
	flag.Parse()

	if len(flag.Args()) < 1 {
		flag.CommandLine.Usage()
		os.Exit(-1)
	}

	args := flag.Args()
	subcmd := args[0]

	switch subcmd {
	case "shell":
		if network, node, err = getNetworkNode(args); err == nil {
			shell := os.Getenv("SHELL")
			if shell == "" {
				shell = "/bin/bash"
			}

			ps1 := fmt.Sprintf("%s/%s: ", network, node)
			os.Setenv("PS1", ps1)

			cmd := []string{shell, "--norc"}
			err = execCommand(network, node, cmd)
		}

	case "exec":
		if network, node, err = getNetworkNode(args); err == nil {
			err = execCommand(network, node, args[2:])
		}

	case "clean":
		if len(args) > 1 {
			network := args[1]
			if err = g.TeardownNetwork(network); err != nil {
				err = fmt.Errorf("failed to teardown network '%s': %w", network, err)
			}
		} else {
			err = g.TeardownAllNetworks()
		}

	case "list":
		if len(args) > 1 {
			network = args[1]
			for _, name := range g.NodeNames(network) {
				fmt.Printf("%s/%s\n", network, name)
			}
		} else {
			for _, name := range g.NetworkNames() {
				fmt.Println(name)
			}

		}

	case "identify":
		if network, node, err = g.Identify(); err == nil {
			fmt.Printf("%s/%s\n", network, node)
		}

	case "version":
		fmt.Printf("%s-%s\n", GitTag, GitCommit[:7])

	case "help":
		flag.Usage()
		err = nil

	default:
		err = fmt.Errorf("unknown sub-command: %s", subcmd)
	}

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(-1)
	}
}

func execCommand(network, node string, args []string) error {
	if len(flag.Args()) <= 1 {
		return fmt.Errorf("not enough arguments")
	}

	if network == "" {
		return fmt.Errorf("there is no active Gont network")
	}

	if err := os.Setenv("GONT_NETWORK", network); err != nil {
		return err
	}
	if err := os.Setenv("GONT_NODE", node); err != nil {
		return err
	}

	return g.Exec(network, node, args)
}

func getNetworkNode(args []string) (string, string, error) {
	var node, network string

	networks := g.NetworkNames()

	c := strings.SplitN(args[1], "/", 2)
	if len(c) == 1 { // no network in name
		if len(networks) == 0 {
			return "", "", errors.New("no-existing Gont network")
		}

		network = networks[0]
		node = c[0]
	} else {
		network = c[0]
		node = c[1]

		if !slices.Contains(networks, network) {
			return "", "", fmt.Errorf("non-existing network '%s'", network)
		}
	}

	nodes := g.NodeNames(network)

	if !slices.Contains(nodes, node) {
		return "", "", fmt.Errorf("non-existing node '%s' in network '%s'", node, network)
	}

	return network, node, nil
}
