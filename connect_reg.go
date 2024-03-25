package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"strconv"
)

type ConnectInfo struct {
    Mode           int     // Cのint。Goのint型はプラットフォームに依存するので、互換性に注意（必要に応じてint32またはint64を使用）
    SrcIP          uint32  // Cの__u32に対応
    DstIP          uint32  // 同上
    SrcPort        uint16  // Cの__u16に対応
    DstPort        uint16  // 同上
    DstCore        uint16  // 同上
    ConnectForward uint8    // Cのboolに対応（C99以降では_Bool型、1バイト）
}

func Connect_reg(src_ip string, src_port_s string, dst_ip string, dst_port_s string, filter_dest_core_s string, connect_forward bool) error {

	src_port, err := strconv.Atoi(src_port_s)
	if err != nil {
		return fmt.Errorf("failed to cast listen_port: %v",err)
	}


	dst_port, err := strconv.Atoi(dst_port_s)
	if err != nil {
		return fmt.Errorf("failed to cast dst_port: %v", err)
	}


	var c_info ConnectInfo
	dst_core, err :=strconv.Atoi(filter_dest_core_s)
	if err != nil {
		return fmt.Errorf("failed to cast dst_core: %v", err)
	}

	c_info.DstCore = uint16(dst_core)

	conn, err := net.Dial("tcp", net.JoinHostPort("127.0.0.1","25909"))
	if err != nil {
		return fmt.Errorf("failed to create socket: %v\n",err)
	}
	defer conn.Close()

	if  connect_forward {
		c_info.ConnectForward = 1
	}else {
		c_info.ConnectForward = 0
	}

	var src_ipUint32 uint32
	ipBytes := net.ParseIP(src_ip).To4()
	for i := 0; i < 4; i++ {
		src_ipUint32 = src_ipUint32<<8 + uint32(ipBytes[i])
	}
	c_info.SrcIP = src_ipUint32

	var dst_ipUint32 uint32
	ipBytes = net.ParseIP(dst_ip).To4()
	for i := 0; i < 4; i++ {
		dst_ipUint32 = dst_ipUint32<<8 + uint32(ipBytes[i])
	}
	c_info.DstIP = dst_ipUint32

	c_info.SrcPort = uint16(src_port)

	c_info.DstPort = uint16(dst_port)

	c_info.Mode = 3

	buf := &bytes.Buffer{}
	binary.Write(buf, binary.LittleEndian, c_info.Mode)
    binary.Write(buf, binary.LittleEndian, c_info.SrcIP)
    binary.Write(buf, binary.LittleEndian, c_info.DstIP)
    binary.Write(buf, binary.LittleEndian, c_info.SrcPort)
    binary.Write(buf, binary.LittleEndian, c_info.DstPort)
    binary.Write(buf, binary.LittleEndian, c_info.DstCore)
    binary.Write(buf, binary.LittleEndian, c_info.ConnectForward)

    //err = binary.Write(buf, binary.LittleEndian, &c_info)
    //if err != nil {
    //    return fmt.Errorf("failed to serialize connect_info: %v", err)
    //}

	_, err = conn.Write(buf.Bytes())
    if err != nil {
        return fmt.Errorf("failed to send connect_information: %v", err)
    }

	return nil
	
}
