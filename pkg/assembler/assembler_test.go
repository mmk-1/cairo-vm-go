package assembler

import (
	"fmt"
	"testing"

	f "github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
	"github.com/stretchr/testify/assert"
)

func TestAssertEqRegisterToInstrList(t *testing.T) {
	instrList := parseSingleInstructionToInstrList("[ap] = [fp], ap++;")

	expected := Instruction{
		OffDest:     0,
		OffOp0:      -1,
		OffOp1:      0,
		DstRegister: 0,
		Op0Register: 1,
		Op1Source:   2,
		Res:         0,
		PcUpdate:    0,
		ApUpdate:    2,
		Opcode:      4,
	}
	assert.Equal(t, expected, instrList[0])
}

func TestCallRelToInstrList(t *testing.T) {
	instrList := parseSingleInstructionToInstrList("call rel 123;")

	expected := Instruction{
		OffDest:     0,
		OffOp0:      1,
		OffOp1:      1,
		DstRegister: 0,
		Op0Register: 0,
		Op1Source:   1,
		Res:         0,
		PcUpdate:    2,
		ApUpdate:    0,
		Opcode:      1,
		Imm:         "123",
	}
	assert.Equal(t, expected, instrList[0])
}

func TestCallAbsToInstrList(t *testing.T) {
	instrList := parseSingleInstructionToInstrList("call abs [fp + 4];")

	expected := Instruction{
		OffDest:     0,
		OffOp0:      1,
		OffOp1:      4,
		DstRegister: 0,
		Op0Register: 0,
		Op1Source:   2,
		Res:         0,
		PcUpdate:    1,
		ApUpdate:    0,
		Opcode:      1,
	}
	assert.Equal(t, expected, instrList[0])
}

func TestRetToInstrList(t *testing.T) {
	instrList := parseSingleInstructionToInstrList("ret;")

	expected := Instruction{
		OffDest:     -2,
		OffOp0:      -1,
		OffOp1:      -1,
		DstRegister: 1,
		Op0Register: 1,
		Op1Source:   2,
		Res:         0,
		PcUpdate:    1,
		ApUpdate:    0,
		Opcode:      2,
	}
	assert.Equal(t, expected, instrList[0])
}

func TestJmpAbsToInstrList(t *testing.T) {
	instrList := parseSingleInstructionToInstrList("jmp abs 123, ap++;")
	// Raw code below gives parsing error! (code taken from whitepaper)
	// instrList := parseSingleInstructionToInstrList("jmp rel [ap + 1] + [fp - 7];")
	// expected := Instruction{
	// 	OffDest:     -1,
	// 	OffOp0:      1,
	// 	OffOp1:      -7,
	// 	DstRegister: 1,
	// 	Op0Register: 0,
	// 	Op1Source:   2,
	// 	Res:         1,
	// 	PcUpdate:    2,
	// 	ApUpdate:    0,
	// 	Opcode:      0,
	// }
	expected := Instruction{
		OffDest:     -1,
		OffOp0:      -1,
		OffOp1:      1,
		Imm:         "123",
		DstRegister: 1,
		Op0Register: 1,
		Op1Source:   1,
		Res:         0,
		PcUpdate:    1,
		ApUpdate:    2,
		Opcode:      0,
	}
	assert.Equal(t, expected, instrList[0])
}

func TestAssertEqRegister(t *testing.T) {
	encode := parseSingleInstruction("[ap] = [fp + 0], ap++;")

	// verify offsets
	dstOffset := uint16(encode)
	assert.Equal(t, biased(0), dstOffset)

	op0Offset := uint16(encode >> 16)
	assert.Equal(t, biased(-1), op0Offset)

	op1Offset := uint16(encode >> 32)
	assert.Equal(t, biased(0), op1Offset)

	// verify flags
	flagsReg := uint16(encode >> flagsOffset)
	assert.True(t, (flagsReg>>dstRegBit)&1 == 0)
	assert.True(t, (flagsReg>>op0RegBit)&1 == 1)
	assert.True(
		t,
		(flagsReg>>op1ImmBit)&1 == 0 &&
			(flagsReg>>op1FpBit)&1 == 1 &&
			(flagsReg>>op1ApBit)&1 == 0,
	)
	assert.True(
		t, (flagsReg>>resAddBit)&1 == 0 && (flagsReg>>resMulBit)&1 == 0,
	)
	assert.True(
		t,
		(flagsReg>>pcJumpAbsBit)&1 == 0 &&
			(flagsReg>>pcJumpRelBit)&1 == 0 &&
			(flagsReg>>pcJnzBit)&1 == 0,
	)
	assert.True(
		t, (flagsReg>>apAddBit)&1 == 0 && (flagsReg>>apAdd1Bit)&1 == 1,
	)
	assert.True(
		t,
		(flagsReg>>opcodeRetBit)&1 == 0 &&
			(flagsReg>>opcodeCallBit)&1 == 0 &&
			(flagsReg>>opcodeAssertEqBit)&1 == 1,
	)
}

