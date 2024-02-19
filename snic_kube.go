package main

import (
	"encoding/json"
	"fmt"

	"github.com/containernetworking/cni/pkg/skel"
	"github.com/containernetworking/cni/pkg/types"
	current "github.com/containernetworking/cni/pkg/types/100"
	"github.com/containernetworking/cni/pkg/version"
)

// 独自の設定構造体を定義
type MyPluginConf struct {
	CNIVersion string `json:"cniVersion"`
	Name       string `json:"name"`
	Type       string `json:"type"`
	PodCIDR    string `json:"podcidr"` // 特定のプラグイン設定
}

// cmdAdd is called for ADD requests
func cmdAdd(args *skel.CmdArgs) error {
	conf := MyPluginConf{}
	if err := json.Unmarshal(args.StdinData, &conf); err != nil {
		return fmt.Errorf("failed to parse network configuration: %v", err)
	}

	// conf内の設定を使用してネットワーク設定を行う
	fmt.Printf("Received CNI Config: %+v\n", conf)

	return types.PrintResult(&current.Result{}, conf.CNIVersion)
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
