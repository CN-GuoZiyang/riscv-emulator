package main

type Clint struct {
	mtime    uint64
	mtimecmp uint64
}

func NewClint() Clint {
	return Clint{}
}

func (c *Clint) Load(addr, size uint64) (uint64, *Exception) {
	if size != 64 {
		return 0, NewException(LoadAccessFault, addr)
	}
	switch addr {
	case CLINT_MTIME:
		return c.mtime, nil
	case CLINT_MTIMECMP:
		return c.mtimecmp, nil
	default:
		return 0, NewException(LoadAccessFault, addr)
	}
}

func (c *Clint) Store(addr, size, value uint64) *Exception {
	if size != 64 {
		return NewException(StoreAMOAccessFault, addr)
	}
	switch addr {
	case CLINT_MTIME:
		c.mtime = value
		return nil
	case CLINT_MTIMECMP:
		c.mtimecmp = value
		return nil
	default:
		return NewException(StoreAMOAccessFault, addr)
	}
}
