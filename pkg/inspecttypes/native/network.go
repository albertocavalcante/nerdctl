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

package native

import "encoding/json"

// Network corresponds to pkg/netutil.NetworkConfigList
type Network struct {
	CNI     json.RawMessage `json:"CNI,omitempty"`
	Nerdctl bool            `json:"Nerdctl"`
	File    string          `json:"File,omitempty"`
}
