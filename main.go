package main

import (
	"fmt"
	"os"
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
