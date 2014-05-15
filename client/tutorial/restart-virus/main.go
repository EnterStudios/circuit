// Copyright 2013 The Go Circuit Project
// Use of this source code is governed by the license for
// The Go Circuit Project, found in the LICENSE file.
//
// Authors:
//   2014 Petar Maymounkov <p@gocircuit.org>

package main

import (
	"os"
	"strconv"

	"github.com/gocircuit/circuit/client"
)

// pick returns the root anchor of a randomly-chosen circuit server in the cluster
func pick(c *client.Client) client.Anchor {
	for _, r := range c.View() {
		return r
	}
	panic(0)
}

func watch(c *client.Client, service string) {
	defer func() {
		recover()
	}()
	t := c.Walk(client.Split(service))
	t.Get().(client.Proc).Wait()
	t.Scrub()
}

// restart-virus dial_url service_anchor?
func main() {
	c := client.Dial(os.Args[1]) // argument is the url of a circuit server
	if len(os.Args) == 3 {
		watch(c, os.Args[2])
	}

	// start service
	service := client.Cmd{ // a pretend long-running user binary
		Path: "/bin/sleep",
		Args: []string{strconv.Itoa(5)}, // with simulated unexpected exits
	}
	a := pick(c)
	serviceAnchor := []string{"restart_virus", "service"}
	pservice, _ := a.Walk(serviceAnchor).MakeProc(service)
	pservice.Stdin().Close()
	println("started service")

	// start watcher
	b := pick(c)
	watcher := client.Cmd{
		Path: os.Args[0], // we assume that the binary of this tool is on the same path everywhere
		Args: []string{b.Addr(), "/" + a.Worker() + "/restart_virus/service"},
	}
	pwatcher, _ := b.Walk([]string{"restart_virus", "watcher"}).MakeProc(watcher)
	pwatcher.Stdin().Close()
	println("started watcher")
}