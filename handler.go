package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/pion/webrtc/v2"
	"github.com/sirupsen/logrus"
)

type handler struct {
	sdpChannel chan webrtc.SessionDescription
}

func newHandler(sdpCh chan webrtc.SessionDescription) *handler {
	return &handler{
		sdpChannel: sdpCh,
	}

}

func (h *handler) sdp(w http.ResponseWriter, r *http.Request) {
	fmt.Println("here?")
	body, _ := ioutil.ReadAll(r.Body)

	b, err := base64.StdEncoding.DecodeString(string(body))
	if err != nil {
		logrus.Error(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	obj := webrtc.SessionDescription{}
	if err := json.Unmarshal(b, &obj); err != nil {
		logrus.Error(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("done"))

	h.sdpChannel <- obj
}
