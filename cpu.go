package main

import "fmt"

type Cpu struct {
	Regs [32]uint64
	Pc   uint64
	Bus  Bus
}

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
		Pc:   DRAM_BASE,
		Bus:  NewBus(code),
	}
}

func (cpu *Cpu) Load(addr, size uint64) (uint64, *Exception) {
	return cpu.Bus.Load(addr, size)
}

func (cpu *Cpu) Store(addr, size, value uint64) *Exception {
	return cpu.Bus.Store(addr, size, value)
}

func (cpu *Cpu) Fetch() (uint64, *Exception) {
	return cpu.Bus.Load(cpu.Pc, 32)
}

func (cpu *Cpu) Execute(inst uint64) {
	opcode := inst & 0x7f
	if opcode == 0x0 {
		// nop
		return
	}
	rd := (inst >> 7) & 0x1f
	rs1 := (inst >> 15) & 0x1f
	rs2 := (inst >> 20) & 0x1f
	fmt.Printf("opcode: 0x%x, rd: %d, rs1: %d, rs2: %d", opcode, rd, rs1, rs2)
	_ = (inst >> 12) & 0x7  // funct3
	_ = (inst >> 25) & 0x7f // funct7

	cpu.Regs[0] = 0

	switch opcode {
	case 0x13:
		// addi
		imm := inst & 0xfff00000 >> 20
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
