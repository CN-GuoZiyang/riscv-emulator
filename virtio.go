package main

type VirtioBlock struct {
	id             uint64
	driverFeatures uint32
	pageSize       uint32
	queueSel       uint32
	queueNum       uint32
	queuePfn       uint32
	queueNotify    uint32
	status         uint32
	disk           []uint8
}

const (
	MAX_BLOCK_QUEUE = 1
)

func NewVirtioBlock(diskImage []uint8) VirtioBlock {
	return VirtioBlock{
		queueNotify: MAX_BLOCK_QUEUE,
		disk:        diskImage,
	}
}

func (v *VirtioBlock) IsInterrupting() bool {
	if v.queueNotify < MAX_BLOCK_QUEUE {
		v.queueNotify = MAX_BLOCK_QUEUE
		return true
	}
	return false
}

func (v *VirtioBlock) Load(addr, size uint64) (uint64, *Exception) {
	if size != 32 {
		return 0, NewException(LoadAccessFault, addr)
	}
	switch addr {
	case VIRTIO_MAGIC:
		return 0x74726976, nil
	case VIRTIO_VERSION:
		return 0x1, nil
	case VIRTIO_DEVICE_ID:
		return 0x2, nil
	case VIRTIO_VENDOR_ID:
		return 0x554d4551, nil
	case VIRTIO_DEVICE_FEATURES:
		return 0, nil
	case VIRTIO_DRIVER_FEATURES:
		return uint64(v.driverFeatures), nil
	case VIRTIO_QUEUE_NUM_MAX:
		return 8, nil
	case VIRTIO_QUEUE_PFN:
		return uint64(v.queuePfn), nil
	case VIRTIO_STATUS:
		return uint64(v.status), nil
	default:
		return 0, nil
	}
}

func (v *VirtioBlock) Store(addr, size, value uint64) *Exception {
	if size != 32 {
		return NewException(StoreAMOAccessFault, addr)
	}
	switch addr {
	case VIRTIO_DEVICE_FEATURES:
		v.driverFeatures = uint32(value)
	case VIRTIO_GUEST_PAGE_SIZE:
		v.pageSize = uint32(value)
	case VIRTIO_QUEUE_SEL:
		v.queueSel = uint32(value)
	case VIRTIO_QUEUE_NUM:
		v.queueNum = uint32(value)
	case VIRTIO_QUEUE_PFN:
		v.queuePfn = uint32(value)
	case VIRTIO_QUEUE_NOTIFY:
		v.queueNotify = uint32(value)
	case VIRTIO_STATUS:
		v.status = uint32(value)
	}
	return nil
}

func (v *VirtioBlock) GetNewID() uint64 {
	v.id += 1
	return v.id
}

func (v *VirtioBlock) DescAddr() uint64 {
	return uint64(v.queuePfn) * uint64(v.pageSize)
}

func (v *VirtioBlock) ReadDisk(addr uint64) uint64 {
	return uint64(v.disk[addr])
}

func (v *VirtioBlock) WriteDisk(addr, value uint64) {
	v.disk[addr] = uint8(value)
}
