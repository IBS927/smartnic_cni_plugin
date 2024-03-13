package cni

import (
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strconv"
	"strings"
)

func main(){
	if len(os.Args) != 7 {
		return fmt.Errorf("Usage: ./listen_req src_ip src_port dst_ip dst_port arm_core wasm_pass")
	}

	src_ip := os.Args[1]

	src_port, err := strconv.Atoi(os.Args[2])
	if err != nil {
		return fmt.Errorf("failed to cast listen_port: %v",err)
	}

	dst_port, err := strconv.Atoi(os.Args[4])
	if err != nil {
		return fmt.Errorf("failed to cast dst_port: %v", err)
	}

	//dst_ip := 0x7f000001

	const XWAMR_WASM_BIN_BUF_SIZE = 20000
	const cmd = 3

	filter_dest_core, err := strconv.Atoi(os.Args[5])
	wasm_pass := os.Args[6]
	
	wasmBin, err := ioutil.ReadFile(wasm_pass)
	if err != nil {
		return fmt.Errorf("failed to read WASM binary file: %v\n", err)
	}

	fileSize := len(wasmBin)
	if fileSize > XWAMR_WASM_BIN_BUF_SIZE {
		return fmt.Errorf("WASM binary size exceeds the maximum buffer size.")
	}

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


	return 0


}