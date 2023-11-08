package main

import (
	"encoding/base64"
	"fmt"
	"log"
	"log/syslog"
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

var (
	icons = map[string]icon{
		"logo": {
			Base64: "iVBORw0KGgoAAAANSUhEUgAAACAAAAAgCAYAAABzenr0AAAFjElEQVR42qWVA5dcWReGv//x2dbYVuy0baOq2raNUmwnYy6M7ZnYTju4qHTdPfute6rv6ppbFZ21nqR5nvecs/fuX5isXzFPMCuYOCb1DokTez0h9g65Hrzv8ZWu2IzmE1m23qt5ZYPTlsphKq5bRSX15lQ2rTankb9fM0y2qqHpvLL+a0l5racefDrMDUcw+a/+++BSe4rVrbW4vybnvhFa96ZEG95WaWMg76i0SbD7fY+PPQHsFex+T6bNb45S/6ZvKKfMrd318FJ7sJt48tkVtfvz278gq0Oi+LarFN10lSKZKEF08zUfMS06sUzuoOQjT5DvZ0jHYpeo2CVT2zaFaoe/pgVRtfvhMguw8pFlbeOpred5cyEGkAaKQes1imPMpKBgGMhUCOwyFTllKnddornR7eNwmQVIe3h5t5radQUnD5ACSIUYtF2jeAbiYFKLH4dMNqZyjURzYnsUuMwC5D8aNuhN75YoxvS0AFJB+zVKYCrWyEHFVoeOzanfQM06heYnDk6zK88sgPXxMLuW3iMFPS2AFCSCDonOj2lUs14xxJAaYl0OXDLV8s8tSnZo7LKYBngywk4ZvVKQ0wJIdXES6JRo2kt0aUKjug2KqRQUA7fs+5klKU6CyyyA7alIO2X2SUFOa0hBMujSA2BdRIiNCosNKShxA4VKVym+7y9JdRJcQQNkcYBgp00WpHTppIoA/nVpUqPGTQoVC6lfDMqY+k03CPB0lJ1y+iVzsZDielFwqVysaYw/wIVxbeYmGjerhni1AQIsDRXgmWgH5Q7IpqcFGDinR7x0nmXlq2VK7zECQIqCxEJNNG9RZ8nLmYbNHCDtBgHyBuVZYpzUf1q87ZkR3Th1TaMWlvgDDOxTafKqNvMU7dtVSAG3qk4TB1gWKsCzMQ4eKHKAmOGTpoNefdjsP6VbVQ8RlP6PsY6f91LDJhZDKqhcq4NbWZ4eIsBzsQ68r5DOFmf0yoC7RKYcfqaPfpymwPXNMS9mQoBYpSqwTqWWrRwgI0SA5zkABoohBZAaZPXr5DDHLxgtgPevglDgl1b7Wa9S6zaVVtwgACYZpMCQCnE24NPjmT7e//Mb+JZvoG4DxIa0RlDLX2/jACtDBXghzklWpxJwWkOcw6AND5w2qYHr+v8nLnBxbpstBnUbVV9hht0ogI0DGFIhBoOy72rPjmozXdC5w+gCx8se39ewRqY06t3jEWKdeqaDfz4sM1SAeCe3mmJImdxBnTwGEw4BMGywYcGw4g/Am3tmhtEIt2HXLg/VQ7zJj4c6dnooLMsVPMCcBCdLFJzWEA8BvLtC+SzEu6K3C+06XuNvwYy8G3JDzANIB6HCQwWYywEwvw0xpDoFAFKBxQGMAFijfPU9u4UcUkHjFh0Ei8gOEWAeB8D89p+2IITY6gRGAH53CAypIaamrTrdu28UIN6pla0OIRVimx+XHmDUV3TXITakQtwsaGF6ECDLqQULYJkXP+zFCLWYntaQFvlQuV5UzH3IA09riAWt2zy+zgjPGJpmV6FZgJw5Mb1qxRpJP+kNxKBklUo9e8xPa4h12rYjgEIr0voUduWaBUh4LqJjrHj4IkRBpIYYlDImYkMqxO1M5w7+f8soLUvqHIPLLMDcxxaXfZfZ9CXhGdANxS4TKVitUpnAVLxdiAGLO3fq118+8DXNCav4Di6zAL/503+ecoZlujRr99fUuH6E2rdJ6F0f3WC3QY/A/dp1H6vA6zqrwRs67ldlGtw9SvXObyjZ6vb+7X9PO9j1W8Z0Pfb3/z/rWhRVvT8ivXM8JrtPicsb0pIsbkq2rfKRIkgVZJXpZAtyBFmlbkq32bWUwgElIa97Iiyp9uB/7n3WBQcTcv1GXFECk81YGNttYhF7JDBzsHeg7CcemFo8XXAOlwAAAABJRU5ErkJggg==",
		},
		// "red": {
		// 	Base64: "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",
		// },
		// "green": {
		// 	Base64: "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",
		// },
	}
	menuItems = make(map[string]menuItem)
)

func main() {
	syslog, err := syslog.New(syslog.LOG_INFO, "bluetooth-menu")
	if err != nil {
		panic("Unable to connect to syslog")
	}
	log.SetOutput(syslog)

	systray.Run(func() {
		systray.SetIcon(getIcon("logo"))
		quit := systray.AddMenuItem("Quit", "")
		go func() {
			for {
				select {
				case <-quit.ClickedCh:
					systray.Quit()
				}
			}
		}()
		systray.AddSeparator()
		addMenuItems()
		tick := time.Tick(30 * time.Second)
		for {
			select {
			case <-tick:
				addMenuItems()
			}
		}
	}, func() {})
}

func addMenuItems() {
	allDevices, err := exec.Command("sh", "-c", "bluetoothctl devices | awk '{printf $2 \"\t\"; for (i=3; i<NF; i++) printf $i \" \"; print $NF}'").Output()
	if err != nil {
		log.Println("all")
		log.Println(err)
		return
	}
	var currentDevices []string
	for _, line := range strings.Split(strings.TrimSpace(string(allDevices)), "\n") {
		parts := strings.Split(line, "\t")
		currentDevices = append(currentDevices, parts[0])
		item, ok := menuItems[parts[0]]
		if !ok {
			newItem := systray.AddMenuItem(parts[1], "")
			go func() {
				for {
					select {
					case <-newItem.ClickedCh:
						if item, ok := menuItems[parts[0]]; ok {
							if item.connected {
								exec.Command("sh", "-c", "bluetoothctl disconnect "+parts[0]).Run()
								item.connected = false
								item.menuItem.SetTitle(fmt.Sprintf("%s: Disconnected", item.name))
							} else {
								exec.Command("sh", "-c", "bluetoothctl connect "+parts[0]).Run()
								item.connected = true
								item.menuItem.SetTitle(fmt.Sprintf("%s: Connected", item.name))
							}
							menuItems[parts[0]] = item
						}
					}
				}
			}()
			item = menuItem{
				mac:      parts[0],
				name:     parts[1],
				menuItem: newItem,
			}
			menuItems[parts[0]] = item
		}
		item.menuItem.Show()
	}
	for mac, item := range menuItems {
		current := false
		for _, currcurrentDevice := range currentDevices {
			if mac == currcurrentDevice {
				current = true
				break
			}
		}
		if !current {
			item.menuItem.Hide()
		}
	}
	connectedDevices, err := exec.Command("sh", "-c", "bluetoothctl devices Connected | awk '{printf $2}'").Output()
	if err != nil {
		log.Println(err)
		return
	}
	connected := strings.Split(strings.TrimSpace(string(connectedDevices)), "\n")
	for mac, item := range menuItems {
		isConnected := false
		status := "Disconnected"
		for _, connectedMac := range connected {
			if mac == connectedMac {
				isConnected = true
				status = "Connected"
				break
			}
		}
		item.connected = isConnected
		item.menuItem.SetTitle(fmt.Sprintf("%s: %s", item.name, status))
		menuItems[mac] = item
	}
}

func decodedIcon(icon string) ([]byte, error) {
	if len(icons[icon].Decoded) < 1 {
		img, err := base64.StdEncoding.DecodeString(icons[icon].Base64)
		if err != nil {
			return []byte(" "), fmt.Errorf("failed to get icon: %s", icon)
		}
		i := icons[icon]
		i.Decoded = img
		icons[icon] = i
	}

	return icons[icon].Decoded, nil
}

func getIcon(icon string) []byte {
	img, err := decodedIcon(icon)
	if err != nil {
		log.Println(err)
	}

	return img
}
