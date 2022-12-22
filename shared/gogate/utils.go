package gogate

import (
	"bytes"
	"encoding/binary"
	"io"
	"io/ioutil"
	"net"
	"net/http"
)

func ToGateHeader(header http.Header) []*GateHeader {
	var res []*GateHeader
	for k, v := range header {
		res = append(res, &GateHeader{
			Key:    k,
			Values: v,
		})
	}
	return res
}

func FromGateHeader(header []*GateHeader) http.Header {
	res := http.Header{}
	for _, h := range header {
		res[h.Key] = h.Values
	}
	return res
}

func ToGateRequest(req *http.Request) (*GateRequest, error) {
	res := &GateRequest{
		Method:        req.Method,
		Url:           req.RequestURI,
		Header:        ToGateHeader(req.Header),
		Host:          req.Host,
		RemoteAddr:    req.RemoteAddr,
		ContentLength: req.ContentLength,
	}

	switch req.Method {
	case "POST", "PUT", "PATCH":
		bodyBytes, err := ioutil.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}
		res.Body = bodyBytes
	}
	return res, nil
}

func FromGateRequest(req *GateRequest) (*http.Request, error) {
	res, err := http.NewRequest(req.Method, req.Url, bytes.NewReader(req.Body))
	if err != nil {
		return nil, err
	}
	res.Header = FromGateHeader(req.Header)
	return res, nil
}

func ToGateResponse(req *http.Response) (*GateResponse, error) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}

	res := &GateResponse{
		StatusCode:    int32(req.StatusCode),
		ContentLength: req.ContentLength,
		Header:        ToGateHeader(req.Header),
		Body:          body,
	}
	return res, nil
}

func addSize(data []byte) []byte {
	ld := len(data)
	return append(getBytesFromInt32((int32)(ld)), data...)
}

func getInt32FromBytes(lens []byte) int {
	return int(binary.LittleEndian.Uint32(lens))
}

func getBytesFromInt32(len int32) []byte {
	bs := make([]byte, 4)
	binary.LittleEndian.PutUint32(bs, uint32(len))
	return bs
}

func readL4VData(conn net.Conn, result chan []byte) error {
	return readL4VDataBuf(conn, result, 4096)
}

func readL4VDataBuf(conn net.Conn, result chan []byte, bufSize int) error {
	buf := make([]byte, bufSize)
	var data []byte
	for {
		readSize, err := conn.Read(buf)
		if err != nil {
			return err
		}
		data = append(data, buf[:readSize]...)
		for {
			ld := len(data)
			if ld >= 4 {
				l4 := getInt32FromBytes(data[:4])
				if ld >= l4+4 {
					result <- data[4 : l4+4]
					data = data[l4+4:]
				} else {
					break
				}
			} else {
				break
			}
		}
	}
}
