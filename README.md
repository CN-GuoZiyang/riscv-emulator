riscv-emulator in go

how to run:

compile add-addi.s with following
```shell
$ riscv64-linux-gnu-gcc -c -Wl,-Ttext=0x0 -nostdlib -march=rv64g -mno-relax -o add-addi add-addi.s
$ riscv64-linux-gnu-objcopy -O binary add-addi add-addi.bin
```

run
```shell
$ go buid
$ ./riscv-emulator add-addi.bin
```

x31(t6) should be 0x2A