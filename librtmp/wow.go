package librtmp

import (
	"bytes"
	"encoding/binary"
	"io"
	"math/rand"

	"github.com/icza/bitio"
	log "github.com/sirupsen/logrus"
)

type RtmpHandShakeHeader struct {
	// RandomBytes []byte // 1528 bytes
}
type RtmpHandShake1 struct {
	Version     uint8
	Timestamp   uint32
	Zero        uint32
	RandomBytes [1528]byte // 1526 bytes
}
type RtmpHandShake2 struct {
	Time1       uint32
	Time2       uint32
	RandomBytes [1528]byte // 1526 bytes
}

type BasicHeader struct {
	Fmt  int
	Csid int
}

type RtmpRecvHandler interface {
	Recv(bh BasicHeader, mh MessageHeader, body []byte)
	OnConnect()
}
type MessageTypeId uint8

const (
	SetChunkSize              MessageTypeId = 1
	AbortMessage              MessageTypeId = 2
	Acknowledgement           MessageTypeId = 3
	WindowAcknowledgementSize MessageTypeId = 5
	SetPeerBandwidth          MessageTypeId = 6
	Audio                     MessageTypeId = 8
	Video                     MessageTypeId = 9
	AFM0                      MessageTypeId = 20
	AFM3                      MessageTypeId = 17
	HandShakeC0Step                         = 0
	HandShakeC2Step                         = 1
	BasicHeaderStep                         = 2
	MessageHeaderStep                       = 3
	ChunkStep                               = 4
)

type MessageHeader struct {
	Timestamp       uint32
	MessageLength   uint32
	MessageTypeId   MessageTypeId //
	MessageStreamId uint32
}

type Stream struct {
	BasicHeader   BasicHeader
	MessageHeader MessageHeader
}
type RtmpContext struct {
	IsHandShake   bool
	ParseStep     int
	BasicHeader   BasicHeader
	MessageHeader MessageHeader
	RecvBuffer    bytes.Buffer
	HandShakeData RtmpHandShake1
	ChunkData     bytes.Buffer
	Handler       RtmpRecvHandler
	// Streams     map[int]Stream
	// BasicHeader BasicHeader
	// BasicHeader BasicHeader
}

