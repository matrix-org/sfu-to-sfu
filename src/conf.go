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
	"errors"
	"log"
	"sync"

	"github.com/pion/webrtc/v3"
)

type localTrackInfo struct {
	streamID string
	trackID  string
	call     *call
}

type localTrackWithInfo struct {
	track *webrtc.TrackLocalStaticRTP
	info  localTrackInfo
}

type calls struct {
	callsMu sync.RWMutex
	calls   map[string]*call // By callID
}

type conf struct {
	confID   string
	calls    calls
	tracksMu sync.RWMutex
	tracks   []localTrackWithInfo
}

func (f *focus) getConf(confID string, create bool) (*conf, error) {
	f.confs.confsMu.Lock()
	defer f.confs.confsMu.Unlock()
	co := f.confs.confs[confID]
	if co == nil {
		if create {
			co = &conf{
				confID: confID,
			}
			f.confs.confs[confID] = co
			co.calls.calls = make(map[string]*call)
			co.tracks = []localTrackWithInfo{}
		} else {
			return nil, errors.New("no such conf")
		}
	}
	return co, nil
}

func (c *conf) getCall(callID string, create bool) (*call, error) {
	c.calls.callsMu.Lock()
	defer c.calls.callsMu.Unlock()
	ca := c.calls.calls[callID]
	if ca == nil {
		if create {
			ca = &call{
				callID: callID,
				conf:   c,
			}
			c.calls.calls[callID] = ca
		} else {
			return nil, errors.New("no such call")
		}
	}
	return ca, nil
}

func (c *conf) getLocalTrackIndicesByInfo(selectInfo localTrackInfo) (tracks []int, err error) {
	foundIndices := []int{}
	for index, track := range c.tracks {
		info := track.info
		if selectInfo.call != nil && selectInfo.call != info.call {
			continue
		}
		if selectInfo.streamID != "" && selectInfo.streamID != info.streamID {
			continue
		}
		if selectInfo.trackID != "" && selectInfo.trackID != info.trackID {
			continue
		}
		foundIndices = append(foundIndices, index)
	}

	if len(foundIndices) == 0 {
		log.Printf("Found no tracks for %+v", selectInfo)
		return nil, errors.New("no such tracks")
	} else {
		return foundIndices, nil
	}
}

func (c *conf) getLocalTrackByInfo(selectInfo localTrackInfo) (tracks []webrtc.TrackLocal, err error) {
	c.tracksMu.Lock()
	defer c.tracksMu.Unlock()

	indices, err := c.getLocalTrackIndicesByInfo(selectInfo)
	if err != nil {
		return nil, err
	}

	foundTracks := []webrtc.TrackLocal{}
	for _, index := range indices {
		foundTracks = append(foundTracks, c.tracks[index].track)
	}

	if len(foundTracks) == 0 {
		log.Printf("No tracks")
		return nil, errors.New("no such tracks")
	} else {
		return foundTracks, nil
	}
}

func (c *conf) removeTracksFromPeerConnectionsByInfo(removeInfo localTrackInfo) error {
	c.tracksMu.Lock()
	defer c.tracksMu.Unlock()

	indices, err := c.getLocalTrackIndicesByInfo(removeInfo)
	if err != nil {
		return err
	}

	// FIXME: the big O of this must be awful...
	for _, index := range indices {
		info := c.tracks[index].info

		for _, call := range c.calls.calls {
			for _, sender := range call.peerConnection.GetSenders() {
				if info.trackID == sender.Track().ID() {
					log.Printf("%s | removing %s track with StreamID %s", call.callID, sender.Track().Kind(), info.streamID)
					if err := sender.Stop(); err != nil {
						log.Printf("%s | failed to stop sender: %s", call.callID, err)
					}
					if err := call.peerConnection.RemoveTrack(sender); err != nil {
						log.Printf("%s | failed to remove track: %s", call.callID, err)
						return err
					}
				}
			}
		}
	}

	return nil
}