func TestAssertEqImm(t *testing.T) {
	encode, imm := parseImmediateInstruction("[fp + 1] = 5;")

	// verify imm
	assert.Equal(t, uint64(5), imm.Uint64())

	// verify offsets
	dstOffset := uint16(encode)
	assert.Equal(t, biased(1), dstOffset)

	op0Offset := uint16(encode >> 16)
	assert.Equal(t, biased(-1), op0Offset)

	op1Offset := uint16(encode >> 32)
	assert.Equal(t, biased(1), op1Offset)

	// verify flags
	flagsReg := uint16(encode >> flagsOffset)
	assert.True(t, (flagsReg>>dstRegBit)&1 == 1)
	assert.True(t, (flagsReg>>op0RegBit)&1 == 1)
	assert.True(
		t,
		(flagsReg>>op1ImmBit)&1 == 1 &&
			(flagsReg>>op1FpBit)&1 == 0 &&
			(flagsReg>>op1ApBit)&1 == 0,
	)
	assert.True(
		t, (flagsReg>>resAddBit)&1 == 0 && (flagsReg>>resMulBit)&1 == 0,
	)
	assert.True(
		t,
		(flagsReg>>pcJumpAbsBit)&1 == 0 &&
			(flagsReg>>pcJumpRelBit)&1 == 0 &&
			(flagsReg>>pcJnzBit)&1 == 0,
	)
	assert.True(
		t, (flagsReg>>apAddBit)&1 == 0 && (flagsReg>>apAdd1Bit)&1 == 0,
	)
	assert.True(
		t,
		(flagsReg>>opcodeRetBit)&1 == 0 &&
			(flagsReg>>opcodeCallBit)&1 == 0 &&
			(flagsReg>>opcodeAssertEqBit)&1 == 1,
	)

}

func TestAssertEqDoubleDeref(t *testing.T) {
	encode := parseSingleInstruction("[ap + 1] = [[ap - 2] - 3];")

	// verify offsets
	dstOffset := uint16(encode)
	assert.Equal(t, biased(1), dstOffset)

	op0Offset := uint16(encode >> 16)
	assert.Equal(t, biased(-2), op0Offset)

	op1Offset := uint16(encode >> 32)
	assert.Equal(t, biased(-3), op1Offset)

	// verify flags
	flagsReg := uint16(encode >> flagsOffset)
	assert.True(t, (flagsReg>>dstRegBit)&1 == 0)
	assert.True(t, (flagsReg>>op0RegBit)&1 == 0)
	assert.True(
		t,
		(flagsReg>>op1ImmBit)&1 == 0 &&
			(flagsReg>>op1FpBit)&1 == 0 &&
			(flagsReg>>op1ApBit)&1 == 0,
	)
	assert.True(
		t, (flagsReg>>resAddBit)&1 == 0 && (flagsReg>>resMulBit)&1 == 0,
	)
	assert.True(
		t,
		(flagsReg>>pcJumpAbsBit)&1 == 0 &&
			(flagsReg>>pcJumpRelBit)&1 == 0 &&
			(flagsReg>>pcJnzBit)&1 == 0,
	)
	assert.True(
		t, (flagsReg>>apAddBit)&1 == 0 && (flagsReg>>apAdd1Bit)&1 == 0,
	)
	assert.True(
		t,
		(flagsReg>>opcodeRetBit)&1 == 0 &&
			(flagsReg>>opcodeCallBit)&1 == 0 &&
			(flagsReg>>opcodeAssertEqBit)&1 == 1,
	)
}

