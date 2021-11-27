package status

import (
	"image/color"
	"log"
	"os/exec"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
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
	icon *widget.Icon
}

func (n *network) Destroy() {
}

func (n *network) wirelessName() (string, error) {
	net := ""
	if iw, _ := exec.LookPath("iw"); iw != "" {
		out, err := exec.Command("bash", []string{"-c", "iw dev | grep Interface | cut -d \" \" -f2"}...).Output()
		if err != nil {
			log.Println("Error running iw", err)
			return "", err
		}
		dev := strings.TrimSpace(string(out))
		if dev == "" {
			return "", nil
		}

		out, err = exec.Command("bash", []string{"-c", "iw dev " + dev + " info | grep ssid | sed 's\\ssid\\\\g'"}...).Output()
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
		out, err := exec.Command("bash", []string{"-c", "ip link | grep \",UP,\" | grep -v LOOPBACK | grep -v \": wl\""}...).Output()
		if err != nil {
			log.Println("Error running ip tool", err)
			return false, err
		}
		if strings.TrimSpace(string(out)) == "" {
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
					n.icon.SetResource(wmtheme.WifiOffIcon)
				} else if val == networkNameEthernet {
					n.icon.SetResource(wmtheme.EthernetIcon)
				} else {
					n.icon.SetResource(wmtheme.WifiIcon)
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
	n.icon = widget.NewIcon(wmtheme.WifiOffIcon)
	prop := canvas.NewRectangle(color.Transparent)
	prop.SetMinSize(n.icon.MinSize().Add(fyne.NewSize(theme.Padding()*4, 0)))
	icon := container.NewCenter(prop, n.icon)
	n.tick()

	return container.NewBorder(nil, nil, icon, nil, n.name)
}

func (n *network) Metadata() fynedesk.ModuleMetadata {
	return networkMeta
}

// NewNetwork creates a new module that will show network information in the status area
func NewNetwork() fynedesk.Module {
	return &network{}
}
