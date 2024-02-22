package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"strconv"

	"github.com/containernetworking/cni/pkg/skel"
	"github.com/containernetworking/cni/pkg/types"
	current "github.com/containernetworking/cni/pkg/types/100"
	"github.com/containernetworking/cni/pkg/version"
	"github.com/vishvananda/netlink"
)

// 独自の設定構造体を定義
type MyPluginConf struct {
	CNIVersion string `json:"cniVersion"`
	Name       string `json:"name"`
	Type       string `json:"type"`
	PodCIDR    string `json:"podcidr"` // 特定のプラグイン設定
}

func incrementIP(ip net.IP, increment int) net.IP {
	// IPアドレスを4バイトの整数に変換してインクリメントし、IPアドレスに戻す
	ipInt := net.IPv4(ip[0], ip[1], ip[2], ip[3]).To4()
	ipInt[3] += byte(increment)
	return ipInt
}

func createBridge(bridgeName string, gatewayIP net.IP, cidr *net.IPNet) error {

	// ブリッジオブジェクトの作成、ここでLinkAttrsに名前を設定
	br := &netlink.Bridge{LinkAttrs: netlink.LinkAttrs{Name: bridgeName}}
	if err := netlink.LinkAdd(br); err != nil {
		return fmt.Errorf("failed to add bridge %s: %v", bridgeName, err)
	}

	// ブリッジをアップする
	if err := netlink.LinkSetUp(br); err != nil {
		return fmt.Errorf("failed to set %s up: %v", bridgeName, err)
	}

	addr := &netlink.Addr{IPNet: &net.IPNet{IP: gatewayIP, Mask: cidr.Mask}}
	if err := netlink.AddrAdd(br, addr); err != nil {
		return fmt.Errorf("failed to add addr to bridge: %v", err)
	}

	return nil
}

// cmdAdd is called for ADD requests
func cmdAdd(args *skel.CmdArgs) error {
	conf := MyPluginConf{}
	if err := json.Unmarshal(args.StdinData, &conf); err != nil {
		return fmt.Errorf("failed to parse network configuration: %v", err)
	}

	_, cidr, err := net.ParseCIDR(conf.PodCIDR)
	if err != nil {
		return fmt.Errorf("failed to parse PodCIDR: %v", err)
	}
	gatewayIP := make(net.IP, len(cidr.IP))
	copy(gatewayIP, cidr.IP)
	gatewayIP[3] = 1

	// createBridge関数を修正した引数で呼び出し
	err = createBridge("cni0", gatewayIP, cidr)
	if err != nil {
		return fmt.Errorf("failed to create bridge: %v", err)
	}

	lastIPFile := "/tmp/last_allocated_ip"
	var n int

	// 最後に割り当てられたIPを読み込む
	content, err := ioutil.ReadFile(lastIPFile)
	if err == nil {
		// ファイルが存在する場合、内容を解析
		lastIP, err := strconv.Atoi(string(content))
		if err != nil {
			return fmt.Errorf("failed to parse last allocated IP: %v", err)
		}
		n = lastIP
	} else {
		// ファイルが存在しない場合は、nを1に設定
		n = 1
	}

	n++ // 次のIPアドレスを計算
	//baseIP := cidr.IP.To4()
	//newIP := incrementIP(baseIP, n) // 新しいIPアドレスを計算

	// 新しいIPアドレスをファイルに書き込む
	err = ioutil.WriteFile(lastIPFile, []byte(strconv.Itoa(n)), 0644)
	if err != nil {
		return fmt.Errorf("failed to write last allocated IP: %v", err)
	}

	// インターフェイス、IP、ルート情報を設定
	interfaces := []*current.Interface{{
		Name:    "eth0",
		Mac:     "02:42:ac:11:00:02",
		Sandbox: "/var/run/netns/cni-12345",
	}}

	ips := []*current.IPConfig{{
		Interface: current.Int(0), // interfacesスライスのインデックスへのポインタ
		Address: net.IPNet{
			IP:   net.ParseIP("10.0.0.2"),
			Mask: net.CIDRMask(24, 32), // 24ビットのサブネットマスク
		},
		Gateway: net.ParseIP("10.0.0.1"),
	}}

	// Resultオブジェクトの作成
	result := &current.Result{
		CNIVersion: "0.3.1",
		Interfaces: interfaces,
		IPs:        ips,
	}

	// 結果を出力
	return types.PrintResult(result, conf.CNIVersion)
}

// cmdDel is called for DELETE requests
func cmdDel(args *skel.CmdArgs) error {
	return nil
}

func main() {
	skel.PluginMain(cmdAdd, cmdGet, cmdDel, version.All, "")
}

func cmdGet(args *skel.CmdArgs) error {
	return types.PrintResult(&current.Result{}, "0.4.0")
}
