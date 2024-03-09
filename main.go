package main

import (
	_ "embed"
	"io"
	"log"
	"os/exec"
	"strings"
	"time"

	"github.com/getlantern/systray"
)

type icon struct {
	Base64  string
	Decoded []byte
}

type menuItem struct {
	mac       string
	name      string
	menuItem  *systray.MenuItem
	connected bool
}

type Status struct {
	powered bool
	discoverable bool
	pairable bool
	discovering bool
	devices map[string]string
	connected map[string]string
}

type App struct {
	quit *systray.MenuItem
	power *systray.MenuItem
	discovering *systray.MenuItem
	devices map[string]menuItem
	status Status
}

type Bluetoothctl struct { }

func on_off(on bool) string {
	if on {
		return "on"
	} else {
		return "off"
	}
}

func connect_disconnect(connect bool) string {
	if connect {
		return "connect"
	} else {
		return "disconnect"
	}
}

func (m *Bluetoothctl) power(on bool) error {
	return exec.Command("bluetoothctl", "power", on_off(on)).Run()
}

func (m *Bluetoothctl) scan(on bool) error {
	return exec.Command("bluetoothctl", "scan", on_off(on)).Run()
}

func (m *Bluetoothctl) discovering(on bool) error {
	return exec.Command("bluetoothctl", "discovering", on_off(on)).Run()
}

func (m *Bluetoothctl) devices(params ...string) (result map[string]string, err error) {
	params = append([]string{"devices"}, params...)
	log.Println(params)
	output, err := exec.Command("bluetoothctl", params...).Output()
	if err != nil {
		return
	}
	result = getDevices(string(output))
	return
}

func (m *Bluetoothctl) status() (result Status, err error) {
	output, err := exec.Command("bluetoothctl", "show").Output()
	for _, line := range strings.Split(strings.TrimSpace(string(output)), "\n") {
		parts := strings.Split(strings.TrimSpace(line), " ")
		switch parts[0] {
		case "Powered:":
			result.powered = parts[1] == "yes"
		case "Discoverable":
			result.discoverable = parts[1] == "yes"
		case "Pairable":
			result.pairable = parts[1] == "yes"
		case "Discovering":
			result.discovering = parts[1] == "yes"
		}
		if parts[0] == "Powered:" {
			result.powered = parts[1] == "yes"
		}
	}
	return
}

func (m *Bluetoothctl) connect(mac string, connect bool) error {
	return exec.Command("bluetoothctl", connect_disconnect(connect), mac).Run()
}

var (
	//go:embed blue.png
	on []byte

	//go:embed red.png
	off []byte

	//go:embed green.png
	connected []byte

	//go:embed yellow.png
	standby []byte

	b Bluetoothctl
)

func main() {

	// syslog, err := syslog.New(syslog.LOG_INFO, "bluetooth-menu")
	// if err != nil {
	// 	panic("Unable to connect to syslog")
	// }
	log.SetOutput(io.Discard)

	systray.Run(func() {
		systray.SetIcon(on)
		status, err := getStatus()
		if err != nil {
			log.Println(err)
			return
		}
		app := App {
			systray.AddMenuItem("Quit", ""),
			systray.AddMenuItem("power", ""),
			systray.AddMenuItem("discover", ""),
			make(map[string]menuItem),
			status,
		}
		systray.AddSeparator()
		tick := time.Tick(30 * time.Second)
		app.update()
		go func() {
			for {
				select {
				case <- app.quit.ClickedCh:
					log.Println("quit")
					systray.Quit()
				case <- app.power.ClickedCh:
					b.power(!app.status.powered)
					app.update()
				case <- app.discovering.ClickedCh:
					b.discovering(!app.status.discovering)
				case <- tick:
					log.Println("tick")
					app.update()
				}
			}
		}()
	}, func() {})
}

func getStatus() (result Status, err error) {
	result, err = b.status()
	if err != nil {
		return
	}

	result.devices, err = b.devices()
	if err != nil {
		return
	}

	result.connected, err = b.devices("Connected")
	if err != nil {
		return
	}

	return
}

func getDevices(output string) (result map[string]string) {
	result = make(map[string]string)
	for _, line := range strings.Split(output, "\n") {
		device, commands, cut := strings.Cut(strings.TrimSpace(line), " ")
		if !cut || device != "Device" {
			continue
		}
		mac, name, cut := strings.Cut(commands, " ")
		if !cut {
			continue
		}
		result[mac] = name
	}
	return
}

func (m *App) update() (err error) {
	m.status, err = getStatus()
	if err != nil {
		log.Println(err)
		return
	}

	if m.status.powered {
		m.power.SetTitle("Turn off")
		if len(m.status.connected) != 0 {
			systray.SetIcon(connected)
		} else if m.status.discovering {
			m.discovering.SetTitle("Silent")
			systray.SetIcon(standby)
		} else {
			m.discovering.SetTitle("Discover")
			systray.SetIcon(on)
		}
	} else {
		m.power.SetTitle("Turn on")
		systray.SetIcon(off)
	}

	for mac, name := range m.status.devices {
		item, exists := m.devices[mac]
		if !exists {
			_, connected := m.status.connected[mac]
			item = menuItem {
				name: name,
				mac: mac,
				connected: connected,
				menuItem: systray.AddMenuItem(name, ""),
			}
			go func() {
				for {
					select {
					case <- item.menuItem.ClickedCh:
						item = m.devices[mac]
						b.connect(mac, !item.connected)
						m.update()
					}
				}
			}()
		} else {
			_, item.connected = m.status.connected[item.mac]
			item.name = name
			item.mac = mac
		}
		if item.connected {
			item.menuItem.SetTitle(item.name + ": connected")
		} else {
			item.menuItem.SetTitle(item.name + ": disconnected")
		}
		m.devices[mac] = item
	}

	return
}
