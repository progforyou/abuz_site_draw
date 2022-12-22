package sshtunnel

import (
	"fmt"
	"golang.org/x/crypto/ssh"
	"io"
	"io/ioutil"
	"log"
	"net"
	"strconv"
	"strings"
)

type Endpoint struct {
	Host string
	Port int
	User string
}

func NewEndpoint(s string) *Endpoint {
	endpoint := &Endpoint{
		Host: s,
	}
	if parts := strings.Split(endpoint.Host, "@"); len(parts) > 1 {
		endpoint.User = parts[0]
		endpoint.Host = parts[1]
	}
	if parts := strings.Split(endpoint.Host, ":"); len(parts) > 1 {
		endpoint.Host = parts[0]
		endpoint.Port, _ = strconv.Atoi(parts[1])
	}
	return endpoint
}

func (endpoint *Endpoint) String() string {
	return fmt.Sprintf("%s:%d", endpoint.Host, endpoint.Port)
}

type SSHTunnel struct {
	Local    *Endpoint
	Server   *Endpoint
	Remote   *Endpoint
	Config   *ssh.ClientConfig
	listener *net.Listener
}

func (tunnel *SSHTunnel) Start() {
	tunnel.Local.Port = (*tunnel.listener).Addr().(*net.TCPAddr).Port
	go func() {
		tunnel.Local.Port = (*tunnel.listener).Addr().(*net.TCPAddr).Port
		for {
			conn, err := (*tunnel.listener).Accept()
			if err != nil {
				log.Println("error accepting connection: ", err)
				return
			}
			log.Println("local connection accepted: ", conn)
			go tunnel.forward(conn)
		}
	}()
}

func (tunnel *SSHTunnel) Stop() {
	(*tunnel.listener).Close()
}

func (tunnel *SSHTunnel) forward(localConn net.Conn) {
	// defer localConn.Close()
	serverConn, err := ssh.Dial("tcp", tunnel.Server.String(), tunnel.Config)
	if err != nil {
		log.Printf("server dial error: %s\n", err)
		return
	}

	log.Printf("connected to %s (1 of 2)\n", tunnel.Server.String())
	remoteConn, err := serverConn.Dial("tcp", tunnel.Remote.String())
	if err != nil {
		log.Printf("remote dial error: %s\n", err)
		return
	}

	log.Printf("connected to %s (2 of 2)\n", tunnel.Remote.String())
	copyConn := func(writer, reader net.Conn) {
		_, err := io.Copy(writer, reader)
		if err != nil {
			log.Printf("io.Copy error: %s", err)
		}
	}

	go copyConn(localConn, remoteConn)
	go copyConn(remoteConn, localConn)
}

func PrivateKeyFile(file string) ssh.AuthMethod {
	buffer, err := ioutil.ReadFile(file)
	if err != nil {
		return nil
	}
	key, err := ssh.ParsePrivateKey(buffer)
	if err != nil {
		return nil
	}
	return ssh.PublicKeys(key)
}

func NewSSHTunnel(tunnel string, auth ssh.AuthMethod, destination string) (*SSHTunnel, error) {
	// A random port will be chosen for us.
	localEndpoint := NewEndpoint("localhost:0")

	server := NewEndpoint(tunnel)
	if server.Port == 0 {
		server.Port = 22
	}

	sshTunnel := &SSHTunnel{
		Config: &ssh.ClientConfig{
			User: server.User,
			Auth: []ssh.AuthMethod{auth},
			HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
				// Always accept key.
				return nil
			},
		},
		Local:  localEndpoint,
		Server: server,
		Remote: NewEndpoint(destination),
	}

	listener, err := net.Listen("tcp", sshTunnel.Local.String())
	if err != nil {
		return nil, err
	}
	sshTunnel.listener = &listener

	return sshTunnel, nil
}
