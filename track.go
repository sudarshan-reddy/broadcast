package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/pion/rtcp"
	"github.com/pion/webrtc/v2"
)

func track(sdp webrtc.SessionDescription,
	sdpCh chan webrtc.SessionDescription,
	api *webrtc.API,
	peerConnection *webrtc.PeerConnection) error {
	localTrackCh := make(chan *webrtc.Track)

	peerConnection.OnTrack(func(remoteTrack *webrtc.Track, receiver *webrtc.RTPReceiver) {
		go func() {
			ticker := time.NewTicker(3 * time.Second)
			for range ticker.C {
				if rtcpSendErr := peerConnection.
					WriteRTCP([]rtcp.Packet{&rtcp.PictureLossIndication{MediaSSRC: remoteTrack.SSRC()}}); rtcpSendErr != nil {
					fmt.Println(rtcpSendErr)
				}
			}
		}()

		localTrack, err := peerConnection.NewTrack(remoteTrack.PayloadType(), remoteTrack.SSRC(), "video", "test")
		if err != nil {
			panic(err)
		}

		localTrackCh <- localTrack

		rtpBuf := make([]byte, 1400)
		for {
			i, err := remoteTrack.Read(rtpBuf)
			if err != nil {
				panic(err)
			}

			// ErrClosedPipe means we don't have any subscribers, this is ok if no peers have connected yet
			if _, err = localTrack.Write(rtpBuf[:i]); err != nil && err != io.ErrClosedPipe {
				panic(err)
			}
		}
	})

	if err := peerConnection.SetRemoteDescription(sdp); err != nil {
		return err
	}

	answer, err := peerConnection.CreateAnswer(nil)
	if err != nil {
		return err
	}

	if err = peerConnection.SetLocalDescription(answer); err != nil {
		panic(err)
	}

	fmt.Println(encode(answer))

	localTrack := <-localTrackCh

	for sdp := range sdpCh {
		fmt.Println("")
		fmt.Println("Curl an base64 SDP to start sendonly peer connection")

		// Create a new PeerConnection
		peerConnection, err := api.NewPeerConnection(getPeerConfig(googleStunServer))
		if err != nil {
			panic(err)
		}

		_, err = peerConnection.AddTrack(localTrack)
		if err != nil {
			panic(err)
		}

		// Set the remote SessionDescription
		err = peerConnection.SetRemoteDescription(sdp)
		if err != nil {
			panic(err)
		}

		// Create answer
		answer, err := peerConnection.CreateAnswer(nil)
		if err != nil {
			panic(err)
		}

		// Sets the LocalDescription, and starts our UDP listeners
		err = peerConnection.SetLocalDescription(answer)
		if err != nil {
			panic(err)
		}

		fmt.Println(encode(answer))

	}

	return nil

}

func encode(obj interface{}) string {
	b, err := json.Marshal(obj)
	if err != nil {
		panic(err)
	}

	return base64.StdEncoding.EncodeToString(b)
}
