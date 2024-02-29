package main

import (
	"encoding/json"
	"fmt"
	"os/exec"

	"github.com/containernetworking/cni/pkg/skel"
	"github.com/containernetworking/cni/pkg/types"
	current "github.com/containernetworking/cni/pkg/types/100"
	"github.com/containernetworking/cni/pkg/version"
)

type PluginEnvArgs struct {
	types.CommonArgs
	CONTAINER_NAME string `json:"container_name"`
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

	cniArgs, err := makeCNIArgs(args)
	if err != nil {
		return err
	}

	var n int
	switch cniArgs.CONTAINER_NAME {
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

	ethName := fmt.Sprintf("eth%d", n-1)
	vethName := fmt.Sprintf("veth%d", n)
	ipAddress := fmt.Sprintf("192.168.11.%d", n)

	if err := exec.Command("ip", "link", "add", ethName, "type", "veth", "peer", "name", vethName).Run(); err != nil {
		return fmt.Errorf("failed to create veth pair: %v", err)
	}

	// vethペアの片方（vethn）をコンテナのネットワーク名前空間に移動
	if err := exec.Command("ip", "link", "set", vethName, "netns", args.Netns).Run(); err != nil {
		return fmt.Errorf("failed to move veth to container netns: %v", err)
	}

	// コンテナ内でeth(n-1)のインターフェースをアップ
	if err := exec.Command("nsenter", "--net="+args.Netns, "ip", "link", "set", ethName, "up").Run(); err != nil {
		return fmt.Errorf("failed to set eth interface up in container: %v", err)
	}

	cmd := fmt.Sprintf("ip addr add %s/24 dev %s && ip link set %s up", ipAddress, ethName, ethName)

	netnsCmd := exec.Command("nsenter", "--net="+args.Netns, "bash", "-c", cmd)

	if err := netnsCmd.Run(); err != nil {
		return fmt.Errorf("failed to assign IP address and bring up interface in container: %v", err)
	}

	//listenのコマンドを打つ

	// コマンドと引数を準備
	cmd_listen := exec.Command("./listen_req", "80", "0")
	// コマンドを実行するディレクトリを指定
	cmd_listen.Dir = "/home/appleuser/nic-toe_buff3/ebpf"

	if err := cmd_listen.Run(); err != nil {
		return fmt.Errorf("failed to run listen command: %v", err)
	}

	//connectのコマンドを打つ
	ipPairs := []IPPair{
		{SNICIP: "192.168.11.202", containerIP: "192.168.11.11"},
		{SNICIP: "192.168.11.202", containerIP: "192.168.11.12"},
		{SNICIP: "192.168.11.201", containerIP: "192.168.11.13"},
		{SNICIP: "192.168.11.201", containerIP: "192.168.11.14"},
		{SNICIP: "192.168.11.201", containerIP: "192.168.11.15"},
		{SNICIP: "192.168.11.201", containerIP: "192.168.11.16"},
	}

	for _, pair := range ipPairs {
		if pair.containerIP != ipAddress {
			cmd_connect := exec.Command("./connect_reg", ipAddress+":9080"+":"+pair.SNICIP+":???", "0")
			cmd_connect.Dir = "/home/appleuser/nic-toe_buff3/ebpf"

			if err := cmd_connect.Run(); err != nil {
				return fmt.Errorf("failed to run connect command: %v", err)
			}
		}
	}
	return conf
}
func cmdDel(args *skel.CmdArgs) error {
	return nil
}

func main() {
	skel.PluginMain(cmdAdd, cmdGet, cmdDel, version.All, "")
}

func cmdGet(args *skel.CmdArgs) error {
	return types.PrintResult(&current.Result{}, "0.4.0")
}
