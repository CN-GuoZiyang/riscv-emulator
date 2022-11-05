package main

import "fmt"

type ExceptionType uint64

const (
	InstructionAddrMisaligned ExceptionType = 0
	InstructionAccessFault    ExceptionType = 1
	IllegalInstruction        ExceptionType = 2
	Breakpoint                ExceptionType = 3
	LoadAccessMisaligned      ExceptionType = 4
	LoadAccessFault           ExceptionType = 5
	StoreAMOAddrMisaligned    ExceptionType = 6
	StoreAMOAccessFault       ExceptionType = 7
	EnvironmentCallFromUMode  ExceptionType = 8
	EnvironmentCallFromSMode  ExceptionType = 9
	EnvironmentCallFromMMode  ExceptionType = 11
	InstructionPageFault      ExceptionType = 12
	LoadPageFault             ExceptionType = 13
	StoreAMOPageFault         ExceptionType = 15
)

type Exception struct {
	Type  ExceptionType
	Store uint64
}

func (e Exception) ToString() string {
	switch e.Type {
	case InstructionAddrMisaligned:
		return fmt.Sprint("Instruction address misaligned %X", e.Store)
	case InstructionAccessFault:
		return fmt.Sprint("Instruction access fault %X", e.Store)
	case IllegalInstruction:
		return fmt.Sprint("Illegal instruction %X", e.Store)
	case Breakpoint:
		return fmt.Sprint("Breakpoint  %X", e.Store)
	case LoadAccessMisaligned:
		return fmt.Sprint("Load access %X", e.Store)
	case LoadAccessFault:
		return fmt.Sprint("Load access fault %X", e.Store)
	case StoreAMOAddrMisaligned:
		return fmt.Sprint("Store or AMO address misaliged %X", e.Store)
	case StoreAMOAccessFault:
		return fmt.Sprint("Store or AMO access fault %X", e.Store)
	case EnvironmentCallFromUMode:
		return fmt.Sprint("Environment call from U-mode {%X", e.Store)
	case EnvironmentCallFromSMode:
		return fmt.Sprint("Environment call from S-mode {%X", e.Store)
	case EnvironmentCallFromMMode:
		return fmt.Sprint("Environment call from M-mode {%X", e.Store)
	case InstructionPageFault:
		return fmt.Sprint("Instruction page fault %X", e.Store)
	case LoadPageFault:
		return fmt.Sprint("Load page fault %X", e.Store)
	case StoreAMOPageFault:
		return fmt.Sprint("Store or AMO page fault %X", e.Store)
	}
	panic("Unknown Exception Type!")
}

func (e Exception) Value() uint64 {
	return e.Store
}

func (e Exception) Code() uint64 {
	return uint64(e.Type)
}

func (e Exception) IsFatal() bool {
	switch e.Type {
	case InstructionAddrMisaligned, InstructionAccessFault, IllegalInstruction, LoadAccessFault, StoreAMOAddrMisaligned, StoreAMOAccessFault:
		return true
	}
	return false
}

func NewException(t ExceptionType, v uint64) *Exception {
	return &Exception{
		Type:  t,
		Store: v,
	}
}
