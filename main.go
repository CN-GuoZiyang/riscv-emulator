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

	for cpu.Pc < DRAM_END {
		inst, exception := cpu.Fetch()
		if exception != nil {
			fmt.Println(exception.ToString())
			break
		}
		newPC, exception := cpu.Execute(inst)
		if exception != nil {
			fmt.Println(exception.ToString())
			break
		}
		cpu.Pc = newPC
	}
	cpu.DumpRegisters()
}
