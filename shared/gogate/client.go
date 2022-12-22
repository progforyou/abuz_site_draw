package gogate

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/rs/zerolog/log"
	"google.golang.org/protobuf/proto"
	"net"
	"net/http"
	"strings"
	"time"
)

var pingInterval = time.Second * 20

type fListener func(request *GateRequest) (*GateResponse, error)

func NewClient(name string, gateHost string, gatePort int, listener fListener) error {
	log.Info().Str("name", name).Str("host", gateHost).Int("port", gatePort).Msg("start gate-client")
	tcpAddr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", gateHost, gatePort))
	if err != nil {
		return err
	}
	for {
		conn, err := net.DialTCP("tcp", nil, tcpAddr)
		if err != nil {
			log.Debug().Err(err).Msg("fail to create tcp-connection")
			time.Sleep(time.Millisecond * 100)
			continue
		}
		data, err := handshake(name)
		if err != nil {
			log.Debug().Err(err).Msg("fail to create handshake")
			continue
		}
		log.Info().Hex("pck", data).Msg("handshake")
		_, err = conn.Write(data)
		if err != nil {
			log.Debug().Err(err).Msg("fail to send data")
			time.Sleep(time.Millisecond * 100)
			continue
		}
		dataChannel := make(chan *Packet)
		go func() {
			for {
				pck, ok := <-dataChannel
				if !ok {
					log.Debug().Msg("channel closed")
					return
				}
				if pck.Requests != nil {
					resp, errt := listener(pck.Requests)
					if errt != nil {
						log.Error().Err(errt).Msg("error in listener")
						conn.Close()
						return
					}
					resp.Id = pck.Requests.Id
					respData, errt := createResponse(resp)
					if errt != nil {
						log.Error().Err(errt).Msg("error in create response")
						conn.Close()
						return
					}
					_, errt = conn.Write(respData)
					if errt != nil {
						log.Error().Err(errt).Msg("error send response")
						conn.Close()
						return
					}
				}
			}
		}()
		err = ReadL4VPacket(conn, dataChannel)
		close(dataChannel)
		if err != nil {
			log.Debug().Err(err).Msg("error in reader")
		}
		time.Sleep(time.Millisecond * 100)
	}
}

func ping(conn net.Conn) chan bool {
	tm := time.NewTicker(pingInterval)
	closeChan := make(chan bool, 1)
	go func() {
		defer close(closeChan)
		for {
			select {
			case <-closeChan:
				return
			case <-tm.C:
				pingPck := &Packet{}
				pingData, err := proto.Marshal(pingPck)
				if err != nil {
					break
				}
				_, err = conn.Write(addSize(pingData))
				if err != nil {
					break
				}
			}
		}
	}()
	return closeChan
}

func handshake(name string) ([]byte, error) {
	pck := &Packet{
		Handshake: &GateHandshake{
			Service: name,
		},
	}
	data, err := proto.Marshal(pck)
	if err != nil {
		return nil, err
	}
	return addSize(data), nil
}

func createResponse(resp *GateResponse) ([]byte, error) {
	pck := &Packet{
		Responses: resp,
	}
	data, err := proto.Marshal(pck)
	if err != nil {
		return nil, err
	}
	log.Debug().Int("size", len(data)).Hex("hex-size", getBytesFromInt32((int32)(len(data)))).Msg("send")
	return addSize(data), nil
}

func NewHttpClient(name string, gateHost string, gatePort int, requestHost string) error {
	tr := &http.Transport{
		MaxIdleConns:       10,
		IdleConnTimeout:    timeoutTTL,
		DisableCompression: true,
	}
	client := &http.Client{Transport: tr}
	if strings.HasSuffix(requestHost, "/") {
		requestHost = requestHost[:len(requestHost)-1]
	}
	return NewClient(name, gateHost, gatePort, func(request *GateRequest) (*GateResponse, error) {
		httpRequest, err := http.NewRequest(request.Method, fmt.Sprintf("%s%s", requestHost, request.Url), bytes.NewReader(request.Body))
		if err != nil {
			return nil, err
		}
		httpRequest.Header = FromGateHeader(request.Header)
		httpResponse, err := client.Do(httpRequest)
		if err != nil {
			return nil, err
		}
		resp, err := ToGateResponse(httpResponse)
		log.Debug().Err(err).Interface("req-header", httpRequest.Header).Interface("resp-header", resp.Header).Str("url", fmt.Sprintf("%s%s", requestHost, request.Url)).Int("size", len(resp.Body)).Msg("body")
		return resp, err
	})
}

type ResponseWriter struct {
	body   []byte
	code   int
	header http.Header
}

func (c *ResponseWriter) Header() http.Header { return c.header }

func (c *ResponseWriter) Write(data []byte) (int, error) {
	c.body = append(c.body, data...)
	return len(data), nil
}

func (c *ResponseWriter) WriteHeader(statusCode int) {
	c.code = statusCode
}

func NewHTTPHandlerClient(name string, gateHost string, gatePort int, handler http.Handler) error {
	return NewClient(name, gateHost, gatePort, func(request *GateRequest) (*GateResponse, error) {
		if handler == nil {
			return nil, errors.New("handler is nil")
		}
		wr := &ResponseWriter{
			header: http.Header{},
			code:   200,
		}
		httpRequest, err := FromGateRequest(request)
		if err != nil {
			return nil, err
		}
		handler.ServeHTTP(wr, httpRequest)
		resp := &GateResponse{
			StatusCode:    int32(wr.code),
			Body:          wr.body,
			ContentLength: int64(len(wr.body)),
			Header:        ToGateHeader(wr.header),
		}
		return resp, nil
	})
}

func ReadL4VPacket(conn net.Conn, result chan *Packet) error {
	dataChannel := make(chan []byte)
	closeChan := ping(conn)
	go func() {
		for {
			res, ok := <-dataChannel
			if !ok {
				return
			}
			var pck Packet
			err := proto.Unmarshal(res, &pck)
			if err != nil {
				conn.Close()
				return
			}
			result <- &pck
		}
	}()

	err := readL4VData(conn, dataChannel)
	closeChan <- true
	close(dataChannel)
	return err
}
