package main

type Bus struct {
	dram Dram
}

func NewBus(code []uint8) Bus {
	return Bus{
		dram: NewDram(code),
	}
}

func (b *Bus) Load(addr, size uint64) (uint64, *Exception) {
	switch {
	case addr >= DRAM_BASE && addr <= DRAM_END:
		return b.dram.Load(addr, size)
	}
	return 0, NewException(LoadAccessFault, addr)
}

func (b *Bus) Store(addr, size, value uint64) *Exception {
	switch {
	case addr >= DRAM_BASE && addr <= DRAM_END:
		return b.dram.Store(addr, size, value)
	}
	return NewException(StoreAMOAccessFault, addr)
}
