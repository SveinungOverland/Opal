package opal

import (
	"crypto/tls"
	"fmt"
	"net"
	"opal/hpack"

	"opal/frame"
	"opal/frame/types"
)

type Conn struct {
	server        *Server
	conn          net.Conn
	tlsConn       *tls.Conn
	hpack         *hpack.Context
	isTLS         bool
	maxConcurrent uint32
	streams       map[uint32]Stream // map streamId to Stream instance
}

func (c *Conn) serve() {
	// start := time.Now() // Request timer

	// Initialize TLS handshake
	if c.isTLS {
		err := c.tlsConn.Handshake()
		if err != nil {
			fmt.Println(err)
			c.isTLS = false
			c.tlsConn.Close()
			return
		}
	}

	prefaceBuffer := make([]byte, 24)
	c.tlsConn.Read(prefaceBuffer)
	fmt.Println(string(prefaceBuffer))

	settingsFrame := frame.ReadFrame(c.tlsConn)
	fmt.Printf("%+v\n", settingsFrame)

	c.hpack = hpack.NewContext(settingsFrame.Payload.(*types.SettingsPayload).IDValuePair[1])

	// TODO: Change actual settings based on the frame above
	settingsResponse := &frame.Frame{
		ID:     0,
		Type:   frame.SettingsType,
		Length: 0,
		Flags: &types.SettingsFlags{
			Ack: true,
		},
	}
	// TODO: Write settingsResponse to client to acknowledge settings frame
	settingsFrameBytes := settingsResponse.ToBytes()
	fmt.Println(settingsFrameBytes)
	c.tlsConn.Write(settingsFrameBytes)

	windowUpdateFrame := frame.ReadFrame(c.tlsConn)
	fmt.Printf("%+v\n", windowUpdateFrame)

	headersFrame := frame.ReadFrame(c.tlsConn)
	fmt.Printf("%+v\n", headersFrame.Flags.(*types.HeadersFlags))

	s := &Stream{
		id:      headersFrame.ID,
		headers: make([]*types.HeadersPayload, 0),
	}
	s.headers = append(s.headers, headersFrame.Payload.(*types.HeadersPayload))

	req, err := s.Build(c.hpack)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(req)

	fmt.Println(c.hpack.Decode((headersFrame.Payload.(*types.HeadersPayload).Fragment)))
}