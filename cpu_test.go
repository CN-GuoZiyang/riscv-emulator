package main

import (
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func generateObj(assemblyFile string) {
	// cc := "riscv64-linux-gnu-gcc"
	cc := "clang"
	pieces := strings.Split(assemblyFile, ".")
	cmd := exec.Command(cc,
		"-c", "-Wl,-Ttext=0x0", "-nostdlib", "--target=riscv64-linux-gnu", "-march=rv64g", "-mabi=lp64", "-mno-relax",
		"-o", pieces[0], assemblyFile)
	err := cmd.Run()
	if err != nil {
		panic(err)
	}
}

func generateBinary(objFile string) {
	objcopy := "riscv64-linux-gnu-objcopy"
	cmd := exec.Command(objcopy,
		"-O", "binary", objFile, objFile+".bin")
	err := cmd.Run()
	if err != nil {
		panic(err)
	}
}

func testHelper(code, testname string, n int) (*Cpu, error) {
	s, err := os.Stat("tmp")
	if err != nil {
		if !os.IsExist(err) {
			err = os.Mkdir("tmp", os.ModePerm)
			if err != nil {
				panic("create tmp dir error")
			}
		}
	} else {
		if !s.IsDir() {
			panic("tmp is not a directory")
		}
	}
	testname = "tmp/" + testname
	fileName := testname + ".s"
	if !strings.HasSuffix(code, "\n") {
		code += "\n"
	}
	err = os.WriteFile(fileName, []byte(code), 0666)
	if err != nil {
		return nil, err
	}
	generateObj(fileName)
	generateBinary(testname)
	binaryCode, err := os.ReadFile(testname + ".bin")
	if err != nil {
		panic("read file error!")
	}
	cpu := NewCPU(binaryCode)
	for i := 0; i < n; i++ {
		inst, exception := cpu.Fetch()
		if exception != nil {
			// println("fetch inst exception occur!, " + exception.ToString())
			break
		}
		newPC, exception := cpu.Execute(inst)
		if exception != nil {
			// println("exception occur!, " + exception.ToString())
			break
		}
		cpu.Pc = newPC
	}
	return cpu, nil
}

type TestExp struct {
	RegName string
	Expect  uint64
}

func riscvTest(t *testing.T, code, name string, n int, exps []TestExp) {
	cpu, err := testHelper(code, name, n)
	assert.Nil(t, err)
	for _, exp := range exps {
		assert.Equal(t, exp.Expect, cpu.Reg(exp.RegName))
	}
}

func TestAddi(t *testing.T) {
	code := "addi x31, x0, 42"
	riscvTest(t, code, "test_addi", 1, []TestExp{
		{RegName: "x31", Expect: 42},
	})
}

func TestSimple(t *testing.T) {
	code := `addi	sp,sp,-16
sd	s0,8(sp)
addi	s0,sp,16
li	a5,42
mv	a0,a5
ld	s0,8(sp)
addi	sp,sp,16
jr	ra`
	riscvTest(t, code, "test_simple", 20, []TestExp{
		{RegName: "a0", Expect: 42},
	})
}

func TestLui(t *testing.T) {
	code := "lui a0, 42"
	riscvTest(t, code, "test_lui", 1, []TestExp{
		{RegName: "a0", Expect: 42 << 12},
	})
}

func TestAuipc(t *testing.T) {
	code := "auipc a0, 42"
	riscvTest(t, code, "test_auipc", 1, []TestExp{
		{RegName: "a0", Expect: DRAM_BASE + (42 << 12)},
	})
}

func TestJal(t *testing.T) {
	code := "jal a0, 42"
	riscvTest(t, code, "test_jal", 1, []TestExp{
		{RegName: "a0", Expect: DRAM_BASE + 4},
		{RegName: "pc", Expect: DRAM_BASE + 42},
	})
}

func TestJalr(t *testing.T) {
	code := `addi a1, zero, 42
jalr a0, -8(a1)`
	riscvTest(t, code, "test_jalr", 2, []TestExp{
		{RegName: "a0", Expect: DRAM_BASE + 8},
		{RegName: "pc", Expect: 34},
	})
}

func TestBeq(t *testing.T) {
	code := "beq  x0, x0, 42"
	riscvTest(t, code, "test_beq", 1, []TestExp{
		{RegName: "pc", Expect: DRAM_BASE + 42},
	})
}

func TestBne(t *testing.T) {
	code := `addi x1, x0, 10
bne  x0, x1, 42`
	riscvTest(t, code, "test_bne", 5, []TestExp{
		{RegName: "pc", Expect: DRAM_BASE + 42 + 4},
	})
}

func TestBlt(t *testing.T) {
	code := `addi x1, x0, 10
addi x2, x0, 20
blt  x1, x2, 42`
	riscvTest(t, code, "test_blt", 10, []TestExp{
		{RegName: "pc", Expect: DRAM_BASE + 42 + 8},
	})
}

func TestBge(t *testing.T) {
	code := `addi x1, x0, 10
addi x2, x0, 20
bge  x2, x1, 42`
	riscvTest(t, code, "test_bge", 10, []TestExp{
		{RegName: "pc", Expect: DRAM_BASE + 42 + 8},
	})
}

func TestBltu(t *testing.T) {
	code := `addi x1, x0, 10
addi x2, x0, 20
bltu x1, x2, 42`
	riscvTest(t, code, "test_bltu", 10, []TestExp{
		{RegName: "pc", Expect: DRAM_BASE + 42 + 8},
	})
}

func TestBgeu(t *testing.T) {
	code := `addi x1, x0, 10
addi x2, x0, 20
bgeu x2, x1, 42`
	riscvTest(t, code, "test_bgeu", 10, []TestExp{
		{RegName: "pc", Expect: DRAM_BASE + 42 + 8},
	})
}

func TestStoreLoad1(t *testing.T) {
	code := `addi s0, zero, 256
addi sp, sp, -16
sd   s0, 8(sp)
lb   t1, 8(sp)
lh   t2, 8(sp)`
	riscvTest(t, code, "test_store_load1", 10, []TestExp{
		{RegName: "t1", Expect: 0},
		{RegName: "t2", Expect: 256},
	})
}

func TestSlt(t *testing.T) {
	code := `addi t0, zero, 14
addi t1, zero, 24
slt  t2, t0, t1
slti t3, t0, 42
sltiu t4, t0, 84`
	riscvTest(t, code, "test_slt", 7, []TestExp{
		{RegName: "t2", Expect: 1},
		{RegName: "t3", Expect: 1},
		{RegName: "t4", Expect: 1},
	})
}

func TestXor(t *testing.T) {
	code := `addi a0, zero, 0b10
xori a1, a0, 0b01
xor a2, a1, a1`
	riscvTest(t, code, "test_xor", 5, []TestExp{
		{RegName: "a1", Expect: 3},
		{RegName: "a2", Expect: 0},
	})
}

func TestOr(t *testing.T) {
	code := `addi a0, zero, 0b10
ori  a1, a0, 0b01
or   a2, a0, a0`
	riscvTest(t, code, "test_or", 3, []TestExp{
		{RegName: "a1", Expect: 0b11},
		{RegName: "a2", Expect: 0b10},
	})
}

func TestAnd(t *testing.T) {
	code := `addi a0, zero, 0b10 
andi a1, a0, 0b11
and  a2, a0, a1`
	riscvTest(t, code, "test_and", 3, []TestExp{
		{RegName: "a1", Expect: 0b10},
		{RegName: "a2", Expect: 0b10},
	})
}

func TestSll(t *testing.T) {
	code := `addi a0, zero, 1
addi a1, zero, 5
sll  a2, a0, a1
slli a3, a0, 5
addi s0, zero, 64
sll  a4, a0, s0`
	riscvTest(t, code, "test_sll", 10, []TestExp{
		{RegName: "a2", Expect: 1 << 5},
		{RegName: "a3", Expect: 1 << 5},
		{RegName: "a4", Expect: 1},
	})
}

func TestSraSrl(t *testing.T) {
	code := `addi a0, zero, -8
addi a1, zero, 1
sra  a2, a0, a1
srai a3, a0, 2
srli a4, a0, 2
srl  a5, a0, a1`
	a := int64(-4)
	b := int64(-2)
	c := int64(-8)
	riscvTest(t, code, "test_sra_srl", 10, []TestExp{
		{RegName: "a2", Expect: uint64(a)},
		{RegName: "a3", Expect: uint64(b)},
		{RegName: "a4", Expect: uint64(c) >> 2},
		{RegName: "a5", Expect: uint64(c) >> 1},
	})
}

func TestWordOp(t *testing.T) {
	code := `addi a0, zero, 42 
lui  a1, 0x7f000
addw a2, a0, a1`
	riscvTest(t, code, "test_word_op", 29, []TestExp{
		{RegName: "a2", Expect: 0x7f00002a},
	})
}
