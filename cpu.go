package main

import (
	"fmt"
	"strconv"
	"strings"
	"unsafe"
)

type Mode uint64

const (
	User       Mode = 0b00
	Supervisor Mode = 0b01
	Machine    Mode = 0b11
)

type AccessType uint64

const (
	Instruction AccessType = 0
	Load        AccessType = 1
	Store       AccessType = 2
)

type Cpu struct {
	Regs         [32]uint64
	Pc           uint64
	Mode         Mode
	Bus          Bus
	Csr          CSR
	EnablePaging bool
	PageTable    uint64
}

var (
	RVABI [32]string = [32]string{
		"zero", "ra", "sp", "gp", "tp", "t0", "t1", "t2",
		"s0", "s1", "a0", "a1", "a2", "a3", "a4", "a5",
		"a6", "a7", "s2", "s3", "s4", "s5", "s6", "s7",
		"s8", "s9", "s10", "s11", "t3", "t4", "t5", "t6",
	}
)

func NewCPU(code, diskImage []uint8) *Cpu {
	regs := [32]uint64{}
	regs[2] = DRAM_END
	return &Cpu{
		Regs:         regs,
		Pc:           DRAM_BASE,
		Mode:         Machine,
		Bus:          NewBus(code, diskImage),
		Csr:          NewCSR(),
		EnablePaging: false,
		PageTable:    0,
	}
}

func (cpu *Cpu) Load(addr, size uint64) (uint64, *Exception) {
	pAddr, exception := cpu.Translate(addr, Load)
	if exception != nil {
		return 0, exception
	}
	return cpu.Bus.Load(pAddr, size)
}

func (cpu *Cpu) Store(addr, size, value uint64) *Exception {
	pAddr, exception := cpu.Translate(addr, Store)
	if exception != nil {
		return exception
	}
	return cpu.Bus.Store(pAddr, size, value)
}

func (cpu *Cpu) Fetch() (uint64, *Exception) {
	pAddr, exception := cpu.Translate(cpu.Pc, Store)
	if exception != nil {
		return 0, exception
	}
	if inst, exp := cpu.Bus.Load(pAddr, 32); exp != nil {
		return 0, NewException(InstructionAccessFault, cpu.Pc)
	} else {
		return inst, nil
	}
}

func (cpu *Cpu) UpdatePC() (uint64, *Exception) {
	return cpu.Pc + 4, nil
}

func (cpu *Cpu) HandleException(e *Exception) {
	pc := cpu.Pc
	mode := cpu.Mode
	cause := e.Code()
	trapInSMode := mode <= Supervisor && cpu.Csr.IsMedelegated(cause)
	var (
		STATUS, TVEC, CAUSE, TVAL, EPC, MASK_PIE, pie_i, MASK_IE, ie_i, MASK_PP, pp_i uint64
	)
	if trapInSMode {
		cpu.Mode = Supervisor
		STATUS, TVEC, CAUSE, TVAL, EPC, MASK_PIE, pie_i, MASK_IE, ie_i, MASK_PP, pp_i =
			SSTATUS, STVEC, SCAUSE, STVAL, SEPC, MASK_SPIE, 5, MASK_SIE, 1, MASK_SPP, 8
	} else {
		cpu.Mode = Machine
		STATUS, TVEC, CAUSE, TVAL, EPC, MASK_PIE, pie_i, MASK_IE, ie_i, MASK_PP, pp_i =
			MSTATUS, MTVEC, MCAUSE, MTVAL, MEPC, MASK_MPIE, 7, MASK_MIE, 3, MASK_MPP, 11
	}
	cpu.Pc = cpu.Csr.Load(TVEC) & ^uint64(0b11)
	cpu.Csr.Store(EPC, pc)
	cpu.Csr.Store(CAUSE, cause)
	cpu.Csr.Store(TVAL, e.Value())
	status := cpu.Csr.Load(STATUS)
	ie := (status & MASK_IE) >> ie_i
	status = (status & ^uint64(MASK_PIE)) | (ie << pie_i)
	status &= ^uint64(MASK_IE)
	status = (status & ^uint64(MASK_PP)) | (uint64(mode) << pp_i)
	cpu.Csr.Store(STATUS, status)
}

