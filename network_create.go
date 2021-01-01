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
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/AkihiroSuda/nerdctl/pkg/netutil"
	"github.com/containerd/containerd/errdefs"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

var networkCreateCommand = &cli.Command{
	Name:        "create",
	Usage:       "Create a network",
	Description: "CNI support is WIP. No isolation across different networks.",
	ArgsUsage:   "[flags] NETWORK",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "subnet",
			Usage:    "Subnet in CIDR format that represents a network segment, e.g. \"10.5.0.0/16\"",
			Required: true, // FIXME
		},
	},
	Action: networkCreateAction,
}

func networkCreateAction(clicontext *cli.Context) error {
	if clicontext.NArg() != 1 {
		return errors.Errorf("requires exactly 1 argument")
	}
	name := clicontext.Args().First()
	if !IsAlphaNum(name) {
		return errors.Errorf("malformed name %s", name)
	}
	bridgeName := "cni-nerdctl-" + name

	netconfpath := clicontext.String("cni-netconfpath")
	if err := os.MkdirAll(netconfpath, 0755); err != nil {
		return err
	}

	ll, err := netutil.ConfigLists(netconfpath)
	if err != nil {
		return err
	}
	for _, l := range ll {
		if l.Name == name {
			return errors.Errorf("network with name %s already exists", name)
		}
		// TODO: check CIDR collision
	}

	l, err := netutil.GenerateConfigList(name, bridgeName, clicontext.String("subnet"))
	if err != nil {
		return err
	}
	filename := filepath.Join(netconfpath, "nerdctl-"+name+".conflist")
	if _, err := os.Stat(filename); err == nil {
		return errdefs.ErrAlreadyExists
	}
	// TODO: lock
	if err := ioutil.WriteFile(filename, l.Bytes, 0644); err != nil {
		return err
	}
	logrus.Warn("CNI support is WIP. No isolation across different networks.")
	return nil
}
