package main

import (
	"fmt"
	"strconv"
	"strings"
)

type Mode uint64

const (
	User       Mode = 0b00
	Supervisor Mode = 0b01
	Machine    Mode = 0b11
)

type Cpu struct {
	Regs [32]uint64
	Pc   uint64
	Mode Mode
	Bus  Bus
	Csr  CSR
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
	regs[2] = DRAM_END
	return &Cpu{
		Regs: regs,
		Pc:   DRAM_BASE,
		Mode: Machine,
		Bus:  NewBus(code),
		Csr:  NewCSR(),
	}
}

func (cpu *Cpu) Reg(r string) uint64 {
	if r == "pc" {
		return cpu.Pc
	}
	if r == "fp" {
		return cpu.Reg("s0")
	}
	for i, abi := range RVABI {
		if abi == r {
			return cpu.Regs[i]
		}
	}
	if strings.HasPrefix(r, "x") {
		indexStr := r[1:]
		index, err := strconv.ParseInt(indexStr, 10, 64)
		if err == nil && index <= 31 {
			return cpu.Regs[index]
		}
	}
	// csr
	switch r {
	case "mhartid":
		return cpu.Csr.Load(MHARTID)
	case "mstatus":
		return cpu.Csr.Load(MSTATUS)
	case "mtvec":
		return cpu.Csr.Load(MTVEC)
	case "mepc":
		return cpu.Csr.Load(MEPC)
	case "mcause":
		return cpu.Csr.Load(MCAUSE)
	case "mtval":
		return cpu.Csr.Load(MTVAL)
	case "medeleg":
		return cpu.Csr.Load(MEDELEG)
	case "mscratch":
		return cpu.Csr.Load(MSCRATCH)
	case "MIP":
		return cpu.Csr.Load(MIP)
	case "mcounteren":
		return cpu.Csr.Load(MCOUNTEREN)
	case "sstatus":
		return cpu.Csr.Load(SSTATUS)
	case "stvec":
		return cpu.Csr.Load(STVEC)
	case "sepc":
		return cpu.Csr.Load(SEPC)
	case "scause":
		return cpu.Csr.Load(SCAUSE)
	case "stval":
		return cpu.Csr.Load(STVAL)
	case "sscratch":
		return cpu.Csr.Load(SSCRATCH)
	case "SIP":
		return cpu.Csr.Load(SIP)
	case "SATP":
		return cpu.Csr.Load(SATP)
	}
	panic(fmt.Sprintf("Invalid registers: %s", r))
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
		imm := uint64(int64(int32(inst & 0xfffff000)))
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
			case 0x01:
				// divu
				if cpu.Regs[rs2] == 0 {
					cpu.Regs[rd] = 0xffffffffffffffff
				} else {
					dividend := cpu.Regs[rs1]
					divisor := cpu.Regs[rs2]
					cpu.Regs[rd] = dividend / divisor
				}
				return cpu.UpdatePC()
			case 0x20:
				// sraw
				cpu.Regs[rd] = uint64(int32(cpu.Regs[rs1]) >> int32(shamt))
				return cpu.UpdatePC()
			default:
				return 0, NewException(IllegalInstruction, inst)
			}
		case 0x7:
			switch funct7 {
			case 0x01:
				// remuw
				if cpu.Regs[rs2] == 0 {
					cpu.Regs[rd] = cpu.Regs[rs1]
				} else {
					dividend := cpu.Regs[rs1]
					divisor := cpu.Regs[rs2]
					cpu.Regs[rd] = uint64(int32(dividend % divisor))
				}
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
		fmt.Printf("imm: %d\n", imm)
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
		t := cpu.Pc + 4
		imm := uint64(int64(int32(inst&0xfff00000)) >> 20)
		newPC := (cpu.Regs[rs1] + imm) & ^(uint64(1))
		cpu.Regs[rd] = t
		return newPC, nil
	case 0x6f:
		// jal
		cpu.Regs[rd] = cpu.Pc + 4
		imm := uint64(int64(int32(inst&0x80000000))>>11) |
			(inst & 0xff000) |
			((inst >> 9) & 0x800) |
			((inst >> 20) & 0x7fe)
		return cpu.Pc + imm, nil
	case 0x73:
		csrAddr := (inst & 0xfff00000) >> 20
		switch funct3 {
		case 0x0:
			if funct7 == 0x9 {
				// sfence.vma
				return cpu.UpdatePC()
			}
			switch rs2 {
			case 0x2:
				switch funct7 {
				case 0x8:
					// sret
					sstatus := cpu.Csr.Load(SSTATUS)
					cpu.Mode = Mode((sstatus & MASK_SPP) >> 8)
					spie := (sstatus & MASK_SPIE) >> 5
					sstatus = (sstatus & ^uint64(MASK_SIE)) | (spie << 1)
					sstatus |= MASK_SPIE
					sstatus &= ^uint64(MASK_SPP)
					cpu.Csr.Store(SSTATUS, sstatus)
					newPC := cpu.Csr.Load(SEPC) & ^uint64(0b11)
					return newPC, nil
				case 0x18:
					// mret
					mstatus := cpu.Csr.Load(MSTATUS)
					cpu.Mode = Mode((mstatus & MASK_MPP) >> 11)
					mpie := (mstatus & MASK_MPIE) >> 7
					mstatus = (mstatus & ^uint64(MASK_MIE)) | (mpie << 3)
					mstatus |= MASK_MPIE
					mstatus &= ^uint64(MASK_MPP)
					mstatus &= ^uint64(MASK_MPRV)
					cpu.Csr.Store(MSTATUS, mstatus)
					newPC := cpu.Csr.Load(MEPC) & ^uint64(0b11)
					return newPC, nil
				default:
					return 0, NewException(IllegalInstruction, inst)
				}
			default:
				return 0, NewException(IllegalInstruction, inst)
			}
		case 0x1:
			// csrrw
			t := cpu.Csr.Load(csrAddr)
			cpu.Csr.Store(csrAddr, cpu.Regs[rs1])
			cpu.Regs[rd] = t
			return cpu.UpdatePC()
		case 0x2:
			// csrrs
			t := cpu.Csr.Load(csrAddr)
			cpu.Csr.Store(csrAddr, t|cpu.Regs[rs1])
			cpu.Regs[rd] = t
			return cpu.UpdatePC()
		case 0x3:
			// csrrc
			t := cpu.Csr.Load(csrAddr)
			cpu.Csr.Store(csrAddr, t & ^cpu.Regs[rs1])
			cpu.Regs[rd] = t
			return cpu.UpdatePC()
		case 0x5:
			// csrrwi
			zimm := rs1
			cpu.Regs[rd] = cpu.Csr.Load(csrAddr)
			cpu.Csr.Store(csrAddr, zimm)
			return cpu.UpdatePC()
		case 0x6:
			// csrrsi
			zimm := rs1
			t := cpu.Csr.Load(csrAddr)
			cpu.Csr.Store(csrAddr, t|zimm)
			cpu.Regs[rd] = t
			return cpu.UpdatePC()
		case 0x7:
			// csrrci
			zimm := rs1
			t := cpu.Csr.Load(csrAddr)
			cpu.Csr.Store(csrAddr, t & ^zimm)
			cpu.Regs[rd] = t
			return cpu.UpdatePC()
		default:
			return 0, NewException(IllegalInstruction, inst)
		}
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
