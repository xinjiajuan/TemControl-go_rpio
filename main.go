package main

import (
	"bytes"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/stianeikeland/go-rpio/v4"
	ini "gopkg.in/ini.v1"
)

func main() {
	arg := os.Args
	if len(arg) != 2 {
		println("请指定配置文件\n格式:temcontrol xxx.ini")
		return
	}
	cfg, err := ini.Load(arg[1])
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
	timer_tem := time.NewTimer(time_refint_dur)
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	var fan_ON bool
	for {
		select {
		case <-timer_tem.C:
			temp := getCPUTemp()
			println("current_temp_is:" + strconv.FormatFloat(temp, 'f', -1, 32))
			if temp >= temp_max_float {
				if !fan_ON {
					pin_io.Mode(rpio.Output)
					fan_ON = true
					println("fan is ON")
				}
			} else if temp <= temp_mix_float {
				if fan_ON {
					pin_io.Mode(rpio.Input)
					fan_ON = false
					println("fan is OFF")
				}
			}
			timer_tem.Reset(time_refint_dur)
		case <-sigs:
			rpio.Close()
			os.Exit(0)
		}
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
