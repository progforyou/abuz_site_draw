package gogate

import (
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
	"net"
	"testing"
	"time"
)

func TestNewTransport(t *testing.T) {
	go func() {
		err := NewTransport("localhost", 9798)
		if err != nil {
			panic(err)
		}
	}()
	tcpAddr, err := net.ResolveTCPAddr("tcp", "localhost:9798")
	assert.Nil(t, err)
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	assert.Nil(t, err)
	data, err := proto.Marshal(&Packet{
		Handshake: &GateHandshake{
			Service: "test-service-" + (string)(make([]byte, 111)),
		},
	})
	assert.Nil(t, err)
	_, err = conn.Write(addSize(data))
	log.Debug().Int("size", len(data)).Msg("send size")
	assert.Nil(t, err)
	time.Sleep(time.Millisecond * 100)
}

func TestTail(t *testing.T) {
	data := []byte{0, 1, 2, 3, 4, 5}
	size := 4
	tail := data[size:]
	data = data[:size]
	assert.EqualValues(t, data, []byte{0, 1, 2, 3})
	assert.EqualValues(t, tail, []byte{4, 5})
}
