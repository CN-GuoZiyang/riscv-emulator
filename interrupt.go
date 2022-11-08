package main

const (
	MASK_INTERRUPT_BIT = 1 << 63
)

type Interrupt uint64

var (
	SupervisorSoftwareInterrupt Interrupt = 1
	MachineSoftwareInterrupt    Interrupt = 3
	SupervisorTimerInterrupt    Interrupt = 5
	MachineTimerInterrupt       Interrupt = 7
	SupervisorExternalInterrupt Interrupt = 9
	MachineExternalInterrupt    Interrupt = 11
)

func (i Interrupt) Code() uint64 {
	return uint64(i | MASK_INTERRUPT_BIT)
}
