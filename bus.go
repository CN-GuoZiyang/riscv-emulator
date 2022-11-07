package main

type Bus struct {
	dram  Dram
	plic  Plic
	clint Clint
	uart  Uart
}

func NewBus(code []uint8) Bus {
	return Bus{
		dram:  NewDram(code),
		plic:  NewPlic(),
		clint: NewClint(),
		uart:  NewUart(),
	}
}

func (b *Bus) Load(addr, size uint64) (uint64, *Exception) {
	switch {
	case addr >= CLINT_BASE && addr <= CLINT_END:
		return b.clint.Load(addr, size)
	case addr >= PLIC_BASE && addr <= PLIC_END:
		return b.plic.Load(addr, size)
	case addr >= DRAM_BASE && addr <= DRAM_END:
		return b.dram.Load(addr, size)
	case addr >= UART_BASE && addr <= UART_END:
		return b.uart.Load(addr, size)
	}
	return 0, NewException(LoadAccessFault, addr)
}

func (b *Bus) Store(addr, size, value uint64) *Exception {
	switch {
	case addr >= CLINT_BASE && addr <= CLINT_END:
		return b.clint.Store(addr, size, value)
	case addr >= PLIC_BASE && addr <= PLIC_END:
		return b.plic.Store(addr, size, value)
	case addr >= DRAM_BASE && addr <= DRAM_END:
		return b.dram.Store(addr, size, value)
	case addr >= UART_BASE && addr <= UART_END:
		return b.uart.Store(addr, size, value)
	}
	return NewException(StoreAMOAccessFault, addr)
}
