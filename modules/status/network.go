package status

import (
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"fyne.io/fynedesk"
	wmtheme "fyne.io/fynedesk/theme"
)

var networkMeta = fynedesk.ModuleMetadata{
	Name:        "Network",
	NewInstance: NewNetwork,
}

const networkNameEthernet = "Ethernet"

type network struct {
	name *widget.Label
	icon *widget.Button
}

func (n *network) Destroy() {
}

func (n *network) wirelessName() (string, error) {
	net := ""
	iw, _ := exec.LookPath("iw")
	if iw == "" {
		iw, _ = exec.LookPath("/usr/sbin/iw")
	}
	if iw != "" {
		out, err := exec.Command("bash", []string{"-c", iw + " dev | grep Interface | cut -d \" \" -f2"}...).Output()
		if err != nil {
			log.Println("Error running iw", err)
			return "", err
		}
		dev := strings.TrimSpace(string(out))
		if dev == "" {
			return "", nil
		}

		out, err = exec.Command("bash", []string{"-c", iw + " dev " + dev + " info | grep ssid | sed 's\\ssid\\\\g'"}...).Output()
		if err != nil {
			log.Println("Error running iw", err)
			return "", err
		}
		net = string(out)
	} else {
		out, err := exec.Command("bash", []string{"-c", "/System/Library/PrivateFrameworks/Apple80211.framework/Resources/airport -I  | awk -F' SSID: '  '/ SSID: / {print $2}'"}...).Output()
		if err != nil {
			log.Println("Error getting network info from airport utility", err)
			return "", err
		}

		net = string(out)
	}
	return strings.TrimSpace(net), nil
}

func (n *network) isEthernetConnected() (bool, error) {
	if ip, _ := exec.LookPath("ip"); ip != "" {
		out, err := exec.Command("bash", []string{"-c", "ip link | grep \",UP,\" | grep -v LOOPBACK | grep -v \": wl\" | wc -l"}...).Output()
		if err != nil {
			log.Println("Error running ip tool", err)
			return false, err
		}
		if strings.TrimSpace(string(out)) == "0" {
			return false, nil
		}
	} else {
		out, err := exec.Command("bash", []string{"-c", "ifconfig | pcregrep -M -o '^[^\\t:]+:([^\\n]|\\n\\t)*status: active'"}...).Output()
		if err != nil {
			log.Println("Error running ifconfig tool", err)
			return false, err
		}
		if !strings.Contains(string(out), "broadcast") {
			return false, nil
		}
	}
	return true, nil
}

func (n *network) networkName() string {
	name, _ := n.wirelessName()
	if name != "" {
		return name
	}

	ether, _ := n.isEthernetConnected()
	if ether {
		return networkNameEthernet
	}
	return ""
}

func (n *network) tick() {
	tick := time.NewTicker(time.Second * 10)
	go func() {
		for {
			val := n.networkName()
			if val != n.name.Text {
				n.name.SetText(val)

				if val == "" {
					n.icon.SetIcon(wmtheme.WifiOffIcon)
				} else if val == networkNameEthernet {
					n.icon.SetIcon(wmtheme.EthernetIcon)
				} else {
					n.icon.SetIcon(wmtheme.WifiIcon)
				}
			}
			<-tick.C
		}
	}()
}

func (n *network) StatusAreaWidget() fyne.CanvasObject {
	if _, err := n.wirelessName(); err != nil {
		if _, err = n.isEthernetConnected(); err != nil {
			return nil
		}
	}

	n.name = widget.NewLabel("")
	n.icon = &widget.Button{Icon: wmtheme.WifiOffIcon, Importance: widget.LowImportance, OnTapped: n.showSettings}
	n.tick()

	return container.NewBorder(nil, nil, n.icon, nil, n.name)
}

func (n *network) Metadata() fynedesk.ModuleMetadata {
	return networkMeta
}

func (n *network) showSettings() {
	gui := &networkApp{}

	if err := fynedesk.Instance().RunApp(gui); err != nil {
		fyne.LogError("Failed to find WiFi settings tool connman-gtk", err)
		return
	}
}

// NewNetwork creates a new module that will show network information in the status area
func NewNetwork() fynedesk.Module {
	return &network{}
}

type networkApp struct {
}

func (n *networkApp) Name() string {
	return "Network Settings"
}

func (n *networkApp) Run(env []string) error {
	vars := os.Environ()
	vars = append(vars, env...)

	cmd := exec.Command("connman-gtk")
	cmd.Env = vars
	return cmd.Start()
}

func (n *networkApp) Categories() []string {
	return []string{"Settings"}
}

func (n *networkApp) Hidden() bool {
	return true
}

func (n *networkApp) Icon(theme string, size int) fyne.Resource {
	return wmtheme.WifiIcon
}
