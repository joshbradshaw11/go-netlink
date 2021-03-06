// This file is part of go-netlink.
//
// Copyright (C) 2015 Max Hille <mh@lambdasoup.com>
//
// go-netlink is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// at your option) any later version.
//
// go-netlink is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with go-netlink.  If not, see <http://www.gnu.org/licenses/>.

// Package netlink provides access to the Linux kernel's Netlink interface
package netlink

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"syscall"

	"github.com/lambdasoup/go-netlink/log"
)

type netlinkMsg struct {
	len     uint32
	msgType uint16
	flags   uint16
	seq     uint32
	pid     uint32
	data    []byte
}

// Socket is a Linux Netlink socket
type Socket struct {
	socketFd int
	lsa      *syscall.SockaddrNetlink
	seq      uint32
}

// Open creates and binds a new Netlink socket
func Open() (*Socket, error) {
	// TODO remove Connector hardcode
	socketFd, err := syscall.Socket(syscall.AF_NETLINK, syscall.SOCK_DGRAM, syscall.NETLINK_CONNECTOR)
	if err != nil {
		return nil, err
	}
	lsa := &syscall.SockaddrNetlink{}
	lsa.Groups = 0
	lsa.Family = syscall.AF_NETLINK
	lsa.Pid = 0
	err = syscall.Bind(socketFd, lsa)
	return &Socket{socketFd, lsa, 0xaffe}, err
}

// Close this Socket's connection
func (s *Socket) Close() {
	syscall.Close(s.socketFd)
}

// Send the given data through this Netlink connection
func (s *Socket) Send(data []byte) error {
	// TODO remove magic numbers
	msg := &netlinkMsg{uint32(syscall.NLMSG_HDRLEN + len(data)), syscall.NLMSG_DONE, 0, s.seq, uint32(os.Getpid()), data}
	s.seq++

	log.Printf("\t\t\tNL SEND: %v", msg)

	// TODO remove magic number
	err := syscall.Sendto(s.socketFd, msg.Bytes(), 0, s.lsa)
	return err
}

func (msg *netlinkMsg) Bytes() []byte {
	buf := new(bytes.Buffer)

	binary.Write(buf, binary.LittleEndian, msg.len)
	binary.Write(buf, binary.LittleEndian, msg.msgType)
	binary.Write(buf, binary.LittleEndian, msg.flags)
	binary.Write(buf, binary.LittleEndian, msg.seq)
	binary.Write(buf, binary.LittleEndian, msg.pid)

	buf.Write(msg.data)

	return buf.Bytes()
}

func (msg *netlinkMsg) String() string {

	// from linux/netlink.h
	msgTypes := map[uint16]string{
		syscall.NLMSG_NOOP:    "NLMSG_NOOP",
		syscall.NLMSG_ERROR:   "NLMSG_ERROR",
		syscall.NLMSG_DONE:    "NLMSG_DONE",
		syscall.NLMSG_OVERRUN: "NLMSG_OVERRUN",
	}

	return fmt.Sprintf("NetlinkMsg{len: %d, %v, %x, seq: %d, port: %d, body: %d}",
		msg.len, msgTypes[msg.msgType], msg.flags, msg.seq, msg.pid, len(msg.data))
}

// Receive data from this Netlink connection
func (s *Socket) Receive() ([]byte, error) {
	// TODO remove magic numbers
	rb := make([]byte, 8192)
	_, _, err := syscall.Recvfrom(s.socketFd, rb, 0)
	if err != nil {
		return nil, err
	}

	msg, err := parseNetlinkMsg(rb)
	log.Printf("\t\t\tNL RECV: %v", msg)
	return msg.data, err
}

func parseNetlinkMsg(bs []byte) (*netlinkMsg, error) {
	msg := &netlinkMsg{}
	buf := bytes.NewBuffer(bs)

	err := error(nil)
	err = binary.Read(buf, binary.LittleEndian, &msg.len)
	err = binary.Read(buf, binary.LittleEndian, &msg.msgType)
	err = binary.Read(buf, binary.LittleEndian, &msg.flags)
	err = binary.Read(buf, binary.LittleEndian, &msg.seq)
	err = binary.Read(buf, binary.LittleEndian, &msg.pid)

	msg.data = make([]byte, msg.len-syscall.NLMSG_HDRLEN)

	_, err = buf.Read(msg.data)

	// check for truncated data
	for {
		bs := make([]byte, 1)
		_, eof := buf.Read(bs)
		if eof != nil {
			break
		}
		if bs[0] == 0 {
			continue
		}

		err = errors.New("NL parse left truncated data")
	}

	return msg, err
}
