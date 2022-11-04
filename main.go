package main

import (
	"fmt"
	"io/ioutil"
	"os"
)

type Cpu struct {
	Regs [32]uint64
	Pc   uint64
	Dram []uint8
}

const (
	DRAM_SIZE = 1024 * 1024 * 1024
)

var (
	RVABI [32]string = [32]string{
		"zero", "ra", "sp", "gp", "tp", "t0", "t1", "t2",
		"s0", "s1", "a0", "a1", "a2", "a3", "a4", "a5",
		"a6", "a7", "s2", "s3", "s4", "s5", "s6", "s7",
		"s8", "s9", "s10", "s11", "t3", "t4", "t5", "t6",
	}
)

func NewCPU(code []uint8) *Cpu {
	regs := [32]uint64{}
	regs[2] = DRAM_SIZE - 1
	return &Cpu{
		Regs: regs,
		Pc:   0,
		Dram: code,
	}
}

func (cpu *Cpu) Fetch() uint32 {
	index := cpu.Pc
	inst := uint32(cpu.Dram[index]) |
		(uint32(cpu.Dram[index+1]) << 8) |
		(uint32(cpu.Dram[index+2]) << 16) |
		(uint32(cpu.Dram[index+3]) << 24)
	return inst
}

func (cpu *Cpu) Execute(inst uint32) {
	opcode := inst & 0x7f
	rd := uint64((inst >> 7) & 0x1f)
	rs1 := uint64((inst >> 15) & 0x1f)
	rs2 := uint64((inst >> 20) & 0x1f)
	fmt.Printf("opcode: 0x%x, rd: %d, rs1: %d, rs2: %d", opcode, rd, rs1, rs2)
	_ = (inst >> 12) & 0x7  // funct3
	_ = (inst >> 25) & 0x7f // funct7

	cpu.Regs[0] = 0

	switch opcode {
	case 0x13:
		// addi
		imm := uint64(inst&0xfff00000) >> 20
		fmt.Printf(", imm: %d", imm)
		cpu.Regs[rd] = cpu.Regs[rs1] + imm
	case 0x33:
		// add
		cpu.Regs[rd] = cpu.Regs[rs1] + cpu.Regs[rs2]
	default:
		panic(fmt.Sprintf("Invalid opcode: [%x]", opcode))
	}
	println()
}

func (cpu *Cpu) DumpRegisters() {
	fmt.Println("registers:")
	for i := 0; i < 32; i += 4 {
		i0 := fmt.Sprintf("x%d", i)
		i1 := fmt.Sprintf("x%d", i+1)
		i2 := fmt.Sprintf("x%d", i+2)
		i3 := fmt.Sprintf("x%d", i+3)
		fmt.Printf(
			"%s(%s) = 0x%X\t%s(%s) = 0x%X\t%s(%s) = 0x%X\t%s(%s) = 0x%X\n",
			i0, RVABI[i], cpu.Regs[i],
			i1, RVABI[i+1], cpu.Regs[i+1],
			i2, RVABI[i+2], cpu.Regs[i+2],
			i3, RVABI[i+3], cpu.Regs[i+3],
		)
	}
}

func main() {
	args := os.Args
	if len(args) != 2 {
		fmt.Printf("run with <filename>")
		return
	}
	file, err := os.Open(args[1])
	if err != nil {
		panic(fmt.Sprintf("open file error, err:[%s]", err.Error()))
	}
	defer file.Close()

	code, err := ioutil.ReadAll(file)
	if err != nil {
		panic("read file error!")
	}

	cpu := NewCPU(code)

	for cpu.Pc < uint64(len(cpu.Dram)) {
		inst := cpu.Fetch()
		cpu.Execute(inst)
		cpu.Pc += 4
	}
	cpu.DumpRegisters()
}
