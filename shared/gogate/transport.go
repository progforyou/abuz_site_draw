package gogate

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"google.golang.org/protobuf/proto"
	"net"
	"sync"
	"time"
)

var (
	serviceChannel            = map[string]*connectionHolder{}
	serviceChannelLock        = sync.RWMutex{}
	requestId          uint64 = 0
	holder                    = map[uint64]*requestHolder{}
	holderLock                = sync.RWMutex{}
	timeoutTTL                = time.Second * 60
)

type connectionHolder struct {
	conn    net.Conn
	reqChan chan *requestHolder
}

func newConnectionHolder(conn net.Conn) *connectionHolder {
	c := &connectionHolder{
		conn:    conn,
		reqChan: make(chan *requestHolder),
	}
	go func() {
		for {
			req, ok := <-c.reqChan
			if !ok {
				return
			}
			data, err := proto.Marshal(&Packet{Requests: req.req})
			if err != nil {
				return
			}
			holderLock.Lock()
			log.Debug().Uint64("id", req.req.Id).Msg("set holder")
			holder[req.req.Id] = req
			holderLock.Unlock()
			_, err = conn.Write(addSize(data))
			if err != nil {
				return
			}
		}
	}()
	return c
}

func (c *connectionHolder) Close() {
	close(c.reqChan)
	_ = c.conn.Close()
}

func NewTransport(host string, port int) error {
	l, err := net.Listen("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		return err
	}
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Error().Err(err).Msg("error accepting connection")
			break
		}
		log.Debug().Str("remote-addr", conn.RemoteAddr().String()).Msg("new connection")
		go dataReader(conn)
	}
	return nil
}

func dataReader(conn net.Conn) {
	defer conn.Close()
	logg := log.With().Str("remote-addr", conn.RemoteAddr().String()).Logger()
	err := conn.SetReadDeadline(time.Now().Add(time.Second * 60))
	if err != nil {
		logg.Error().Err(err).Msg("set timeout error")
		return
	}
	dataChannel := make(chan *Packet)
	go func() {
		cHolder := newConnectionHolder(conn)
		defer cHolder.Close()
		for {
			pck, ok := <-dataChannel
			if !ok {
				logg.Debug().Msg("channel closed")
				return
			}
			_ = conn.SetReadDeadline(time.Now().Add(timeoutTTL))
			switch {
			case pck.Handshake != nil && pck.Handshake.Service != "":
				logg = logg.With().Str("service", pck.Handshake.Service).Logger()
				logg.Debug().Msg("handshake")
				serviceChannelLock.Lock()
				if oldHolder, okc := serviceChannel[pck.Handshake.Service]; okc {
					oldHolder.conn.Close()
				}
				serviceChannel[pck.Handshake.Service] = cHolder
				serviceChannelLock.Unlock()
				break
			case pck.Responses != nil:
				holderLock.Lock()
				holded, ok := holder[pck.Responses.Id]
				if ok {
					delete(holder, pck.Responses.Id)
					holded.resp <- pck.Responses
				}
				holderLock.Unlock()
				break
			}
		}
	}()
	err = ReadL4VPacket(conn, dataChannel)
	close(dataChannel)
	if err != nil {
		logg.Debug().Err(err).Msg("error in reader")
	}
}
