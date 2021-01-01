/*
   Copyright (C) nerdctl authors.
   Copyright (C) containerd authors.

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package main

import (
	"fmt"
	"text/tabwriter"

	"github.com/AkihiroSuda/nerdctl/pkg/netutil"
	"github.com/urfave/cli/v2"
)

var networkLsCommand = &cli.Command{
	Name:    "ls",
	Aliases: []string{"list"},
	Usage:   "List networks",
	Action:  networkLsAction,
}

func networkLsAction(clicontext *cli.Context) error {
	ll, err := netutil.ConfigLists(clicontext.String("cni-netconfpath"))
	if err != nil {
		return err
	}
	w := tabwriter.NewWriter(clicontext.App.Writer, 4, 8, 4, ' ', 0)
	fmt.Fprintln(w, "NAME\tKIND")
	for _, l := range ll {
		kind := "nerdctl"
		if l.File == "" {
			kind = "nerdctl (builtin)"
		} else if !l.Nerdctl {
			kind = "external"
		}
		fmt.Fprintf(w, "%s\t%s\n", l.Name, kind)
	}
	return w.Flush()
}
