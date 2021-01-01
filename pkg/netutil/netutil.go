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

package netutil

import (
	"bytes"
	"encoding/json"
	"html/template"
	"net"
	"os"
	"sort"
	"strings"

	"github.com/containerd/containerd/errdefs"
	"github.com/containernetworking/cni/libcni"
	"github.com/pkg/errors"
)

type NetworkConfigList struct {
	*libcni.NetworkConfigList
	Nerdctl bool
	File    string
}

const (
	DefaultNetworkName = "nerdctl"
	DefaultBridgeName  = "cni-nerdctl0"
	DefaultCIDR        = "10.4.0.0/16"
)

func DefaultConfigList() *NetworkConfigList {
	l, err := GenerateConfigList(DefaultNetworkName, DefaultBridgeName, DefaultCIDR)
	if err != nil {
		panic(err)
	}
	return l
}

type ConfigListTemplateOpts struct {
	Name       string // e.g. "nerdctl"
	BridgeName string // e.g. "cni-nerdctl0"
	Subnet     string // e.g. "10.4.0.0/16"
	Gateway    string // e.g. "10.4.0.1"
}

// ConfigListTemplate was copied from https://github.com/containers/podman/blob/v2.2.0/cni/87-podman-bridge.conflist
const ConfigListTemplate = `{
  "cniVersion": "0.4.0",
  "name": "{{.Name}}",
  "nerdctl": true,
  "plugins": [
    {
      "type": "bridge",
      "bridge": "{{.BridgeName}}",
      "isGateway": true,
      "ipMasq": true,
      "hairpinMode": true,
      "ipam": {
        "type": "host-local",
        "routes": [{ "dst": "0.0.0.0/0" }],
        "ranges": [
          [
            {
              "subnet": "{{.Subnet}}",
              "gateway": "{{.Gateway}}"
            }
          ]
        ]
      }
    },
    {
      "type": "portmap",
      "capabilities": {
        "portMappings": true
      }
    },
    {
      "type": "firewall"
    },
    {
      "type": "tuning"
    }
  ]
}`

func RequiredCNIPlugins(l *NetworkConfigList) []string {
	plugins := make([]string, len(l.Plugins))
	for i, f := range l.Plugins {
		plugins[i] = f.Network.Type
	}
	return plugins
}

// GenerateConfigList creates NetworkConfigList.
// GenerateConfigList does not fill "File" field.
func GenerateConfigList(name, bridgeName, cidr string) (*NetworkConfigList, error) {
	if name == "" || bridgeName == "" || cidr == "" {
		return nil, errdefs.ErrInvalidArgument
	}
	subnetIP, subnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, errors.Errorf("failed to parse CIDR %q", cidr)
	}
	if !subnet.IP.Equal(subnetIP) {
		return nil, errors.Errorf("unexpected CIDR %q, maybe you meant %q?", cidr, subnet.String())
	}
	gateway := make(net.IP, len(subnet.IP))
	copy(gateway, subnet.IP)
	gateway[3] += 1
	opts := &ConfigListTemplateOpts{
		Name:       name,
		BridgeName: bridgeName,
		Subnet:     subnet.String(),
		Gateway:    gateway.String(),
	}
	tmpl, err := template.New("").Parse(ConfigListTemplate)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, opts); err != nil {
		return nil, err
	}
	l, err := libcni.ConfListFromBytes(buf.Bytes())
	if err != nil {
		return nil, err
	}
	return &NetworkConfigList{
		NetworkConfigList: l,
		Nerdctl:           true,
		File:              "",
	}, nil
}

// ConfigLists loads config from dir if dir exists.
// The result also contains DefaultConfigList
func ConfigLists(dir string) ([]*NetworkConfigList, error) {
	l := []*NetworkConfigList{
		DefaultConfigList(),
	}
	if _, err := os.Stat(dir); err != nil {
		if os.IsNotExist(err) {
			return l, nil
		}
		return nil, err
	}
	fileNames, err := libcni.ConfFiles(dir, []string{".conf", ".conflist", ".json"})
	if err != nil {
		return nil, err
	}
	sort.Strings(fileNames)
	for _, fileName := range fileNames {
		var lcl *libcni.NetworkConfigList
		if strings.HasSuffix(fileName, ".conflist") {
			lcl, err = libcni.ConfListFromFile(fileName)
			if err != nil {
				return nil, err
			}
		} else {
			lc, err := libcni.ConfFromFile(fileName)
			if err != nil {
				return nil, err
			}
			lcl, err = libcni.ConfListFromConf(lc)
			if err != nil {
				return nil, err
			}
		}
		l = append(l, &NetworkConfigList{
			NetworkConfigList: lcl,
			Nerdctl:           !IsExternalConfigList(lcl.Bytes),
			File:              fileName,
		})
	}
	return l, nil
}

// IsExternalConfigList returns true if the config list is managed outside nerdctl.
func IsExternalConfigList(b []byte) bool {
	type nerdctlConfigList struct {
		Nerdctl bool `json:"nerdctl,omitempty"`
	}
	var ncl nerdctlConfigList
	if err := json.Unmarshal(b, &ncl); err != nil {
		return true
	}
	return !ncl.Nerdctl
}
