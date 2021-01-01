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
	"encoding/json"
	"fmt"

	"github.com/AkihiroSuda/nerdctl/pkg/inspecttypes/native"
	"github.com/AkihiroSuda/nerdctl/pkg/netutil"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
)

var networkInspectCommand = &cli.Command{
	Name:        "inspect",
	Usage:       "Display detailed information on one or more networks",
	ArgsUsage:   "[flags] NETWORK [NETWORK, ...]",
	Description: "NOTE: The output format is not compatible with Docker.",
	Action:      networkInspectAction,
}

func networkInspectAction(clicontext *cli.Context) error {
	if clicontext.NArg() == 0 {
		return errors.Errorf("requires at least 1 argument")
	}

	ll, err := netutil.ConfigLists(clicontext.String("cni-netconfpath"))
	if err != nil {
		return err
	}
	filled := make(map[string]struct{})
	var result []native.Network
	for _, l := range ll {
		for _, name := range clicontext.Args().Slice() {
			if l.Name == name {
				r := native.Network{
					CNI:     json.RawMessage(l.Bytes),
					Nerdctl: l.Nerdctl,
					File:    l.File,
				}
				result = append(result, r)
				filled[name] = struct{}{}
				break
			}
		}
	}
	for _, name := range clicontext.Args().Slice() {
		if _, ok := filled[name]; !ok {
			return errors.Errorf("no such network: %s", name)
		}
	}
	b, err := json.MarshalIndent(result, "", "    ")
	if err != nil {
		return err
	}
	fmt.Fprintln(clicontext.App.Writer, string(b))
	return nil
}
