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

func (cpu *Cpu) UpdatePC() (uint64, *Exception) {
	return cpu.Pc + 4, nil
}

func (cpu *Cpu) Execute(inst uint64) (uint64, *Exception) {
	opcode := inst & 0x7f
	if opcode == 0x0 {
		// nop
		return cpu.UpdatePC()
	}
	rd := (inst >> 7) & 0x1f
	rs1 := (inst >> 15) & 0x1f
	rs2 := (inst >> 20) & 0x1f
	funct3 := (inst >> 12) & 0x7
	funct7 := (inst >> 25) & 0x7f
	fmt.Printf("opcode: 0x%x, rd: %d, rs1: %d, rs2: %d, funct3: 0x%x, funct7: 0x%x\n", opcode, rd, rs1, rs2, funct3, funct7)

	cpu.Regs[0] = 0

	switch opcode {
	case 0x03:
		// imm = inst[31:20]
		imm := uint64(uint32(inst)) >> 20
		addr := cpu.Regs[rs1] + imm
		switch funct3 {
		case 0x0:
			// lb
			val, err := cpu.Load(addr, 8)
			if err != nil {
				return 0, err
			}
			cpu.Regs[rd] = uint64(int64(int8(val)))
			return cpu.UpdatePC()
		case 0x1:
			// lh
			val, err := cpu.Load(addr, 16)
			if err != nil {
				return 0, err
			}
			cpu.Regs[rd] = uint64(int64(int16(val)))
			return cpu.UpdatePC()
		case 0x2:
			// lw
			val, err := cpu.Load(addr, 32)
			if err != nil {
				return 0, err
			}
			cpu.Regs[rd] = uint64(int64(int32(val)))
			return cpu.UpdatePC()
		case 0x3:
			// ld
			val, err := cpu.Load(addr, 64)
			if err != nil {
				return 0, err
			}
			cpu.Regs[rd] = val
			return cpu.UpdatePC()
		case 0x4:
			// lbu
			val, err := cpu.Load(addr, 8)
			if err != nil {
				return 0, err
			}
			cpu.Regs[rd] = val
			return cpu.UpdatePC()
		case 0x5:
			// lhu
			val, err := cpu.Load(addr, 16)
			if err != nil {
				return 0, err
			}
			cpu.Regs[rd] = val
			return cpu.UpdatePC()
		case 0x6:
			// lwu
			val, err := cpu.Load(addr, 32)
			if err != nil {
				return 0, err
			}
			cpu.Regs[rd] = val
			return cpu.UpdatePC()
		default:
			return 0, NewException(IllegalInstruction, inst)
		}
	case 0x13:
		imm := uint64(int64(int32(inst&0xfff00000)) >> 20)
		shamt := uint32(imm & 0x3f)
		switch funct3 {
		case 0x0:
			// addi
			cpu.Regs[rd] = cpu.Regs[rs1] + imm
			return cpu.UpdatePC()
		case 0x1:
			// slli
			cpu.Regs[rd] = cpu.Regs[rs1] << uint64(shamt)
			return cpu.UpdatePC()
		case 0x2:
			// slti
			if int64(cpu.Regs[rs1]) < int64(imm) {
				cpu.Regs[rd] = 1
			} else {
				cpu.Regs[rd] = 0
			}
			return cpu.UpdatePC()
		case 0x3:
			// sltiu
			if cpu.Regs[rs1] < imm {
				cpu.Regs[rd] = 1
			} else {
				cpu.Regs[rd] = 0
			}
			return cpu.UpdatePC()
		case 0x4:
			// xori
			cpu.Regs[rd] = cpu.Regs[rs1] ^ imm
			return cpu.UpdatePC()
		case 0x5:
			switch funct7 >> 1 {
			case 0x00:
				// srli
				cpu.Regs[rd] = cpu.Regs[rs1] >> shamt
				return cpu.UpdatePC()
			case 0x10:
				// srai
				cpu.Regs[rd] = uint64(int64(cpu.Regs[rs1]) >> shamt)
				return cpu.UpdatePC()
			default:
				return 0, NewException(IllegalInstruction, inst)
			}
		case 0x6:
			// ori
			cpu.Regs[rd] = cpu.Regs[rs1] | imm
			return cpu.UpdatePC()
		case 0x7:
			// andi
			cpu.Regs[rd] = cpu.Regs[rs1] & imm
			return cpu.UpdatePC()
		default:
			return 0, NewException(IllegalInstruction, inst)
		}
	case 0x17:
		// auipc
		imm := uint64(int64(int32(inst&0xfffff000)) >> 20)
		cpu.Regs[rd] = cpu.Pc + imm
		return cpu.UpdatePC()
	case 0x1b:
		imm := uint64(int64(int32(inst)) >> 20)
		shamt := uint32(imm & 0x1f)
		switch funct3 {
		case 0x0:
			// addiw
			cpu.Regs[rd] = uint64(int64(int32(cpu.Regs[rs1] + imm)))
			return cpu.UpdatePC()
		case 0x1:
			// slliw
			cpu.Regs[rd] = uint64(int64(int32(cpu.Regs[rs1] << shamt)))
			return cpu.UpdatePC()
		case 0x5:
			switch funct7 {
			case 0x00:
				// srliw
				cpu.Regs[rd] = uint64(int64(int32(uint32(cpu.Regs[rs1]) >> shamt)))
				return cpu.UpdatePC()
			case 0x20:
				// sraiw
				cpu.Regs[rd] = uint64(int64(int32(cpu.Regs[rs1]) >> shamt))
				return cpu.UpdatePC()
			default:
				return 0, NewException(IllegalInstruction, inst)
			}
		default:
			return 0, NewException(IllegalInstruction, inst)
		}
	case 0x23:
		imm := uint64(int64(int32(inst&0xfe000000))>>20) | ((inst >> 7) & 0x1f)
		addr := cpu.Regs[rs1] + imm
		switch funct3 {
		case 0x0:
			// sb
			err := cpu.Store(addr, 8, cpu.Regs[rs2])
			if err != nil {
				return 0, err
			}
			return cpu.UpdatePC()
		case 0x1:
			// sh
			err := cpu.Store(addr, 16, cpu.Regs[rs2])
			if err != nil {
				return 0, err
			}
			return cpu.UpdatePC()
		case 0x2:
			// sw
			err := cpu.Store(addr, 32, cpu.Regs[rs2])
			if err != nil {
				return 0, err
			}
			return cpu.UpdatePC()
		case 0x3:
			// sd
			err := cpu.Store(addr, 64, cpu.Regs[rs2])
			if err != nil {
				return 0, err
			}
			return cpu.UpdatePC()
		default:
			return 0, NewException(IllegalInstruction, inst)
		}
	case 0x33:
		shamt := uint32(uint64(cpu.Regs[rs2] & 0x3f))
		switch funct3 {
		case 0x0:
			switch funct7 {
			case 0x00:
				// add
				cpu.Regs[rd] = cpu.Regs[rs1] + cpu.Regs[rs2]
				return cpu.UpdatePC()
			case 0x01:
				// mul
				cpu.Regs[rd] = cpu.Regs[rs1] * cpu.Regs[rs2]
				return cpu.UpdatePC()
			case 0x20:
				// sub
				cpu.Regs[rd] = cpu.Regs[rs1] - cpu.Regs[rs2]
				return cpu.UpdatePC()
			default:
				return 0, NewException(IllegalInstruction, inst)
			}
		case 0x1:
			// sll
			cpu.Regs[rd] = cpu.Regs[rs1] << shamt
			return cpu.UpdatePC()
		case 0x2:
			// slt
			if int64(cpu.Regs[rs1]) < int64(cpu.Regs[rs2]) {
				cpu.Regs[rd] = 1
			} else {
				cpu.Regs[rd] = 0
			}
			return cpu.UpdatePC()
		case 0x3:
			// sltu
			if cpu.Regs[rs1] < cpu.Regs[rs2] {
				cpu.Regs[rd] = 1
			} else {
				cpu.Regs[rd] = 0
			}
			return cpu.UpdatePC()
		case 0x4:
			// xor
			cpu.Regs[rd] = cpu.Regs[rs1] ^ cpu.Regs[rs2]
			return cpu.UpdatePC()
		case 0x5:
			switch funct7 {
			case 0x00:
				// srl
				cpu.Regs[rd] = cpu.Regs[rs1] >> shamt
				return cpu.UpdatePC()
			case 0x20:
				// sra
				cpu.Regs[rd] = uint64(int64(cpu.Regs[rs1]) >> shamt)
				return cpu.UpdatePC()
			default:
				return 0, NewException(IllegalInstruction, inst)
			}
		case 0x6:
			// or
			cpu.Regs[rd] = cpu.Regs[rs1] | cpu.Regs[rs2]
			return cpu.UpdatePC()
		case 0x7:
			// and
			cpu.Regs[rd] = cpu.Regs[rs1] & cpu.Regs[rs2]
			return cpu.UpdatePC()
		default:
			return 0, NewException(IllegalInstruction, inst)
		}
	case 0x37:
		// lui
		cpu.Regs[rd] = uint64(int64(int32(inst & 0xfffff000)))
		return cpu.UpdatePC()
	case 0x3b:
		shamt := uint32(cpu.Regs[rs2] & 0x1f)
		switch funct3 {
		case 0x0:
			switch funct7 {
			case 0x00:
				// addw
				cpu.Regs[rd] = uint64(int64(int32(cpu.Regs[rs1] + cpu.Regs[rs2])))
				return cpu.UpdatePC()
			case 0x20:
				// subw
				cpu.Regs[rd] = uint64(int32(cpu.Regs[rs1] - cpu.Regs[rs2]))
				return cpu.UpdatePC()
			default:
				return 0, NewException(IllegalInstruction, inst)
			}
		case 0x1:
			// sllw
			cpu.Regs[rd] = uint64(int32(uint32(cpu.Regs[rs1]) << shamt))
			return cpu.UpdatePC()
		case 0x5:
			switch funct7 {
			case 0x00:
				// srlw
				cpu.Regs[rd] = uint64(int32(uint32(cpu.Regs[rs1]) >> shamt))
				return cpu.UpdatePC()
			case 0x20:
				// sraw
				cpu.Regs[rd] = uint64(int32(cpu.Regs[rs1]) >> int32(shamt))
				return cpu.UpdatePC()
			default:
				return 0, NewException(IllegalInstruction, inst)
			}
		default:
			return 0, NewException(IllegalInstruction, inst)
		}
	case 0x63:
		imm := uint64(int64(int32(inst&0x80000000))>>19) |
			((inst & 0x80) << 4) |
			((inst >> 20) & 0x7e0) |
			((inst >> 7) & 0x1e)
		switch funct3 {
		case 0x0:
			// beq
			if cpu.Regs[rs1] == cpu.Regs[rs2] {
				return cpu.Pc + imm, nil
			}
			return cpu.UpdatePC()
		case 0x1:
			// bne
			if cpu.Regs[rs1] != cpu.Regs[rs2] {
				return cpu.Pc + imm, nil
			}
			return cpu.UpdatePC()
		case 0x4:
			// blt
			if int64(cpu.Regs[rs1]) < int64(cpu.Regs[rs2]) {
				return cpu.Pc + imm, nil
			}
			return cpu.UpdatePC()
		case 0x5:
			// bge
			if int64(cpu.Regs[rs1]) >= int64(cpu.Regs[rs2]) {
				return cpu.Pc + imm, nil
			}
			return cpu.UpdatePC()
		case 0x6:
			// bltu
			if cpu.Regs[rs1] < cpu.Regs[rs2] {
				return cpu.Pc + imm, nil
			}
			return cpu.UpdatePC()
		case 0x7:
			// bgeu
			if cpu.Regs[rs1] >= cpu.Regs[rs2] {
				return cpu.Pc + imm, nil
			}
			return cpu.UpdatePC()
		default:
			return 0, NewException(IllegalInstruction, inst)
		}
	case 0x67:
		// jalr
		cpu.Regs[rd] = cpu.Pc + 4
		imm := uint64(int64(int32(inst&0xfff00000)) >> 20)
		newPC := (cpu.Regs[rs1] + imm) & ^(uint64(1))
		return newPC, nil
	case 0x6f:
		// jal
		cpu.Regs[rd] = cpu.Pc + 4
		imm := uint64(int64(int32(inst&0x80000000))>>11) |
			(inst & 0xff000) |
			((inst >> 9) & 0x800) |
			((inst >> 20) & 0x7fe)
		return cpu.Pc + imm, nil
	default:
		return 0, NewException(IllegalInstruction, inst)
	}
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
