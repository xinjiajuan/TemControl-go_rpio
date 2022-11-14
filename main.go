package main

import (
	"bytes"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/stianeikeland/go-rpio/v4"
	ini "gopkg.in/ini.v1"
)

func main() {
	cfg, err := ini.Load("conf.ini")
	if err != nil {
		panic(err.Error())
		return
	}
	temp_mix := cfg.Section("temp_conf").Key("temp_Mix").String()
	temp_max := cfg.Section("temp_conf").Key("temp_Max").String()
	time_refint := cfg.Section("temp_conf").Key("RefreshInterval").String()
	pin, err := strconv.Atoi(cfg.Section("temp_conf").Key("pin").String())
	if err != nil {
		panic(err.Error())
		return
	}
	pin_io := rpio.Pin(pin)
	if err := rpio.Open(); err != nil {
		panic(err.Error())
		return
	}
	temp_mix_float, err := strconv.ParseFloat(temp_mix, 32)
	if err != nil {
		panic(err.Error())
		return
	}
	temp_max_float, err := strconv.ParseFloat(temp_max, 32)
	if err != nil {
		panic(err.Error())
		return
	}
	time_refint_dur, err := time.ParseDuration(time_refint)
	if err != nil {
		panic(err.Error())
		return
	}
	for {
		temp := getCPUTemp()
		println("current_temp_is:" + strconv.FormatFloat(temp, 'f', -1, 32))
		if temp >= temp_max_float {
			pin_io.Output()
			println("fans is ON")
		} else if temp <= temp_mix_float {
			pin_io.Input()
			println("fans is OFF")
		}
		time.Sleep(time_refint_dur)
	}
}
func getCPUTemp() float64 {
	cmd := exec.Command("cat", "/sys/class/thermal/thermal_zone0/temp")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		panic(err.Error())
		return 0
	}
	tempStr := strings.Replace(out.String(), "\n", "", -1)
	temp, err := strconv.ParseFloat(tempStr, 32)
	if err != nil {
		panic(err.Error())
		return 0
	}
	return temp / 1000
}
