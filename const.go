package main

const (
	DRAM_SIZE = 1024 * 1024 * 1024
	DRAM_BASE = 0x80000000
	DRAM_END  = DRAM_BASE + DRAM_SIZE - 1

	CSRS_NUM = 4096
)

// CLINT
const (
	CLINT_BASE = 0x200_0000
	CLINT_SIZE = 0x10000
	CLINT_END  = CLINT_BASE + CLINT_SIZE - 1

	CLINT_MTIMECMP = CLINT_BASE + 0x4000
	CLINT_MTIME    = CLINT_BASE + 0xbff8
)

// PLIC
const (
	PLIC_BASE = 0xc000000
	PLIC_SIZE = 0x4000000
	PLIC_END  = PLIC_BASE + PLIC_SIZE - 1

	PLIC_PENDING   = PLIC_BASE + 0x1000
	PLIC_SENABLE   = PLIC_BASE + 0x2000
	PLIC_SPRIORITY = PLIC_BASE + 0x201000
	PLIC_SCLAIM    = PLIC_BASE + 0x201004
)

// UART
const (
	UART_BASE = 0x1000_0000
	UART_SIZE = 0x100
	UART_END  = UART_BASE + UART_SIZE - 1
	// uart interrupt request
	UART_IRQ = 10
	// Receive holding register (for input bytes).
	UART_RHR = 0
	// Transmit holding register (for output bytes).
	UART_THR = 0
	// Line control register.
	UART_LCR = 3
	// Line status register.
	// LSR BIT 0:
	//
	//	0 = no data in receive holding register or FIFO.
	//	1 = data has been receive and saved in the receive holding register or FIFO.
	//
	// LSR BIT 5:
	//
	//	0 = transmit holding register is full. 16550 will not accept any data for transmission.
	//	1 = transmitter hold register (or FIFO) is empty. CPU can load the next character.
	UART_LSR = 5
	// The receiver (RX) bit MASK.
	MASK_UART_LSR_RX = 1
	// The transmitter (TX) bit MASK.
	MASK_UART_LSR_TX = 1 << 5
)

// CSR and MASK
const (
	MHARTID = 0xf14
	/// Machine status register.
	MSTATUS = 0x300
	/// Machine exception delefation register.
	MEDELEG = 0x302
	/// Machine interrupt delefation register.
	MIDELEG = 0x303
	/// Machine interrupt-enable register.
	MIE = 0x304
	/// Machine trap-handler base address.
	MTVEC = 0x305
	/// Machine counter enable.
	MCOUNTEREN = 0x306
	/// Scratch register for machine trap handlers.
	MSCRATCH = 0x340
	/// Machine exception program counter.
	MEPC = 0x341
	/// Machine trap cause.
	MCAUSE = 0x342
	/// Machine bad address or instruction.
	MTVAL = 0x343
	/// Machine interrupt pending.
	MIP = 0x344

	// Supervisor-level CSRs.
	/// Supervisor status register.
	SSTATUS = 0x100
	/// Supervisor interrupt-enable register.
	SIE = 0x104
	/// Supervisor trap handler base address.
	STVEC = 0x105
	/// Scratch register for supervisor trap handlers.
	SSCRATCH = 0x140
	/// Supervisor exception program counter.
	SEPC = 0x141
	/// Supervisor trap cause.
	SCAUSE = 0x142
	/// Supervisor bad address or instruction.
	STVAL = 0x143
	/// Supervisor interrupt pending.
	SIP = 0x144
	/// Supervisor address translation and protection.
	SATP = 0x180

	// mstatus and sstatus field mask
	MASK_SIE     = 1 << 1
	MASK_MIE     = 1 << 3
	MASK_SPIE    = 1 << 5
	MASK_UBE     = 1 << 6
	MASK_MPIE    = 1 << 7
	MASK_SPP     = 1 << 8
	MASK_VS      = 0b11 << 9
	MASK_MPP     = 0b11 << 11
	MASK_FS      = 0b11 << 13
	MASK_XS      = 0b11 << 15
	MASK_MPRV    = 1 << 17
	MASK_SUM     = 1 << 18
	MASK_MXR     = 1 << 19
	MASK_TVM     = 1 << 20
	MASK_TW      = 1 << 21
	MASK_TSR     = 1 << 22
	MASK_UXL     = 0b11 << 32
	MASK_SXL     = 0b11 << 34
	MASK_SBE     = 1 << 36
	MASK_MBE     = 1 << 37
	MASK_SD      = 1 << 63
	MASK_SSTATUS = MASK_SIE | MASK_SPIE | MASK_UBE | MASK_SPP | MASK_FS |
		MASK_XS | MASK_SUM | MASK_MXR | MASK_UXL | MASK_SD

	// MIP / SIP field mask
	MASK_SSIP = 1 << 1
	MASK_MSIP = 1 << 3
	MASK_STIP = 1 << 5
	MASK_MTIP = 1 << 7
	MASK_SEIP = 1 << 9
	MASK_MEIP = 1 << 11

	MASK_PPN = (1 << 44) - 1
)

// virtio
const (
	// The address which virtio starts.
	VIRTIO_BASE = 0x1000_1000
	// The size of virtio.
	VIRTIO_SIZE = 0x1000
	// The interrupt request of virtio.
	VIRTIO_END = VIRTIO_BASE + VIRTIO_SIZE - 1
	VIRTIO_IRQ = 1

	// The number of virtio descriptors. It must be a power of two.
	DESC_NUM = 8

	// Always return 0x74726976.
	VIRTIO_MAGIC = VIRTIO_BASE + 0x000
	// The version. 1 is legacy.
	VIRTIO_VERSION = VIRTIO_BASE + 0x004
	// device type 1 is net, 2 is disk.
	VIRTIO_DEVICE_ID = VIRTIO_BASE + 0x008
	// Always return 0x554d4551
	VIRTIO_VENDOR_ID = VIRTIO_BASE + 0x00c
	// Device features.
	VIRTIO_DEVICE_FEATURES = VIRTIO_BASE + 0x010
	// Driver features.
	VIRTIO_DRIVER_FEATURES = VIRTIO_BASE + 0x020
	// Page size for PFN, write-only.
	VIRTIO_GUEST_PAGE_SIZE = VIRTIO_BASE + 0x028
	// Select queue, write-only.
	VIRTIO_QUEUE_SEL = VIRTIO_BASE + 0x030
	// Max size of current queue, read-only. In QEMU, `VIRTIO_COUNT = 8`.
	VIRTIO_QUEUE_NUM_MAX = VIRTIO_BASE + 0x034
	// Size of current queue, write-only.
	VIRTIO_QUEUE_NUM = VIRTIO_BASE + 0x038
	// Physical page number for queue, read and write.
	VIRTIO_QUEUE_PFN = VIRTIO_BASE + 0x040
	// Notify the queue number, write-only.
	VIRTIO_QUEUE_NOTIFY = VIRTIO_BASE + 0x050
	// Device status, read and write. Reading from this register returns the current device status flags.
	// Writing non-zero values to this register sets the status flags, indicating the OS/driver
	// progress. Writing zero (0x0) to this register triggers a device reset.
	VIRTIO_STATUS = VIRTIO_BASE + 0x070

	PAGE_SIZE   = 4096
	SECTOR_SIZE = 512

	// virtio block request type
	VIRTIO_BLK_T_IN  = 0
	VIRTIO_BLK_T_OUT = 1

	// virtqueue descriptor flags
	VIRTQ_DESC_F_NEXT     = 1
	VIRTQ_DESC_F_WRITE    = 2
	VIRTQ_DESC_F_INDIRECT = 4
)