func TestAssertEqMathOperation(t *testing.T) {
	encode := parseSingleInstruction("[fp - 10] = [ap + 2] * [ap - 3];")

	// verify offsets
	dstOffset := uint16(encode)
	assert.Equal(t, biased(-10), dstOffset)

	op0Offset := uint16(encode >> 16)
	assert.Equal(t, biased(2), op0Offset)

	op1Offset := uint16(encode >> 32)
	assert.Equal(t, biased(-3), op1Offset)

	// verify flags
	flagsReg := uint16(encode >> flagsOffset)
	assert.True(t, (flagsReg>>dstRegBit)&1 == 1)
	assert.True(t, (flagsReg>>op0RegBit)&1 == 0)
	assert.True(
		t,
		(flagsReg>>op1ImmBit)&1 == 0 &&
			(flagsReg>>op1FpBit)&1 == 0 &&
			(flagsReg>>op1ApBit)&1 == 1,
	)
	assert.True(
		t, (flagsReg>>resAddBit)&1 == 0 && (flagsReg>>resMulBit)&1 == 1,
	)
	assert.True(
		t,
		(flagsReg>>pcJumpAbsBit)&1 == 0 &&
			(flagsReg>>pcJumpRelBit)&1 == 0 &&
			(flagsReg>>pcJnzBit)&1 == 0,
	)
	assert.True(
		t, (flagsReg>>apAddBit)&1 == 0 && (flagsReg>>apAdd1Bit)&1 == 0,
	)
	assert.True(
		t,
		(flagsReg>>opcodeRetBit)&1 == 0 &&
			(flagsReg>>opcodeCallBit)&1 == 0 &&
			(flagsReg>>opcodeAssertEqBit)&1 == 1,
	)
}

func TestCallAbs(t *testing.T) {
	encode, imm := parseImmediateInstruction("call abs 123;")

	// verify imm
	assert.Equal(t, uint64(123), imm.Uint64())

	// verify offsets
	dstOffset := uint16(encode)
	assert.Equal(t, biased(0), dstOffset)

	op0Offset := uint16(encode >> 16)
	assert.Equal(t, biased(1), op0Offset)

	op1Offset := uint16(encode >> 32)
	assert.Equal(t, biased(1), op1Offset)

	// verify flags
	flagsReg := uint16(encode >> flagsOffset)
	assert.True(t, (flagsReg>>dstRegBit)&1 == 0)
	assert.True(t, (flagsReg>>op0RegBit)&1 == 0)
	assert.True(
		t,
		(flagsReg>>op1ImmBit)&1 == 1 &&
			(flagsReg>>op1FpBit)&1 == 0 &&
			(flagsReg>>op1ApBit)&1 == 0,
	)
	assert.True(
		t, (flagsReg>>resAddBit)&1 == 0 && (flagsReg>>resMulBit)&1 == 0,
	)
	assert.True(
		t,
		(flagsReg>>pcJumpAbsBit)&1 == 1 &&
			(flagsReg>>pcJumpRelBit)&1 == 0 &&
			(flagsReg>>pcJnzBit)&1 == 0,
	)
	assert.True(
		t, (flagsReg>>apAddBit)&1 == 0 && (flagsReg>>apAdd1Bit)&1 == 0,
	)
	assert.True(
		t,
		(flagsReg>>opcodeRetBit)&1 == 0 &&
			(flagsReg>>opcodeCallBit)&1 == 1 &&
			(flagsReg>>opcodeAssertEqBit)&1 == 0,
	)
}

func TestCallRel(t *testing.T) {
	encode := parseSingleInstruction("call rel [ap - 3];")

	// verify offsets
	dstOffset := uint16(encode)
	assert.Equal(t, biased(0), dstOffset)

	op0Offset := uint16(encode >> 16)
	assert.Equal(t, biased(1), op0Offset)

	op1Offset := uint16(encode >> 32)
	assert.Equal(t, biased(-3), op1Offset)

	// verify flags
	flagsReg := uint16(encode >> flagsOffset)
	assert.True(t, (flagsReg>>dstRegBit)&1 == 0)
	assert.True(t, (flagsReg>>op0RegBit)&1 == 0)
	assert.True(
		t,
		(flagsReg>>op1ImmBit)&1 == 0 &&
			(flagsReg>>op1FpBit)&1 == 0 &&
			(flagsReg>>op1ApBit)&1 == 1,
	)
	assert.True(
		t, (flagsReg>>resAddBit)&1 == 0 && (flagsReg>>resMulBit)&1 == 0,
	)
	assert.True(
		t,
		(flagsReg>>pcJumpAbsBit)&1 == 0 &&
			(flagsReg>>pcJumpRelBit)&1 == 1 &&
			(flagsReg>>pcJnzBit)&1 == 0,
	)
	assert.True(
		t, (flagsReg>>apAddBit)&1 == 0 && (flagsReg>>apAdd1Bit)&1 == 0,
	)
	assert.True(
		t,
		(flagsReg>>opcodeRetBit)&1 == 0 &&
			(flagsReg>>opcodeCallBit)&1 == 1 &&
			(flagsReg>>opcodeAssertEqBit)&1 == 0,
	)
}

