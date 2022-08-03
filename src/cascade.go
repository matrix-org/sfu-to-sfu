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

/*
import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/pion/webrtc/v3"
)

// Given a FOCI + CallID + DeviceID + Purpose establish a session and Subscribe. Take
// the media from the remote and copy it to a `webrtc.TrackLocal` so we can re-send
func remoteStreamLookup(msg dataChannelMessage) (webrtc.TrackLocal, webrtc.TrackLocal) {
	audioTrack, videoTrack := make(chan webrtc.TrackLocal, 1), make(chan webrtc.TrackLocal, 1)

	peerConnection, err := webrtc.NewPeerConnection(webrtc.Configuration{})
	if err != nil {
		panic(err)
	}

	peerConnection.OnTrack(func(trackRemote *webrtc.TrackRemote, _ *webrtc.RTPReceiver) {
		trackLocal, err := webrtc.NewTrackLocalStaticRTP(trackRemote.Codec().RTPCodecCapability, trackRemote.ID(), trackRemote.StreamID())
		if err != nil {
			panic(err)
		}

		if strings.Contains(trackRemote.Codec().MimeType, "video") {
			videoTrack <- trackLocal
		} else {
			audioTrack <- trackLocal
		}

		copyRemoteToLocal(trackRemote, trackLocal)
	})

	dataChannel, err := peerConnection.CreateDataChannel("signaling", nil)
	if err != nil {
		panic(err)
	}

	dataChannel.OnOpen(func() {
		if _, err := peerConnection.AddTransceiverFromKind(webrtc.RTPCodecTypeVideo, webrtc.RTPTransceiverInit{Direction: webrtc.RTPTransceiverDirectionRecvonly}); err != nil {
			panic(err)
		}

		if _, err := peerConnection.AddTransceiverFromKind(webrtc.RTPCodecTypeAudio, webrtc.RTPTransceiverInit{Direction: webrtc.RTPTransceiverDirectionRecvonly}); err != nil {
			panic(err)
		}

		offer, err := peerConnection.CreateOffer(nil)
		if err != nil {
			panic(err)
		}

		if err := peerConnection.SetLocalDescription(offer); err != nil {
			panic(err)
		}

		msg.SDP = offer.SDP
		marshaled, err := json.Marshal(msg)
		if err != nil {
			panic(err)
		}

		if err = dataChannel.SendText(string(marshaled)); err != nil {
			panic(err)
		}
	})

	dataChannel.OnMessage(func(m webrtc.DataChannelMessage) {
		if !m.IsString {
			log.Fatal("Inbound message is not string")
		}

		cascadedMsg := &dataChannelMessage{}
		if err := json.Unmarshal(m.Data, cascadedMsg); err != nil {
			log.Fatal(err)
		}

		switch cascadedMsg.Event {
		case "error":
			audioTrack <- nil
			videoTrack <- nil
		case "subscribe":
			if err := peerConnection.SetRemoteDescription(webrtc.SessionDescription{Type: webrtc.SDPTypeAnswer, SDP: cascadedMsg.SDP}); err != nil {
				panic(err)
			}

		default:
			log.Fatalf("Unknown msg Event type %s", msg.Event)
		}
	})

	offer, err := peerConnection.CreateOffer(nil)
	if err != nil {
		panic(err)
	}

	if err := peerConnection.SetLocalDescription(offer); err != nil {
		panic(err)
	}

	resp, err := http.Post("http://"+msg.FOCI+"/createSession", "application/text", bytes.NewBuffer([]byte(offer.SDP)))
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		panic(fmt.Sprintf("Got HTTP Status code %d", resp.StatusCode))
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	if err := peerConnection.SetRemoteDescription(webrtc.SessionDescription{Type: webrtc.SDPTypeAnswer, SDP: string(body)}); err != nil {
		panic(err)
	}

	return <-audioTrack, <-videoTrack

}
*/
