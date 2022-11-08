package main

type VirtqDesc struct {
	addr   uint64
	length uint32
	flags  uint16
	next   uint16
}

type VirtqAvail struct {
	flags     uint64
	idx       uint16
	ring      [DESC_NUM]uint16
	usedEvent uint16
}

type VirtqUsedElem struct {
	id     uint32
	length uint32
}

type VirtqUsed struct {
	flags      uint16
	idx        uint16
	ring       [DESC_NUM]VirtqUsedElem
	availEvent uint16
}

type VirtioBlkRequest struct {
	iotype   uint32
	reserved uint32
	sector   uint64
}
