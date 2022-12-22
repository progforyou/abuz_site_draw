package gogate

import (
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCreateRequest(t *testing.T) {
	req := &GateRequest{
		Method: "GET",
		Url:    "/test/1.html",
		Header: []*GateHeader{
			{Key: "Content-Type", Values: []string{"plain/text"}},
			{Key: "Transfer-Encoding", Values: []string{"deflate"}},
		},
	}

	res, err := FromGateRequest(req)
	assert.Nil(t, err)
	log.Info().Interface("res", res.TransferEncoding).Msg("res")

}
