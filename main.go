package main

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"net"

	log "github.com/sirupsen/logrus"

	"github.com/hsnks100/librtmp"
	"github.com/hsnks100/tcpengine"
	"github.com/yutopp/go-amf0"
)

type kHandler struct {
	RecvBuffer  bytes.Buffer
	step        int
	RtmpContext librtmp.RtmpContext
}

func (k *kHandler) Recv(conn net.Conn, b []byte) {
	fmt.Println("recv: ", hex.Dump(b))
	k.RecvBuffer.Write(b)
	k.RtmpContext.Parse(b, conn)
}
func (k *kHandler) OnConnect(conn net.Conn) {
	fmt.Println("onConnect")
	k.step = 1
	// k.RecvBuffer = bufio.NewReader(conn)

}
func (k *kHandler) OnClose(conn net.Conn) {
	fmt.Println("onClose")
}

type rHandler struct {
}

func (r *rHandler) Recv(bh librtmp.BasicHeader, mh librtmp.MessageHeader, body []byte) {
	log.Infof("rtmp Recv, %+v / %+v", bh, mh)

	switch mh.MessageTypeId {
	case librtmp.SetChunkSize:
		log.Infof("SetChunkSize: %s", hex.Dump(body))

	case librtmp.AFM0:
		buf := new(bytes.Buffer)
		buf.Write(body)
		log.Infof("body: %s", hex.Dump(body))
		dec := amf0.NewDecoder(buf)
		for {
			var object interface{}
			// var v interface{}

			if err := dec.Decode(&object); err != nil {
				break
			}
			log.Infof("AMF0: %+v", object)
		}
		// log.Infof("AFM0: %s", hex.Dump(body))
	}
}
func (r *rHandler) OnConnect() {
	log.Infof("rtmp OnConnect")
}

func main() {
	log.SetReportCaller(true)
	librtmp.Wow()
	kh := &kHandler{}
	kh.RtmpContext.Handler = &rHandler{}
	te := tcpengine.NewTcpEngine(kh)
	te.BufSize = 2000
	te.Listen(1935)
}
