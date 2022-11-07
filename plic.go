package main

type Plic struct {
	pending   uint64
	senable   uint64
	spriority uint64
	sclaim    uint64
}

func NewPlic() Plic {
	return Plic{}
}

func (p *Plic) Load(addr, size uint64) (uint64, *Exception) {
	if size != 32 {
		return 0, NewException(LoadAccessFault, addr)
	}
	switch addr {
	case PLIC_PENDING:
		return p.pending, nil
	case PLIC_SENABLE:
		return p.senable, nil
	case PLIC_SPRIORITY:
		return p.spriority, nil
	case PLIC_SCLAIM:
		return p.sclaim, nil
	}
	return 0, nil
}

func (p *Plic) Store(addr, size, value uint64) *Exception {
	if size != 32 {
		return NewException(StoreAMOAccessFault, addr)
	}
	switch addr {
	case PLIC_PENDING:
		p.pending = value
	case PLIC_SENABLE:
		p.senable = value
	case PLIC_SPRIORITY:
		p.spriority = value
	case PLIC_SCLAIM:
		p.sclaim = value
	}
	return nil
}