func (cpu *Cpu) Execute(inst uint64) (uint64, *Exception) {
	opcode := inst & 0x7f
	rd := (inst >> 7) & 0x1f
	rs1 := (inst >> 15) & 0x1f
	rs2 := (inst >> 20) & 0x1f
	funct3 := (inst >> 12) & 0x7
	funct7 := (inst >> 25) & 0x7f
	//fmt.Printf("opcode: 0x%x, rd: %d, rs1: %d, rs2: %d, funct3: 0x%x, funct7: 0x%x\n", opcode, rd, rs1, rs2, funct3, funct7)

	cpu.Regs[0] = 0

	switch opcode {
	case 0x03:
		// imm = inst[31:20]
		imm := uint64(int64(int32(inst)) >> 20)
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
	case 0x0f:
		switch funct3 {
		case 0x0:
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
	case 0x2f:
		funct5 := (funct7 & 0b1111100) >> 2
		switch funct3 {
		case 0x2:
			switch funct5 {
			case 0x00:
				// amoadd.w
				t, e := cpu.Load(cpu.Regs[rs1], 32)
				if e != nil {
					return 0, e
				}
				e = cpu.Store(cpu.Regs[rs1], 32, t+cpu.Regs[rs2])
				if e != nil {
					return 0, e
				}
				cpu.Regs[rd] = t
				return cpu.UpdatePC()
			case 0x01:
				// amoswap.w
				t, e := cpu.Load(cpu.Regs[rs1], 32)
				if e != nil {
					return 0, e
				}
				e = cpu.Store(cpu.Regs[rs1], 32, cpu.Regs[rs2])
				if e != nil {
					return 0, e
				}
				cpu.Regs[rd] = t
				return cpu.UpdatePC()
			default:
				return 0, NewException(IllegalInstruction, inst)
			}
		case 0x3:
			switch funct5 {
			case 0x00:
				// amoadd.d
				t, e := cpu.Load(cpu.Regs[rs1], 64)
				if e != nil {
					return 0, e
				}
				e = cpu.Store(cpu.Regs[rs1], 64, t+cpu.Regs[rs2])
				if e != nil {
					return 0, e
				}
				cpu.Regs[rd] = t
				return cpu.UpdatePC()
			case 0x01:
				// amoswap.d
				t, e := cpu.Load(cpu.Regs[rs1], 64)
				if e != nil {
					return 0, e
				}
				e = cpu.Store(cpu.Regs[rs1], 64, cpu.Regs[rs2])
				if e != nil {
					return 0, e
				}
				cpu.Regs[rd] = t
				return cpu.UpdatePC()
			default:
				return 0, NewException(IllegalInstruction, inst)
			}
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
			case 0x0:
				switch cpu.Mode {
				case User:
					return 0, NewException(EnvironmentCallFromUMode, cpu.Pc)
				case Supervisor:
					return 0, NewException(EnvironmentCallFromSMode, cpu.Pc)
				case Machine:
					return 0, NewException(EnvironmentCallFromMMode, cpu.Pc)
				default:
					panic("Invalud mode")
				}
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
			cpu.UpdatePaging(csrAddr)
			return cpu.UpdatePC()
		case 0x2:
			// csrrs
			t := cpu.Csr.Load(csrAddr)
			cpu.Csr.Store(csrAddr, t|cpu.Regs[rs1])
			cpu.Regs[rd] = t
			cpu.UpdatePaging(csrAddr)
			return cpu.UpdatePC()
		case 0x3:
			// csrrc
			t := cpu.Csr.Load(csrAddr)
			cpu.Csr.Store(csrAddr, t & ^cpu.Regs[rs1])
			cpu.Regs[rd] = t
			cpu.UpdatePaging(csrAddr)
			return cpu.UpdatePC()
		case 0x5:
			// csrrwi
			zimm := rs1
			cpu.Regs[rd] = cpu.Csr.Load(csrAddr)
			cpu.Csr.Store(csrAddr, zimm)
			cpu.UpdatePaging(csrAddr)
			return cpu.UpdatePC()
		case 0x6:
			// csrrsi
			zimm := rs1
			t := cpu.Csr.Load(csrAddr)
			cpu.Csr.Store(csrAddr, t|zimm)
			cpu.Regs[rd] = t
			cpu.UpdatePaging(csrAddr)
			return cpu.UpdatePC()
		case 0x7:
			// csrrci
			zimm := rs1
			t := cpu.Csr.Load(csrAddr)
			cpu.Csr.Store(csrAddr, t & ^zimm)
			cpu.Regs[rd] = t
			cpu.UpdatePaging(csrAddr)
			return cpu.UpdatePC()
		default:
			return 0, NewException(IllegalInstruction, inst)
		}
	default:
		return 0, NewException(IllegalInstruction, inst)
	}
}

func (cpu *Cpu) HandleInterrupt(interrupt Interrupt) {
	pc := cpu.Pc
	mode := cpu.Mode
	cause := interrupt.Code()
	trapInSMode := mode <= Supervisor && cpu.Csr.IsMidelegated(cause)
	var STATUS, TVEC, CAUSE, TVAL, EPC, MASK_PIE, pie_i, MASK_IE, ie_i, MASK_PP, pp_i uint64
	if trapInSMode {
		cpu.Mode = Supervisor
		STATUS, TVEC, CAUSE, TVAL, EPC, MASK_PIE, pie_i, MASK_IE, ie_i, MASK_PP, pp_i =
			SSTATUS, STVEC, SCAUSE, STVAL, SEPC, MASK_SPIE, 5, MASK_SIE, 1, MASK_SPP, 8
	} else {
		cpu.Mode = Machine
		STATUS, TVEC, CAUSE, TVAL, EPC, MASK_PIE, pie_i, MASK_IE, ie_i, MASK_PP, pp_i =
			MSTATUS, MTVEC, MCAUSE, MTVAL, MEPC, MASK_MPIE, 7, MASK_MIE, 3, MASK_MPP, 11
	}
	tvec := cpu.Csr.Load(TVEC)
	tvecMode := tvec & 0b11
	tvecBase := tvec & ^uint64(0b11)
	switch tvecMode {
	case 0:
		cpu.Pc = tvecBase
	case 1:
		cpu.Pc = tvecBase + cause<<2
	}
	cpu.Csr.Store(EPC, pc)
	cpu.Csr.Store(CAUSE, cause)
	cpu.Csr.Store(TVAL, 0)
	status := cpu.Csr.Load(STATUS)
	ie := (status & MASK_IE) >> ie_i
	status = (status & ^MASK_PIE) | (ie << pie_i)
	status &= ^MASK_IE
	status = (status & ^MASK_PP) | (uint64(mode) << pp_i)
	cpu.Csr.Store(STATUS, status)
}

func (cpu *Cpu) CheckPendingInterrupt() *Interrupt {
	if cpu.Mode == Machine && (cpu.Csr.Load(MSTATUS)&MASK_MIE) == 0 {
		return nil
	}
	if cpu.Mode == Supervisor && (cpu.Csr.Load(SSTATUS)&MASK_SIE) == 0 {
		return nil
	}
	if cpu.Bus.uart.IsInterrupting() {
		cpu.Bus.Store(PLIC_SCLAIM, 32, UART_IRQ)
		cpu.Csr.Store(MIP, cpu.Csr.Load(MIP)|MASK_SEIP)
	} else if cpu.Bus.virtioBlock.IsInterrupting() {
		cpu.DiskAccess()
		cpu.Bus.Store(PLIC_SCLAIM, 32, VIRTIO_IRQ)
		cpu.Csr.Store(MIP, cpu.Csr.Load(MIP)|MASK_SEIP)
	}
	pending := cpu.Csr.Load(MIE) & cpu.Csr.Load(MIP)
	if (pending & MASK_MEIP) != 0 {
		cpu.Csr.Store(MIP, cpu.Csr.Load(MIP) & ^uint64(MASK_MEIP))
		return &MachineExternalInterrupt
	}
	if (pending & MASK_MSIP) != 0 {
		cpu.Csr.Store(MIP, cpu.Csr.Load(MIP) & ^uint64(MASK_MSIP))
		return &MachineSoftwareInterrupt
	}
	if (pending & MASK_MTIP) != 0 {
		cpu.Csr.Store(MIP, cpu.Csr.Load(MIP) & ^uint64(MASK_MTIP))
		return &MachineTimerInterrupt
	}
	if (pending & MASK_SEIP) != 0 {
		cpu.Csr.Store(MIP, cpu.Csr.Load(MIP) & ^uint64(MASK_SEIP))
		return &SupervisorExternalInterrupt
	}
	if (pending & MASK_SSIP) != 0 {
		cpu.Csr.Store(MIP, cpu.Csr.Load(MIP) & ^uint64(MASK_SSIP))
		return &SupervisorSoftwareInterrupt
	}
	if (pending & MASK_STIP) != 0 {
		cpu.Csr.Store(MIP, cpu.Csr.Load(MIP) & ^uint64(MASK_STIP))
		return &SupervisorTimerInterrupt
	}
	return nil
}

func (cpu *Cpu) DiskAccess() {
	descSize := uint64(unsafe.Sizeof(VirtqDesc{}))
	descAddr := cpu.Bus.virtioBlock.DescAddr()
	availAddr := descAddr + DESC_NUM*descSize
	usedAddr := descAddr + PAGE_SIZE

	virtqAvail := VirtqAvail{}
	virtqUsed := VirtqUsed{}

	idx, _ := cpu.Bus.Load(availAddr+(uint64(uintptr(unsafe.Pointer(&virtqAvail.idx)))-uint64(uintptr(unsafe.Pointer(&virtqAvail)))), 16)
	index, _ := cpu.Bus.Load(availAddr+(uint64(uintptr(unsafe.Pointer(&virtqAvail.ring[idx%DESC_NUM])))-uint64(uintptr(unsafe.Pointer(&virtqAvail)))), 16)

	descAddr0 := descAddr + descSize*index
	virtqDesc0 := VirtqDesc{}

	reqAddr, _ := cpu.Bus.Load(descAddr0+(uint64(uintptr(unsafe.Pointer(&virtqDesc0.addr)))-uint64(uintptr(unsafe.Pointer(&virtqDesc0)))), 64)
	virtqBlkReq := VirtioBlkRequest{}
	blkSector, _ := cpu.Bus.Load(reqAddr+(uint64(uintptr(unsafe.Pointer(&virtqBlkReq.sector)))-uint64(uintptr(unsafe.Pointer(&virtqBlkReq)))), 64)
	ioType, _ := cpu.Bus.Load(reqAddr+(uint64(uintptr(unsafe.Pointer(&virtqBlkReq.iotype)))-uint64(uintptr(unsafe.Pointer(&virtqBlkReq)))), 32)
	next0, _ := cpu.Bus.Load(descAddr0+(uint64(uintptr(unsafe.Pointer(&virtqDesc0.next)))-uint64(uintptr(unsafe.Pointer(&virtqDesc0)))), 16)

	descAddr1 := descAddr + descSize*next0
	virtqDesc1 := VirtqDesc{}
	addr1, _ := cpu.Bus.Load(descAddr1+(uint64(uintptr(unsafe.Pointer(&virtqDesc1.addr)))-uint64(uintptr(unsafe.Pointer(&virtqDesc1)))), 64)
	len1, _ := cpu.Bus.Load(descAddr1+(uint64(uintptr(unsafe.Pointer(&virtqDesc1.length)))-uint64(uintptr(unsafe.Pointer(&virtqDesc1)))), 32)
	switch ioType {
	case VIRTIO_BLK_T_OUT:
		for i := uint64(0); i < len1; i++ {
			data, _ := cpu.Bus.Load(addr1+i, 8)
			cpu.Bus.virtioBlock.WriteDisk(blkSector*SECTOR_SIZE+i, data)
		}
	case VIRTIO_BLK_T_IN:
		for i := uint64(0); i < len1; i++ {
			data := cpu.Bus.virtioBlock.ReadDisk(blkSector*SECTOR_SIZE + i)
			_ = cpu.Bus.Store(addr1+i, 8, data)
		}
	}
	newID := cpu.Bus.virtioBlock.GetNewID()
	cpu.Bus.Store(usedAddr+(uint64(uintptr(unsafe.Pointer(&virtqUsed.idx)))-uint64(uintptr(unsafe.Pointer(&virtqUsed)))), 16, newID%8)
}

func (cpu *Cpu) UpdatePaging(csrAddr uint64) {
	if csrAddr != SATP {
		return
	}
	satp := cpu.Csr.Load(SATP)
	cpu.PageTable = (satp & MASK_PPN) * PAGE_SIZE
	mode := satp >> 60
	cpu.EnablePaging = mode == 8
}

func (cpu *Cpu) Translate(addr uint64, accessType AccessType) (uint64, *Exception) {
	if !cpu.EnablePaging {
		return addr, nil
	}
	levels := 3
	vpn := []uint64{
		(addr >> 12) & 0x1ff,
		(addr >> 21) & 0x1ff,
		(addr >> 30) & 0x1ff,
	}
	a := cpu.PageTable
	i := levels - 1
	var pte uint64
	for {
		var exception *Exception
		pte, exception = cpu.Bus.Load(a+vpn[i]*8, 64)
		if exception != nil {
			return 0, exception
		}
		v := pte & 1
		r := (pte >> 1) & 1
		w := (pte >> 2) & 1
		x := (pte >> 3) & 1
		if v == 0 || (r == 0 && w == 1) {
			switch accessType {
			case Instruction:
				return 0, NewException(InstructionPageFault, addr)
			case Load:
				return 0, NewException(LoadPageFault, addr)
			case Store:
				return 0, NewException(StoreAMOPageFault, addr)
			}
		}
		if r == 1 || x == 1 {
			break
		}
		i -= 1
		ppn := (pte >> 10) & 0x0fff_ffff_ffff
		a = ppn * PAGE_SIZE
		if i < 0 {
			switch accessType {
			case Instruction:
				return 0, NewException(InstructionPageFault, addr)
			case Load:
				return 0, NewException(LoadPageFault, addr)
			case Store:
				return 0, NewException(StoreAMOPageFault, addr)
			}
		}
	}

	ppn := []uint64{
		(pte >> 10) & 0x1ff,
		(pte >> 19) & 0x1ff,
		(pte >> 28) & 0x03ff_ffff,
	}

	offset := addr & 0xfff

	switch i {
	case 0:
		ppn := (pte >> 10) & 0x0fff_ffff_ffff
		return (ppn << 12) | offset, nil
	case 1, 2:
		return (ppn[2] << 30) | (ppn[1] << 21) | (vpn[0] << 12) | offset, nil
	default:
		switch accessType {
		case Instruction:
			return 0, NewException(InstructionPageFault, addr)
		case Load:
			return 0, NewException(LoadPageFault, addr)
		case Store:
			return 0, NewException(StoreAMOPageFault, addr)
		default:
			panic("Illegal access type")
		}
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
