package main

type Dram struct {
	dram []uint8
}

func NewDram(code []uint8) Dram {
	dram := make([]uint8, DRAM_SIZE)
	dram = append(code, dram...)[:DRAM_SIZE]
	return Dram{
		dram: dram,
	}
}

func (d *Dram) Load(addr, size uint64) (uint64, *Exception) {
	if _, ok := map[uint64]struct{}{
		8:  {},
		16: {},
		32: {},
		64: {}}[size]; !ok {
		return 0, NewException(LoadAccessFault, addr)
	}
	nbytes := size / 8
	index := addr - DRAM_BASE
	code := uint64(d.dram[index])
	for i := uint64(1); i < nbytes; i++ {
		code |= uint64(d.dram[index+i]) << (i * 8)
	}
	return code, nil
}

func (d *Dram) Store(addr, size, value uint64) *Exception {
	if _, ok := map[uint64]struct{}{
		8:  {},
		16: {},
		32: {},
		64: {}}[size]; !ok {
		return NewException(StoreAMOAccessFault, addr)
	}
	nbytes := size / 8
	index := addr - DRAM_BASE
	for i := uint64(0); i < nbytes; i++ {
		offset := 8 * i
		d.dram[index+i] = uint8((value >> offset) & 0xff)
	}
	return nil
}
