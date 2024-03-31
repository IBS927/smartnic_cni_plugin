package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os/exec"
	"strconv"
	"strings"
	"os"
	"github.com/containernetworking/cni/pkg/skel"
	"github.com/containernetworking/cni/pkg/types"
	current "github.com/containernetworking/cni/pkg/types/100"
	"github.com/containernetworking/cni/pkg/version"
)

type PluginEnvArgs struct {
	types.CommonArgs
	CONTAINER_NAME string
	CNI_NETNS      string
}
type PluginConf struct {
	CNIVersion string `json:"cniVersion"`
	Name       string `json:"name"`
	Type       string `json:"type"`
}
type IPPair struct {
	SNICIP      string
	containerIP string
}

func parseConfig(stdin []byte) (*PluginConf, error) {
	conf := &PluginConf{}
	if err := json.Unmarshal(stdin, conf); err != nil {
		return nil, fmt.Errorf("failed to parse network configuration: %v", err)
	}

	return conf, nil
}
func makeCNIArgs(args *skel.CmdArgs) (*PluginEnvArgs, error) {
	env := &PluginEnvArgs{}
	if err := types.LoadArgs(args.Args, env); err != nil {
		return nil, types.NewError(types.ErrInvalidEnvironmentVariables, "failed to load CNI_ARGS", err.Error())
	}
	return env, nil
}
func cmdAdd(args *skel.CmdArgs) error {
	conf, err := parseConfig(args.StdinData)
	if err != nil {
		return err
	}

	//cniArgs, err := makeCNIArgs(args)
	//if err != nil {
	//	return err
	//}
	c_name:=os.Getenv("CONTAINER_NAME")
	c_netns:=os.Getenv("CNI_NETNS")
	//fmt.Printf("ko %v\n",c_name)
	// 事前に決めたIPの最後をnとする
	var n int
	switch c_name {
	case "productpage":
		n = 11
	case "details":
		n = 12
	case "reviews-v1":
		n = 13
	case "reviews-v2":
		n = 14
	case "reviews-v3":
		n = 15
	case "ratings":
		n = 16
	}
	// eth(n-1)とvethnのペアを作る。
	ethName := fmt.Sprintf("eth%d", n-1)
	vethName := fmt.Sprintf("veth%d", n)
	ipAddress := fmt.Sprintf("192.168.0.%d", n)
	ipAdd_cidr:=ipAddress+"/24"
	software_bridge := "my_bridge"
	//bridge_ip := fmt.Sprintf("192.168.11.%d", 100+n)
	listen_port := fmt.Sprintf("%d", 10000+n)
	if err := exec.Command("ip", "link", "add", ethName, "type", "veth", "peer", "name", vethName).Run(); err != nil {
		return fmt.Errorf("failed to create veth pair: %v", err)
	}

	// vethペアの片方（vethn）をコンテナのネットワーク名前空間に移動
	if err := exec.Command("ip", "link", "set", vethName, "netns", c_netns).Run(); err != nil {
		return fmt.Errorf("failed to move veth to container netns: %v", err)
	}

	// vethにIPをつける
	if err := exec.Command("nsenter", "--net="+c_netns, "ip", "addr", "add", ipAdd_cidr, "dev", vethName).Run(); err != nil {
		return fmt.Errorf("failed to assign ip to veth: %v", err)
	}

	// コンテナ内でvethnのインターフェースをアップ
	if err := exec.Command("nsenter", "--net="+c_netns, "ip", "link", "set", vethName, "up").Run(); err != nil {
		return fmt.Errorf("failed to set eth interface up in container: %v", err)
	}

	// ethとsoftware_bridgeを繋げる
	if err := exec.Command("ip", "link", "set", ethName, "master", software_bridge).Run(); err != nil {
		return fmt.Errorf("failed to attach eth to bridge: %v", err)
	}
	if err := exec.Command("ip","link","set",ethName,"up").Run(); err != nil {
		return fmt.Errorf("failed to eth up: %v",err)
	}

	// software_bridgeに新たなIPをくっつける
	//if err := exec.Command("ip", "addr", "add", bridge_ip+"/24", "dev", software_bridge).Run(); err != nil {
	//	return fmt.Errorf("failed to assigne ip to bridge: %v", err)
	//}
	// コンテナのipとそのノードのipの対応表
	ipPairs := []IPPair{
		{SNICIP: "192.168.0.202", containerIP: "192.168.0.11"},
		{SNICIP: "192.168.0.202", containerIP: "192.168.0.12"},
		{SNICIP: "192.168.0.201", containerIP: "192.168.0.13"},
		{SNICIP: "192.168.0.201", containerIP: "192.168.0.14"},
		{SNICIP: "192.168.0.201", containerIP: "192.168.0.15"},
		{SNICIP: "192.168.0.201", containerIP: "192.168.0.16"},
	}
	//listenとconnectを行う
	//SNICIP==..201なら100,202なら101にiptablesでルーティングを行う。
	var gw_str string
	var snic_now_ip string
	for _, pair_now := range ipPairs {
		if pair_now.containerIP == ipAddress {
			snic_now_ip = pair_now.SNICIP
		}
	}

	if snic_now_ip == "192.168.0.202" {
		if err := exec.Command("nsenter", "--net="+c_netns, "ip", "route", "add", "default", "via", "192.168.0.100").Run(); err != nil {
			return fmt.Errorf("can't attch route %v", err)
		}
	} else {
		if err := exec.Command("nsenter", "--net="+c_netns, "ip", "route", "add", "default", "via", "192.168.0.102").Run(); err != nil {
			return fmt.Errorf("can't attch route %v", err)
		}
	}

	for _, pair := range ipPairs {
		if pair.containerIP == ipAddress {
			if pair.SNICIP == "192.168.0.201" {
				gw_str = "192.168.0.41"
			} else {
				gw_str = "192.168.0.150"
			}
			err:=Listen_req(ipAddress,"9080",pair.SNICIP, listen_port, "0","/home/appleuser/nic-toe_buff3/sdk_work_zynq/wamer_work/src/sample/pass.wasm")
			if err != nil {
				return fmt.Errorf("failed to listen_req : %v\n",err)
			}
			var forward_ip string
			if snic_now_ip == "192.168.0.202" {
				forward_ip = "192.168.0.100"
			} else {
				forward_ip = "192.168.0.102"
			}
			err = Connect_reg(forward_ip,listen_port,snic_now_ip,listen_port,"0", true)
			if err != nil {
				return fmt.Errorf("failed to connect: %v", err)
			}
			//cmd_listen := exec.Command("./listen_req", ipAddress, "9080", pair.SNICIP, listen_port, "0","../sdk_work_zynq/wamer_work/src/sample/pass.wasm")
			//cmd_listen.Dir = "/home/appleuser/nic-toe_buff3/ebpf"
			//if err := cmd_listen.Run(); err != nil {
			//	return fmt.Errorf("failed to listen_req: %v", err)
			//}
		} else {
			var forward_ip string
			if snic_now_ip == "192.168.0.202" {
				forward_ip = "192.168.0.100"
			} else {
				forward_ip = "192.168.0.102"
			}
			split_ip := strings.Split(pair.containerIP, ".")
			if len(split_ip) < 4 {
				fmt.Errorf("not IPAddress")
			}
			last_ip, err := strconv.Atoi(split_ip[3])
			if err != nil {
				fmt.Errorf("The last ip is not int type: %v", err)
			}
			forward_port := fmt.Sprintf("%d", last_ip+10000)
			if snic_now_ip == pair.SNICIP {
				if err := exec.Command("iptables", "-t", "nat", "-I", "PREROUTING", "-p", "tcp", "-s", ipAddress, "-d", pair.containerIP, "-j", "DNAT", "--to-destination", forward_ip+":15006").Run(); err != nil {
					return fmt.Errorf("failed to set iptables: %v", err)
				}
			}else{
				if err := exec.Command("iptables", "-t", "nat", "-I", "PREROUTING", "-p", "tcp", "-s", ipAddress, "-d", pair.containerIP, "-j", "DNAT", "--to-destination", forward_ip+":"+forward_port).Run(); err != nil {
					return fmt.Errorf("failed to set iptables: %v", err)
				}
			}
			//if err := exec.Command("iptables", "-t", "nat", "-I", "PREROUTING", "-p", "tcp", "-s", ipAddress, "-d", pair.containerIP, "-j", "DNAT", "--to-destination", forward_ip+":"+forward_port).Run(); err != nil {
			//	return fmt.Errorf("failed to set iptables: %v", err)
			//}
			err = Connect_reg(forward_ip,forward_port,pair.SNICIP,forward_port,"0", true)
			if err != nil {
				return fmt.Errorf("failed to connect: %v", err)
			}
		}
	}
	
	iface, err := net.InterfaceByName(ethName)
	if err != nil {
		return fmt.Errorf("No veth interface mac: %v", err)
	}

	// MACアドレスを取得
	mac := iface.HardwareAddr.String()

	ip, ipNet, err := net.ParseCIDR(ipAddress + "/24")
	if err != nil {
		return fmt.Errorf("failed to parse IP CIDR: %v", err)
	}
	gw := net.ParseIP(gw_str)
	// CNI Resultの構築
	result := &current.Result{
		CNIVersion: "0.4.0",
		Interfaces: []*current.Interface{{
			Name:    ethName,
			Mac:     mac,
			Sandbox: c_netns,
			// Sandboxはコンテナのネットワーク名前空間へのパスですが、ここでは例として空にしています
		}},
		IPs: []*current.IPConfig{{
			Address: net.IPNet{IP: ip, Mask: ipNet.Mask},
			Gateway: gw,
			// InterfaceはInterfacesスライス内の対象インターフェースのインデックスです。
			// ここでは最初の（そして唯一の）インターフェースを指しています。
			Interface: current.Int(0),
		}},
	}

	// 出力をJSON形式で標準出力に出力
	return types.PrintResult(result, conf.CNIVersion)
}
func cmdDel(args *skel.CmdArgs) error {
	return nil
}

func main(){
	co:= os.Getenv("CNI_COMMAND")
	fmt.Printf("%v\n",co)
	skel.PluginMain(cmdAdd, cmdCheck, cmdDel, version.All, "")
}

func cmdCheck(args *skel.CmdArgs) error {
	return types.PrintResult(&current.Result{}, "0.4.0")
}