func TestRet(t *testing.T) {
	encode := parseSingleInstruction("ret;")

	// verify offsets
	dstOffset := uint16(encode)
	assert.Equal(t, biased(-2), dstOffset)

	op0Offset := uint16(encode >> 16)
	assert.Equal(t, biased(-1), op0Offset)

	op1Offset := uint16(encode >> 32)
	assert.Equal(t, biased(-1), op1Offset)

	// verify flags
	flagsReg := uint16(encode >> flagsOffset)
	assert.True(t, (flagsReg>>dstRegBit)&1 == 1)
	assert.True(t, (flagsReg>>op0RegBit)&1 == 1)
	assert.True(
		t,
		(flagsReg>>op1ImmBit)&1 == 0 &&
			(flagsReg>>op1FpBit)&1 == 1 &&
			(flagsReg>>op1ApBit)&1 == 0,
	)
	assert.True(
		t, (flagsReg>>resAddBit)&1 == 0 && (flagsReg>>resMulBit)&1 == 0,
	)
	assert.True(
		t,
		(flagsReg>>pcJumpAbsBit)&1 == 1 &&
			(flagsReg>>pcJumpRelBit)&1 == 0 &&
			(flagsReg>>pcJnzBit)&1 == 0,
	)
	assert.True(
		t, (flagsReg>>apAddBit)&1 == 0 && (flagsReg>>apAdd1Bit)&1 == 0,
	)
	assert.True(
		t,
		(flagsReg>>opcodeRetBit)&1 == 1 &&
			(flagsReg>>opcodeCallBit)&1 == 0 &&
			(flagsReg>>opcodeAssertEqBit)&1 == 0,
	)
}

func TestJumpAbs(t *testing.T) {
	encode := parseSingleInstruction("jmp abs [fp - 5] + [fp + 3];")

	// verify offsets
	dstOffset := uint16(encode)
	assert.Equal(t, biased(-1), dstOffset)

	op0Offset := uint16(encode >> 16)
	assert.Equal(t, biased(-5), op0Offset)

	op1Offset := uint16(encode >> 32)
	assert.Equal(t, biased(3), op1Offset)

	// verify flags
	flagsReg := uint16(encode >> flagsOffset)
	assert.True(t, (flagsReg>>dstRegBit)&1 == 1)
	assert.True(t, (flagsReg>>op0RegBit)&1 == 1)
	assert.True(
		t,
		(flagsReg>>op1ImmBit)&1 == 0 &&
			(flagsReg>>op1FpBit)&1 == 1 &&
			(flagsReg>>op1ApBit)&1 == 0,
	)
	assert.True(
		t, (flagsReg>>resAddBit)&1 == 1 && (flagsReg>>resMulBit)&1 == 0,
	)
	assert.True(
		t,
		(flagsReg>>pcJumpAbsBit)&1 == 1 &&
			(flagsReg>>pcJumpRelBit)&1 == 0 &&
			(flagsReg>>pcJnzBit)&1 == 0,
	)
	assert.True(
		t, (flagsReg>>apAddBit)&1 == 0 && (flagsReg>>apAdd1Bit)&1 == 0,
	)
	assert.True(
		t,
		(flagsReg>>opcodeRetBit)&1 == 0 &&
			(flagsReg>>opcodeCallBit)&1 == 0 &&
			(flagsReg>>opcodeAssertEqBit)&1 == 0,
	)
}

