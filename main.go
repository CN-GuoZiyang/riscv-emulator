package main

import (
	"fmt"
	"os"
)

func main() {
	args := os.Args
	if len(args) != 2 {
		fmt.Printf("run with <filename>")
		return
	}

	code, err := os.ReadFile(args[1])
	if err != nil {
		panic("read file error!")
	}

	cpu := NewCPU(code)

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
			continue
		}
		cpu.Pc = newPC
	}
}