func (k *RtmpContext) EncodeData(csid int, data []byte) []byte {
	// ret := make([]byte, 0)
	w := new(bytes.Buffer)
	fmt := 0
	bh := byte((fmt << 6) | (csid))
	binary.Write(w, binary.BigEndian, bh)
	// ret = append(ret, bh)
	mh := MessageHeader{}
	mh.Timestamp = 0
	mh.MessageTypeId = 5
	mh.MessageLength = 4
	mh.MessageStreamId = 0
	binary.Write(w, binary.BigEndian, mh)
	ws := uint32(2500000)
	binary.Write(w, binary.BigEndian, ws)
	return w.Bytes()
}
func (k *RtmpContext) Parse(b []byte, writer io.Writer) {
	k.RecvBuffer.Write(b)
	for k.RecvBuffer.Len() > 0 {
		// log.Printf("!!! step: %d\n", k.ParseStep)
		if k.ParseStep == HandShakeC0Step {
			// log.Println("!!!")
			if k.RecvBuffer.Len() >= 1537 {
				log.Println("!!! C0")
				shake := RtmpHandShake1{}
				err := binary.Read(&k.RecvBuffer, binary.LittleEndian, &shake)
				if err != nil {

				}
				sendShake := shake // librtmp.RtmpHandShake1{}
				for i := 0; i < len(sendShake.RandomBytes); i++ {
					sendShake.RandomBytes[i] = byte(rand.Intn(256))
				}
				err = binary.Write(writer, binary.LittleEndian, sendShake)
				S2 := RtmpHandShake1{}
				S2.Version = 3
				S2.Timestamp = sendShake.Timestamp
				// S2.Time2 = shake.Timestamp
				for i := 0; i < len(S2.RandomBytes); i++ {
					S2.RandomBytes[i] = byte(rand.Intn(256))
				}
				err = binary.Write(writer, binary.LittleEndian, S2)
				k.ParseStep = HandShakeC2Step
				log.Printf("********************** step2 remain: %d *********************\n", k.RecvBuffer.Len())
			} else {
				return
			}
		} else if k.ParseStep == HandShakeC2Step {
			// log.Println("!!!")
			if k.RecvBuffer.Len() >= 1536 {
				log.Println("!!! C2")
				shake := RtmpHandShake2{}
				err := binary.Read(&k.RecvBuffer, binary.LittleEndian, &shake)
				if err != nil {

				}
				sendShake := shake // librtmp.RtmpHandShake1{}
				for i := 0; i < len(sendShake.RandomBytes); i++ {
					sendShake.RandomBytes[i] = byte(rand.Intn(256))
				}
				err = binary.Write(writer, binary.LittleEndian, sendShake)
				S2 := RtmpHandShake2{}
				S2.Time1 = sendShake.Time1
				S2.Time2 = shake.Time1
				for i := 0; i < len(S2.RandomBytes); i++ {
					S2.RandomBytes[i] = byte(rand.Intn(256))
				}
				err = binary.Write(writer, binary.LittleEndian, S2)
				k.ParseStep = BasicHeaderStep
				log.Printf("********************** step2 remain: %d *********************\n", k.RecvBuffer.Len())
			} else {
				return
			}
		} else if k.ParseStep == BasicHeaderStep {
			if k.RecvBuffer.Len() >= 3 {
				r := bitio.NewReader(&k.RecvBuffer)
				{
					tmp, err := r.ReadBits(2) //   1100 = 0x08
					if err != nil {

					}
					k.BasicHeader.Fmt = int(tmp)
				}
				{
					tmp, err := r.ReadBits(6) //   1100 = 0x08
					if err != nil {

					}
					if tmp == 0 {
						// binary.Read(conn, binary.BigEndian, &k.RtmpContext.BasicHeader.Csid)
					} else if tmp == 1 {
					} else {
						k.BasicHeader.Csid = int(tmp)
					}
				}
				log.Printf("********************** step3 basicheader: %+v *********************\n", k.BasicHeader)
				log.Printf("********************** step3 remain: %d *********************\n", k.RecvBuffer.Len())
				k.ParseStep = MessageHeaderStep
			} else {
				return
			}
		} else if k.ParseStep == MessageHeaderStep {
			if k.BasicHeader.Fmt == 0 {
				if k.RecvBuffer.Len() >= 11 {
					r := make([]byte, 11) // 3 + 3 + 1 + 4
					_, err := k.RecvBuffer.Read(r)
					if err != nil {

					}
					k.MessageHeader.Timestamp = uint32(BigEndianToInt(r[0:3]))
					k.MessageHeader.MessageLength = uint32(BigEndianToInt(r[3:6]))
					k.MessageHeader.MessageTypeId = MessageTypeId(BigEndianToInt(r[6:7]))
					k.MessageHeader.MessageStreamId = uint32(BigEndianToInt(r[7:11]))
					k.ParseStep = ChunkStep
					log.Printf("********************** step4 message header: %+v *********************\n", k.MessageHeader)
					log.Printf("********************** step4 remain: %d *********************\n", k.RecvBuffer.Len())
				} else {
					return
				}
			} else if k.BasicHeader.Fmt == 1 {
				if k.RecvBuffer.Len() >= 7 {
					r := make([]byte, 7)
					_, err := k.RecvBuffer.Read(r)
					if err != nil {

					}
					k.MessageHeader.Timestamp = uint32(BigEndianToInt(r[0:3]))
					k.MessageHeader.MessageLength = uint32(BigEndianToInt(r[3:6]))
					k.MessageHeader.MessageTypeId = MessageTypeId(BigEndianToInt(r[6:7]))
					// k.MessageHeader.MessageStreamId = uint32(BigEndianToInt(r[7:12]))
					k.ParseStep = ChunkStep
					log.Printf("********************** step4 message header: %+v *********************\n", k.MessageHeader)
					log.Printf("********************** step4 remain: %d *********************\n", k.RecvBuffer.Len())
				} else {
					return
				}

			} else if k.BasicHeader.Fmt == 2 {
				if k.RecvBuffer.Len() >= 3 {
					r := make([]byte, 3)
					_, err := k.RecvBuffer.Read(r)
					if err != nil {

					}
					k.MessageHeader.Timestamp = uint32(BigEndianToInt(r[0:3]))
					k.ParseStep = ChunkStep
					log.Printf("********************** step4 message header: %+v *********************\n", k.MessageHeader)
					log.Printf("********************** step4 remain: %d *********************\n", k.RecvBuffer.Len())
				} else {
					return
				}

			} else if k.BasicHeader.Fmt == 3 {
				k.ParseStep = ChunkStep
			}
		} else if k.ParseStep == ChunkStep {
			if k.RecvBuffer.Len() >= int(k.MessageHeader.MessageLength) {
				tmp := make([]byte, k.MessageHeader.MessageLength)
				_, err := k.RecvBuffer.Read(tmp)
				if err != nil {
				}
				k.ChunkData.Write(tmp)
				k.Handler.Recv(k.BasicHeader, k.MessageHeader, k.ChunkData.Bytes())
				k.ChunkData.Reset()
				k.ParseStep = BasicHeaderStep
			}
		}

	}

}

// func (r* rtmpContext) 	Read(b []byte) (n int, err error) {

// }

// // Write writes data to the connection.
// // Write can be made to time out and return an error after a fixed
// // time limit; see SetDeadline and SetWriteDeadline.
// Write(b []byte) (n int, err error)

func Wow() {

}