func TestJnz(t *testing.T) {
	encode := parseSingleInstruction("jmp rel [ap - 2] if [fp - 7] != 0;")

	// verify offsets
	dstOffset := uint16(encode)
	assert.Equal(t, biased(-7), dstOffset)

	op0Offset := uint16(encode >> 16)
	assert.Equal(t, biased(-1), op0Offset)

	op1Offset := uint16(encode >> 32)
	assert.Equal(t, biased(-2), op1Offset)

	// verify flags
	flagsReg := uint16(encode >> flagsOffset)
	assert.True(t, (flagsReg>>dstRegBit)&1 == 1)
	assert.True(t, (flagsReg>>op0RegBit)&1 == 1)
	assert.True(
		t,
		(flagsReg>>op1ImmBit)&1 == 0 &&
			(flagsReg>>op1FpBit)&1 == 0 &&
			(flagsReg>>op1ApBit)&1 == 1,
	)
	assert.True(
		t, (flagsReg>>resAddBit)&1 == 0 && (flagsReg>>resMulBit)&1 == 0,
	)
	assert.True(
		t,
		(flagsReg>>pcJumpAbsBit)&1 == 0 &&
			(flagsReg>>pcJumpRelBit)&1 == 0 &&
			(flagsReg>>pcJnzBit)&1 == 1,
	)
	assert.True(
		t, (flagsReg>>apAddBit)&1 == 0 && (flagsReg>>apAdd1Bit)&1 == 0,
	)
	assert.True(
		t,
		(flagsReg>>opcodeRetBit)&1 == 0 &&
			(flagsReg>>opcodeCallBit)&1 == 0 &&
			(flagsReg>>opcodeAssertEqBit)&1 == 0,
	)
}

func TestAddApImm(t *testing.T) {
	encode, imm := parseImmediateInstruction("ap += 150;")

	// verify imm
	assert.Equal(t, uint64(150), imm.Uint64())

	// verify offsets
	dstOffset := uint16(encode)
	assert.Equal(t, biased(-1), dstOffset)

	op0Offset := uint16(encode >> 16)
	assert.Equal(t, biased(-1), op0Offset)

	op1Offset := uint16(encode >> 32)
	assert.Equal(t, biased(1), op1Offset)

	// verify flags
	flagsReg := uint16(encode >> flagsOffset)
	assert.True(t, (flagsReg>>dstRegBit)&1 == 1)
	assert.True(t, (flagsReg>>op0RegBit)&1 == 1)
	assert.True(
		t,
		(flagsReg>>op1ImmBit)&1 == 1 &&
			(flagsReg>>op1FpBit)&1 == 0 &&
			(flagsReg>>op1ApBit)&1 == 0,
	)
	assert.True(
		t, (flagsReg>>resAddBit)&1 == 0 && (flagsReg>>resMulBit)&1 == 0,
	)
	assert.True(
		t,
		(flagsReg>>pcJumpAbsBit)&1 == 0 &&
			(flagsReg>>pcJumpRelBit)&1 == 0 &&
			(flagsReg>>pcJnzBit)&1 == 0,
	)
	assert.True(
		t, (flagsReg>>apAddBit)&1 == 1 && (flagsReg>>apAdd1Bit)&1 == 0,
	)
	assert.True(
		t,
		(flagsReg>>opcodeRetBit)&1 == 0 &&
			(flagsReg>>opcodeCallBit)&1 == 0 &&
			(flagsReg>>opcodeAssertEqBit)&1 == 0,
	)

}

func parseImmediateInstruction(casmCode string) (uint64, *f.Element) {
	instructions, err := CasmToBytecode(casmCode)
	if err != nil {
		panic(err)
	}

	if len(instructions) != 2 {
		panic(fmt.Errorf("Expected a sized 2 instruction, got %d", len(instructions)))
	}

	return instructions[0].Uint64(), instructions[1]
}

func parseSingleInstruction(casmCode string) uint64 {
	instructions, err := CasmToBytecode(casmCode)
	if err != nil {
		panic(err)
	}

	if len(instructions) != 1 {
		panic(fmt.Errorf("Expected 1 instruction, got %d", len(instructions)))
	}
	return instructions[0].Uint64()
}

func parseSingleInstructionToInstrList(casmCode string) []Instruction {
	casmAst, err := parser.ParseString("", casmCode)
	if err != nil {
		panic(err)
	}
	instructions, err := astToInstruction(casmAst)
	if err != nil {
		panic(err)
	}

	// if len(instructions) != 1 {
	// 	panic(fmt.Errorf("Expected 1 instruction, got %d", len(instructions)))
	// }
	return instructions
}

func biased(num int16) uint16 {
	return uint16(num) ^ 0x8000
}
