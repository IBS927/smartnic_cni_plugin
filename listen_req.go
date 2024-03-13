package main

import (
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"net"
	"strconv"
)

func Listen_req(src_ip string, src_port_s string, dst_ip string, dst_port_s string, filter_dest_core_s string, wasm_pass string) error {
	

	src_port, err := strconv.Atoi(src_port_s)
	if err != nil {
		return fmt.Errorf("failed to cast listen_port: %v",err)
	}

	dst_port, err := strconv.Atoi(dst_port_s)
	if err != nil {
		return fmt.Errorf("failed to cast dst_port: %v", err)
	}

	//dst_ip := 0x7f000001

	//const XWAMR_WASM_BIN_BUF_SIZE = 20000
	const cmd = uint16(2)

	filter_dest_core, err := strconv.Atoi(filter_dest_core_s)
	
	conn, err := net.Dial("tcp", net.JoinHostPort("127.0.0.1","25908"))
	if err != nil {
		return fmt.Errorf("failed to create socket: %v\n",err)
	}
	defer conn.Close()

	if err := binary.Write(conn, binary.LittleEndian, cmd); err != nil {
		return fmt.Errorf("Failed to send cmd: %v\n", err)
	}

	if err := binary.Write(conn, binary.LittleEndian, uint16(dst_port)); err != nil {
		return fmt.Errorf("failed to send SNIC port:%v\n", err)
	}

	if err := binary.Write(conn, binary.LittleEndian, uint16(src_port)); err != nil {
		return fmt.Errorf("failed to send container port:%v\n", err)
	}

	var ipUint32 uint32
	ipBytes := net.ParseIP(src_ip).To4()
	for i := 0; i < 4; i++ {
		ipUint32 = ipUint32<<8 + uint32(ipBytes[i])
	}
	if err := binary.Write(conn, binary.LittleEndian, ipUint32); err != nil {
		return fmt.Errorf("failed to send IP: %v\n",err)
	}

	if err:= binary.Write(conn, binary.LittleEndian, uint16(filter_dest_core)); err != nil {
		return fmt.Errorf("failed to send filter dest core : %v\n",err)
	}


	return nil


}
