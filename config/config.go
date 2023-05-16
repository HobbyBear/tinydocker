package config

import (
	"fmt"
	"github.com/ThreeKing2018/gocolor"
)

const (
	IpAmStorageFsPath = "/root/subnet.json"
	NetStoragePath    = "/root/network.json"
)

func Title() string {
	return fmt.Sprintf("%s %s %s %s %s %s ",
		gocolor.SRedBG("welcome"),
		gocolor.SGreenBG("to"),
		gocolor.SYellowBG("use"),
		gocolor.SBlueBG("tinydocker"),
		"🦍 🦍 🦍",
		"❗️❗")
}
