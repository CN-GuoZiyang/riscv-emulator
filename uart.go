package main

import (
	"fmt"
	"os"
	"os/exec"
	"sync"
	"sync/atomic"
)

type Uart struct {
	uart      *[UART_SIZE]uint8
	cond      *sync.Cond
	interrupt *atomic.Bool
}

func NewUart() Uart {
	array := [UART_SIZE]uint8{}
	array[UART_LSR] |= MASK_UART_LSR_TX

	cond := sync.NewCond(&sync.Mutex{})
	interrupt := &atomic.Bool{}

	// 关闭终端缓冲
	exec.Command("stty", "-F", "/dev/tty", "cbreak", "min", "1").Run()

	go func() {
		for {
			func() {
				bs := make([]byte, 1)
				_, err := os.Stdin.Read(bs)
				if err != nil {
					fmt.Println(err.Error())
					return
				}
				b := bs[0]
				cond.L.Lock()
				defer cond.L.Unlock()
				for array[UART_LSR]&MASK_UART_LSR_RX == 1 {
					cond.Wait()
				}
				array[UART_RHR] = b
				interrupt.Store(true)
				array[UART_LSR] |= MASK_UART_LSR_RX
			}()
		}
	}()

	return Uart{
		uart:      &array,
		cond:      cond,
		interrupt: interrupt,
	}
}

func (u *Uart) Load(addr, size uint64) (uint64, *Exception) {
	if size != 8 {
		return 0, NewException(LoadAccessFault, addr)
	}
	u.cond.L.Lock()
	defer u.cond.L.Unlock()
	array := u.uart
	index := addr - UART_BASE
	switch index {
	case UART_RHR:
		u.cond.Signal()
		array[UART_LSR] &= ^uint8(MASK_UART_LSR_RX)
		return uint64(array[UART_RHR]), nil
	default:
		return uint64(array[index]), nil
	}
}

func (u *Uart) Store(addr, size, value uint64) *Exception {
	if size != 8 {
		return NewException(StoreAMOAccessFault, addr)
	}
	u.cond.L.Lock()
	defer u.cond.L.Unlock()
	array := u.uart
	index := addr - UART_BASE
	switch index {
	case UART_THR:
		fmt.Print(string(rune(value)))
		return nil
	default:
		array[index] = uint8(value)
		return nil
	}
}

func (u *Uart) IsInterrupting() bool {
	return u.interrupt.Swap(false)
}
