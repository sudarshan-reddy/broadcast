package main

import (
	"log"
	"net/http"

	"github.com/pion/webrtc/v2"
)

const googleStunServer = "stun:stun.l.google.com:19302"

func main() {

	sdpCh := make(chan webrtc.SessionDescription)

	m := http.NewServeMux()
	h := newHandler(sdpCh)

	m.HandleFunc("/sdp", h.sdp)

	go func() {
		if err := http.ListenAndServe(":1111", m); err != nil {
			log.Fatal(err)
		}
	}()

	sdp := <-sdpCh
	me, err := makeMediaEngine(sdp)
	if err != nil {
		panic(err)
	}

	api := webrtc.NewAPI(webrtc.WithMediaEngine(*me))

	peerConnection, err := api.NewPeerConnection(getPeerConfig(googleStunServer))

	if err != nil {
		panic(err)
	}

	if _, err = peerConnection.AddTransceiverFromKind(webrtc.RTPCodecTypeVideo); err != nil {
		panic(err)
	}

	track(sdp, sdpCh, api, peerConnection)

}

func getPeerConfig(stunServer string) webrtc.Configuration {
	return webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{stunServer},
			},
		},
	}
}

func makeMediaEngine(sdp webrtc.SessionDescription) (*webrtc.MediaEngine, error) {
	mediaEngine := webrtc.MediaEngine{}

	if err := mediaEngine.PopulateFromSDP(sdp); err != nil {
		return nil, err
	}

	return &mediaEngine, nil
}
