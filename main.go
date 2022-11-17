package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
)

func main() {
	args := os.Args
	if len(args) != 2 && len(args) != 3 {
		fmt.Printf("run with <filename> <(optional) image>")
		return
	}

	code, err := os.ReadFile(args[1])
	if err != nil {
		panic("read file error!")
	}

	var diskImage []uint8
	if len(args) == 3 {
		diskImage, err = os.ReadFile(args[2])
		if err != nil {
			panic("read file error!")
		}
	}

	cpu := NewCPU(code, diskImage)

	// 关闭终端缓冲
	exec.Command("stty", "-F", "/dev/tty", "cbreak", "min", "1").Run()
	// 关闭终端显示
	exec.Command("stty", "-F", "/dev/tty", "-echo").Run()
	// 恢复终端显示
	defer exec.Command("stty", "-F", "/dev/tty", "echo").Run()

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		// 恢复终端显示
		exec.Command("stty", "-F", "/dev/tty", "echo").Run()
		os.Exit(0)
	}()

	for {
		inst, exception := cpu.Fetch()
		if exception != nil {
			cpu.HandleException(exception)
			if exception.IsFatal() {
				fmt.Println(exception.ToString())
				break
			}
			continue
		}
		newPC, exception := cpu.Execute(inst)
		if exception != nil {
			cpu.HandleException(exception)
			if exception.IsFatal() {
				fmt.Println(exception.ToString())
				break
			}
		} else {
			cpu.Pc = newPC
		}
		if interrupt := cpu.CheckPendingInterrupt(); interrupt != nil {
			cpu.HandleInterrupt(*interrupt)
		}
	}
}
