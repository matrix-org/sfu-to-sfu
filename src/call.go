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
	"encoding/json"
	"log"
	"strings"
	"time"

	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"

	"github.com/pion/rtcp"
	"github.com/pion/webrtc/v3"
)

type call struct {
	callID          string
	userID          id.UserID
	deviceID        id.DeviceID
	localSessionID  string
	remoteSessionID string
	client          *mautrix.Client
	peerConnection  *webrtc.PeerConnection
	conf            *conf
	dataChannel     *webrtc.DataChannel
}

func (c *call) dataChannelHandler(d *webrtc.DataChannel) {
	c.dataChannel = d
	peerConnection := c.peerConnection

	d.OnOpen(func() {
		log.Printf("%s | DC opened", c.callID)
	})

	d.OnClose(func() {
		log.Printf("%s | DC closed", c.callID)
	})

	d.OnError(func(err error) {
		log.Fatalf("%s | DC error: %s", c.callID, err)
	})

	d.OnMessage(func(m webrtc.DataChannelMessage) {
		if !m.IsString {
			log.Fatal("Inbound message is not string")
		}

		msg := &dataChannelMessage{}
		if err := json.Unmarshal(m.Data, msg); err != nil {
			log.Fatalf("%s | failed to unmarshal: %s", c.callID, err)
		}

		log.Printf("%s | received DC %s confId=%s start=%+v", c.callID, msg.Op, msg.ConfID, msg.Start)

		// TODO: hook cascade back up.
		// As we're not an AS, we'd rely on the client
		// to send us a "connect" op to tell us how to
		// connect to another focus in order to select
		// its streams.

		switch msg.Op {
		case "select":
			var tracks []webrtc.TrackLocal
			for _, trackDesc := range msg.Start {
				foundTracks, err := c.conf.getLocalTrackByInfo(localTrackInfo{streamID: trackDesc.StreamID})
				if err != nil {
					c.sendDataChannelError("No Such Stream")
					return
				} else {
					tracks = append(tracks, foundTracks...)
				}
			}

			for _, track := range tracks {
				log.Printf("%s | adding %s track with StreamID %s", c.callID, track.Kind(), track.StreamID())
				if _, err := peerConnection.AddTrack(track); err != nil {
					panic(err)
				}
			}

		case "publish":
			peerConnection.SetRemoteDescription(webrtc.SessionDescription{
				Type: webrtc.SDPTypeOffer,
				SDP:  msg.SDP,
			})

			offer, err := c.peerConnection.CreateAnswer(nil)
			if err != nil {
				panic(err)
			}
			err = c.peerConnection.SetLocalDescription(offer)
			if err != nil {
				panic(err)
			}

			c.sendDataChannelMessage(dataChannelMessage{
				Op:  "answer",
				SDP: offer.SDP,
			})

		case "answer":
			peerConnection.SetRemoteDescription(webrtc.SessionDescription{
				Type: webrtc.SDPTypeAnswer,
				SDP:  msg.SDP,
			})

		default:
			log.Fatalf("Unknown operation %s", msg.Op)
			// TODO: hook up msg.Stop to unsubscribe from tracks
		}
	})
}

func (c *call) negotiationNeededHandler() {
	log.Printf("%s | negotiation needed", c.callID)

	offer, err := c.peerConnection.CreateOffer(nil)
	if err != nil {
		panic(err)
	}
	err = c.peerConnection.SetLocalDescription(offer)
	if err != nil {
		panic(err)
	}

	c.sendDataChannelMessage(dataChannelMessage{
		Op:  "offer",
		SDP: offer.SDP,
	})
}

func (c *call) iceCandidateHandler(candidate *webrtc.ICECandidate) {
	if candidate == nil {
		return
	}

	ice := candidate.ToJSON()

	log.Printf("%s | discovered local candidate %s", c.callID, ice.Candidate)

	// TODO: batch these up a bit
	candidateEvtContent := &event.Content{
		Parsed: event.CallCandidatesEventContent{
			BaseCallEventContent: event.BaseCallEventContent{
				CallID:          c.callID,
				ConfID:          c.conf.confID,
				DeviceID:        c.client.DeviceID,
				SenderSessionID: c.localSessionID,
				DestSessionID:   c.remoteSessionID,
				PartyID:         string(c.client.DeviceID),
				Version:         event.CallVersion("1"),
			},
			Candidates: []event.CallCandidate{
				{
					Candidate:     ice.Candidate,
					SDPMLineIndex: int(*ice.SDPMLineIndex),
					SDPMID:        *ice.SDPMid,
					// XXX: what about ice.UsernameFragment?
				},
			},
		},
	}
	c.sendToDevice(event.CallCandidates, candidateEvtContent)
}

func (c *call) trackHandler(trackRemote *webrtc.TrackRemote, rec *webrtc.RTPReceiver) {
	log.Printf("%s | discovered track with streamID %s and kind %s", c.callID, trackRemote.StreamID(), trackRemote.Kind())
	if strings.Contains(trackRemote.Codec().MimeType, "video") {
		// Send a PLI on an interval so that the publisher is pushing a keyframe every rtcpPLIInterval
		go func() {
			ticker := time.NewTicker(time.Millisecond * 200)
			for range ticker.C {
				if err := c.peerConnection.WriteRTCP([]rtcp.Packet{&rtcp.PictureLossIndication{MediaSSRC: uint32(trackRemote.SSRC())}}); err != nil {
					log.Printf("%s | failed to write RTCP on trackID %s: %s", c.callID, trackRemote.ID(), err)
					break
				}
			}
		}()
	}

	c.conf.tracksMu.Lock()
	trackLocal, err := webrtc.NewTrackLocalStaticRTP(trackRemote.Codec().RTPCodecCapability, trackRemote.ID(), trackRemote.StreamID())
	if err != nil {
		panic(err)
	}

	c.conf.tracks = append(c.conf.tracks, localTrackWithInfo{
		track: trackLocal,
		info: localTrackInfo{
			trackID:  trackLocal.ID(),
			streamID: trackLocal.StreamID(),
			call:     c,
		},
	})

	log.Printf("%s | published track with streamID %s and kind %s", c.callID, trackLocal.StreamID(), trackLocal.Kind())
	c.conf.tracksMu.Unlock()

	copyRemoteToLocal(trackRemote, trackLocal)
}

