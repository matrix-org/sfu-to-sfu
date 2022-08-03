/*
Copyright 2022 The Matrix.org Foundation C.I.C.

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
	"flag"
	"fmt"
	"io/ioutil"
	"log"

	yaml "gopkg.in/yaml.v3"

	"maunium.net/go/mautrix/id"
)

func loadConfig(configFilePath string) (*config, error) {
	log.Printf("Loading %s", configFilePath)
	file, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		log.Fatal("Failed to read config", err)
	}
	var config config
	if err := yaml.Unmarshal(file, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal YAML: %s", err)
	}
	return &config, nil
}

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)
	configFilePath := flag.String("config", "config.yaml", "Configuration file path")
	flag.Parse()

	var config *config
	var err error
	if config, err = loadConfig(*configFilePath); err != nil {
		log.Fatal("Failed to load config file", err)
	}

	if err := initMatrix(config); err != nil {
		log.Fatal("Failed to init Matrix", err)
	}
}

type config struct {
	UserID        id.UserID
	HomeserverURL string
	AccessToken   string
}

type trackDesc struct {
	StreamID string `json:"stream_id"`
	TrackID  string `json:"track_id"`
}

type dataChannelMessage struct {
	Op      string `json:"op"`
	ID      string `json:"id"`
	Message string `json:"message,omitempty"`
	// XXX: is this even needed? we know which conf a given call is for...
	ConfID string      `json:"conf_id,omitempty"`
	Start  []trackDesc `json:"start,omitempty"`
	Stop   []trackDesc `json:"stop,omitempty"`
	SDP    string      `json:"sdp,omitempty"`
}
