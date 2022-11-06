package main

type CSR struct {
	csrs [CSRS_NUM]uint64
}

func NewCSR() CSR {
	return CSR{
		csrs: [CSRS_NUM]uint64{},
	}
}

func (c *CSR) Load(addr uint64) uint64 {
	switch addr {
	case SIE:
		return c.csrs[MIE] & c.csrs[MIDELEG]
	case SIP:
		return c.csrs[MIP] & c.csrs[MIDELEG]
	case SSTATUS:
		return c.csrs[MSTATUS] & MASK_SSTATUS
	default:
		return c.csrs[addr]
	}
}

func (c *CSR) Store(addr, value uint64) {
	switch addr {
	case SIE:
		c.csrs[MIE] = (c.csrs[MIE] & ^c.csrs[MIDELEG]) | (value & c.csrs[MIDELEG])
	case SIP:
		c.csrs[MIP] = (c.csrs[MIP] & ^c.csrs[MIDELEG]) | (value & c.csrs[MIDELEG])
	case SSTATUS:
		c.csrs[MSTATUS] = (c.csrs[MSTATUS] & ^uint64(MASK_SSTATUS)) | (value & MASK_SSTATUS)
	default:
		c.csrs[addr] = value
	}
}