func (c *call) onInvite(content *event.CallInviteEventContent) error {
	offer := content.Offer

	peerConnection, err := webrtc.NewPeerConnection(webrtc.Configuration{})
	if err != nil {
		return err
	}
	c.peerConnection = peerConnection

	peerConnection.OnTrack(func(track *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
		c.trackHandler(track, receiver)
	})
	peerConnection.OnDataChannel(func(d *webrtc.DataChannel) {
		c.dataChannelHandler(d)
	})
	peerConnection.OnICECandidate(func(candidate *webrtc.ICECandidate) {
		c.iceCandidateHandler(candidate)
	})
	peerConnection.OnNegotiationNeeded(func() {
		c.negotiationNeededHandler()
	})

	peerConnection.SetRemoteDescription(webrtc.SessionDescription{
		Type: webrtc.SDPTypeOffer,
		SDP:  offer.SDP,
	})

	answer, err := peerConnection.CreateAnswer(nil)
	if err != nil {
		return err
	}

	// TODO: trickle ICE for fast conn setup, rather than block here
	gatherComplete := webrtc.GatheringCompletePromise(peerConnection)
	if err = peerConnection.SetLocalDescription(answer); err != nil {
		return err
	}
	<-gatherComplete

	answerSdp := peerConnection.LocalDescription().SDP

	answerEvtContent := &event.Content{
		Parsed: event.CallAnswerEventContent{
			BaseCallEventContent: event.BaseCallEventContent{
				CallID:          c.callID,
				ConfID:          c.conf.confID,
				DeviceID:        c.client.DeviceID,
				SenderSessionID: c.localSessionID,
				DestSessionID:   c.remoteSessionID,
				PartyID:         string(c.client.DeviceID),
				Version:         event.CallVersion("1"),
			},
			Answer: event.CallData{
				Type: "answer",
				SDP:  answerSdp,
			},
		},
	}
	c.sendToDevice(event.CallAnswer, answerEvtContent)

	return err
}

func (c *call) onSelectAnswer(content *event.CallSelectAnswerEventContent) {
	selectedPartyId := content.SelectedPartyID
	if selectedPartyId != string(c.client.DeviceID) {
		c.terminate()
		log.Printf("%s | Call was answered on a different device: %s", content.CallID, selectedPartyId)
	}
}

func (c *call) onHangup(content *event.CallHangupEventContent) {
	c.terminate()
}

func (c *call) onCandidates(content *event.CallCandidatesEventContent) error {
	for _, candidate := range content.Candidates {
		sdpMLineIndex := uint16(candidate.SDPMLineIndex)
		ice := webrtc.ICECandidateInit{
			Candidate:        candidate.Candidate,
			SDPMid:           &candidate.SDPMID,
			SDPMLineIndex:    &sdpMLineIndex,
			UsernameFragment: new(string),
		}
		if err := c.peerConnection.AddICECandidate(ice); err != nil {
			log.Print("Failed to add ICE candidate", content)
			return err
		}
	}
	return nil
}

func (c *call) terminate() error {
	log.Printf("%s | Terminating call", c.callID)

	if err := c.peerConnection.Close(); err != nil {
		log.Printf("%s | error closing peer connection: %s", c.callID, err)
	}

	c.conf.calls.callsMu.Lock()
	delete(c.conf.calls.calls, c.callID)
	c.conf.calls.callsMu.Unlock()

	if err := c.conf.removeTracksFromPeerConnectionsByInfo(localTrackInfo{call: c}); err != nil {
		return err
	}

	// TODO: Remove the tracks from conf.tracks

	return nil
}

func (c *call) sendToDevice(callType event.Type, content *event.Content) error {
	log.Printf("%s | sending to device %s", c.callID, callType.Type)
	toDevice := &mautrix.ReqSendToDevice{
		Messages: map[id.UserID]map[id.DeviceID]*event.Content{
			c.userID: {
				c.deviceID: content,
			},
		},
	}

	// TODO: E2EE
	// TODO: to-device reliability
	c.client.SendToDevice(callType, toDevice)

	return nil
}

func (c *call) sendDataChannelMessage(msg dataChannelMessage) {
	msg.ConfID = c.conf.confID
	// TODO: Set ID

	marshaled, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}

	err = c.dataChannel.SendText(string(marshaled))
	if err != nil {
		log.Printf("%s | failed to send over DC: %s", c.callID, err)
	}

	log.Printf("%s | sent DC %s", c.callID, msg.Op)
}

func (c *call) sendDataChannelError(errMsg string) {
	log.Printf("%s | sending DC error %s", c.callID, errMsg)
	marshaled, err := json.Marshal(&dataChannelMessage{
		Op:      "error",
		Message: errMsg,
	})
	if err != nil {
		panic(err)
	}

	if err = c.dataChannel.SendText(string(marshaled)); err != nil {
		panic(err)
	}
}
