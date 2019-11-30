// Copyright 2012 The Freetype-Go Authors. All rights reserved.
// Use of this source code is governed by your choice of either the
// FreeType License or the GNU General Public License version 2 (or
// any later version), both of which can be found in the LICENSE file.

package truetype

// This file implements a Truetype bytecode interpreter.
// The opcodes are described at https://developer.apple.com/fonts/TTRefMan/RM05/Chap5.html

import (
	"errors"
)

// Hinter implements bytecode hinting. Pass a Hinter to GlyphBuf.Load to hint
// the resulting glyph. A Hinter can be re-used to hint a series of glyphs from
// a Font.
type Hinter struct {
	stack, store []int32

	// The fields below constitue the graphics state, which is described at
	// https://developer.apple.com/fonts/TTRefMan/RM04/Chap4.html

	// Projection vector, freedom vector and dual projection vector.
	pv, fv, dv [2]f2dot14
	// Minimum distance.
	minDist f26dot6
	// Loop count.
	loop int32
	// Rounding policy.
	roundPeriod, roundPhase, roundThreshold f26dot6
}

func (h *Hinter) init(f *Font) {
	if x := int(f.maxStackElements); x > len(h.stack) {
		x += 255
		x &^= 255
		h.stack = make([]int32, x)
	}
	if x := int(f.maxStorage); x > len(h.store) {
		x += 15
		x &^= 15
		h.store = make([]int32, x)
	}
}

func (h *Hinter) run(program []byte) error {
	// The default vectors are along the X axis.
	h.pv = [2]f2dot14{0x4000, 0}
	h.fv = [2]f2dot14{0x4000, 0}
	h.dv = [2]f2dot14{0x4000, 0}
	// The default minimum distance is 1.
	h.minDist = 1 << 6
	// The default loop count is 1.
	h.loop = 1
	// The default rounding policy is round to grid.
	h.roundPeriod = 1 << 6
	h.roundPhase = 0
	h.roundThreshold = 1 << 5

	if len(program) > 50000 {
		return errors.New("truetype: hinting: too many instructions")
	}
	var (
		steps, pc, top int
		opcode         uint8
	)
	for 0 <= pc && pc < len(program) {
		steps++
		if steps == 100000 {
			return errors.New("truetype: hinting: too many steps")
		}
		opcode = program[pc]
		if popCount[opcode] == q {
			return errors.New("truetype: hinting: unimplemented instruction")
		}
		if top < int(popCount[opcode]) {
			return errors.New("truetype: hinting: stack underflow")
		}
		switch opcode {

		case opSVTCA0:
			h.pv = [2]f2dot14{0, 0x4000}
			h.fv = [2]f2dot14{0, 0x4000}
			// TODO: h.dv = h.pv ??

		case opSVTCA1:
			h.pv = [2]f2dot14{0x4000, 0}
			h.fv = [2]f2dot14{0x4000, 0}
			// TODO: h.dv = h.pv ??

		case opSPVTCA0:
			h.pv = [2]f2dot14{0, 0x4000}
			// TODO: h.dv = h.pv ??

		case opSPVTCA1:
			h.pv = [2]f2dot14{0x4000, 0}
			// TODO: h.dv = h.pv ??

		case opSFVTCA0:
			h.fv = [2]f2dot14{0, 0x4000}

		case opSFVTCA1:
			h.fv = [2]f2dot14{0x4000, 0}

		case opSPVFS:
			top -= 2
			h.pv[0] = f2dot14(h.stack[top+0])
			h.pv[1] = f2dot14(h.stack[top+1])
			// TODO: normalize h.pv ??
			// TODO: h.dv = h.pv ??

		case opSFVFS:
			top -= 2
			h.fv[0] = f2dot14(h.stack[top+0])
			h.fv[1] = f2dot14(h.stack[top+1])
			// TODO: normalize h.fv ??

		case opGPV:
			if top+1 >= len(h.stack) {
				return errors.New("truetype: hinting: stack overflow")
			}
			h.stack[top+0] = int32(h.pv[0])
			h.stack[top+1] = int32(h.pv[1])
			top += 2

		case opGFV:
			if top+1 >= len(h.stack) {
				return errors.New("truetype: hinting: stack overflow")
			}
			h.stack[top+0] = int32(h.fv[0])
			h.stack[top+1] = int32(h.fv[1])
			top += 2

		case opSFVTPV:
			h.fv = h.pv

		case opSLOOP:
			top--
			if h.stack[top] <= 0 {
				return errors.New("truetype: hinting: invalid data")
			}
			h.loop = h.stack[top]

		case opRTG:
			h.roundPeriod = 1 << 6
			h.roundPhase = 0
			h.roundThreshold = 1 << 5

		case opRTHG:
			h.roundPeriod = 1 << 6
			h.roundPhase = 1 << 5
			h.roundThreshold = 1 << 5

		case opSMD:
			top--
			h.minDist = f26dot6(h.stack[top])

		case opELSE:
			opcode = 1
			goto ifelse

		case opJMPR:
			top--
			pc += int(h.stack[top])
			continue

		case opDUP:
			if top >= len(h.stack) {
				return errors.New("truetype: hinting: stack overflow")
			}
			h.stack[top] = h.stack[top-1]
			top++

		case opPOP:
			top--

		case opCLEAR:
			top = 0

		case opSWAP:
			h.stack[top-1], h.stack[top-2] = h.stack[top-2], h.stack[top-1]

		case opDEPTH:
			if top >= len(h.stack) {
				return errors.New("truetype: hinting: stack overflow")
			}
			h.stack[top] = int32(top)
			top++

		case opCINDEX, opMINDEX:
			x := int(h.stack[top-1])
			if x <= 0 || x >= top {
				return errors.New("truetype: hinting: invalid data")
			}
			h.stack[top-1] = h.stack[top-1-x]
			if opcode == opMINDEX {
				copy(h.stack[top-1-x:top-1], h.stack[top-x:top])
				top--
			}

		case opRTDG:
			h.roundPeriod = 1 << 5
			h.roundPhase = 0
			h.roundThreshold = 1 << 4

		case opNPUSHB:
			opcode = 0
			goto push

		case opNPUSHW:
			opcode = 0x80
			goto push

		case opWS:
			top -= 2
			i := int(h.stack[top])
			if i < 0 || len(h.store) <= i {
				return errors.New("truetype: hinting: invalid data")
			}
			h.store[i] = h.stack[top+1]

		case opRS:
			i := int(h.stack[top-1])
			if i < 0 || len(h.store) <= i {
				return errors.New("truetype: hinting: invalid data")
			}
			h.stack[top-1] = h.store[i]

		case opDEBUG:
			// No-op.

		case opLT:
			top--
			h.stack[top-1] = bool2int32(h.stack[top-1] < h.stack[top])

		case opLTEQ:
			top--
			h.stack[top-1] = bool2int32(h.stack[top-1] <= h.stack[top])

		case opGT:
			top--
			h.stack[top-1] = bool2int32(h.stack[top-1] > h.stack[top])

		case opGTEQ:
			top--
			h.stack[top-1] = bool2int32(h.stack[top-1] >= h.stack[top])

		case opEQ:
			top--
			h.stack[top-1] = bool2int32(h.stack[top-1] == h.stack[top])

		case opNEQ:
			top--
			h.stack[top-1] = bool2int32(h.stack[top-1] != h.stack[top])

		case opODD, opEVEN:
			i := h.round(f26dot6(h.stack[top-1])) >> 6
			h.stack[top-1] = int32(i&1) ^ int32(opcode-opODD)

		case opIF:
			top--
			if h.stack[top] == 0 {
				opcode = 0
				goto ifelse
			}

		case opEIF:
			// No-op.

		case opAND:
			top--
			h.stack[top-1] = bool2int32(h.stack[top-1] != 0 && h.stack[top] != 0)

		case opOR:
			top--
			h.stack[top-1] = bool2int32(h.stack[top-1]|h.stack[top] != 0)

		case opNOT:
			h.stack[top-1] = bool2int32(h.stack[top-1] == 0)

		case opADD:
			top--
			h.stack[top-1] += h.stack[top]

		case opSUB:
			top--
			h.stack[top-1] -= h.stack[top]

		case opDIV:
			top--
			if h.stack[top] == 0 {
				return errors.New("truetype: hinting: division by zero")
			}
			h.stack[top-1] = int32(f26dot6(h.stack[top-1]).div(f26dot6(h.stack[top])))

		case opMUL:
			top--
			h.stack[top-1] = int32(f26dot6(h.stack[top-1]).mul(f26dot6(h.stack[top])))

		case opABS:
			if h.stack[top-1] < 0 {
				h.stack[top-1] = -h.stack[top-1]
			}

		case opNEG:
			h.stack[top-1] = -h.stack[top-1]

		case opFLOOR:
			h.stack[top-1] &^= 63

		case opCEILING:
			h.stack[top-1] += 63
			h.stack[top-1] &^= 63

		case opROUND00, opROUND01, opROUND10, opROUND11:
			// The four flavors of opROUND are equivalent. See the comment below on
			// opNROUND for the rationale.
			h.stack[top-1] = int32(h.round(f26dot6(h.stack[top-1])))

		case opNROUND00, opNROUND01, opNROUND10, opNROUND11:
			// No-op. The spec says to add one of four "compensations for the engine
			// characteristics", to cater for things like "different dot-size printers".
			// https://developer.apple.com/fonts/TTRefMan/RM02/Chap2.html#engine_compensation
			// This code does not implement engine compensation, as we don't expect to
			// be used to output on dot-matrix printers.

		case opSROUND, opS45ROUND:
			top--
			switch (h.stack[top] >> 6) & 0x03 {
			case 0:
				h.roundPeriod = 1 << 5
			case 1, 3:
				h.roundPeriod = 1 << 6
			case 2:
				h.roundPeriod = 1 << 7
			}
			if opcode == opS45ROUND {
				// The spec says to multiply by √2, but the C Freetype code says 1/√2.
				// We go with 1/√2.
				h.roundPeriod *= 46341
				h.roundPeriod /= 65536
			}
			h.roundPhase = h.roundPeriod * f26dot6((h.stack[top]>>4)&0x03) / 4
			if x := h.stack[top] & 0x0f; x != 0 {
				h.roundThreshold = h.roundPeriod * f26dot6(x-4) / 8
			} else {
				h.roundThreshold = h.roundPeriod - 1
			}

		case opJROT:
			top -= 2
			if h.stack[top+1] != 0 {
				pc += int(h.stack[top])
				continue
			}

		case opJROF:
			top -= 2
			if h.stack[top+1] == 0 {
				pc += int(h.stack[top])
				continue
			}

		case opROFF:
			h.roundPeriod = 0
			h.roundPhase = 0
			h.roundThreshold = 0

		case opRUTG:
			h.roundPeriod = 1 << 6
			h.roundPhase = 0
			h.roundThreshold = 1<<6 - 1

		case opRDTG:
			h.roundPeriod = 1 << 6
			h.roundPhase = 0
			h.roundThreshold = 0

		case opSANGW, opAA:
			// These ops are "anachronistic" and no longer used.
			top--

		case opIDEF:
			// IDEF is for ancient versions of the bytecode interpreter, and is no longer used.
			return errors.New("truetype: hinting: unsupported IDEF instruction")

		case opROLL:
			h.stack[top-1], h.stack[top-3], h.stack[top-2] = h.stack[top-3], h.stack[top-2], h.stack[top-1]

		case opMAX:
			top--
			if h.stack[top-1] < h.stack[top] {
				h.stack[top-1] = h.stack[top]
			}

		case opMIN:
			top--
			if h.stack[top-1] > h.stack[top] {
				h.stack[top-1] = h.stack[top]
			}

		case opPUSHB000, opPUSHB001, opPUSHB010, opPUSHB011, opPUSHB100, opPUSHB101, opPUSHB110, opPUSHB111:
			opcode -= opPUSHB000 - 1
			goto push

		case opPUSHW000, opPUSHW001, opPUSHW010, opPUSHW011, opPUSHW100, opPUSHW101, opPUSHW110, opPUSHW111:
			opcode -= opPUSHW000 - 1
			opcode += 0x80
			goto push

		default:
			return errors.New("truetype: hinting: unrecognized instruction")
		}
		pc++
		continue

	ifelse:
		// Skip past bytecode until the next ELSE (if opcode == 0) or the
		// next EIF (for all opcodes). Opcode == 0 means that we have come
		// from an IF. Opcode == 1 means that we have come from an ELSE.
		{
		ifelseloop:
			for depth := 0; ; {
				pc++
				if pc >= len(program) {
					return errors.New("truetype: hinting: unbalanced IF or ELSE")
				}
				switch program[pc] {
				case opIF:
					depth++
				case opELSE:
					if depth == 0 && opcode == 0 {
						break ifelseloop
					}
				case opEIF:
					depth--
					if depth < 0 {
						break ifelseloop
					}
				case opNPUSHB:
					pc++
					if pc >= len(program) {
						return errors.New("truetype: hinting: unbalanced IF or ELSE")
					}
					pc += int(program[pc])
				case opNPUSHW:
					pc++
					if pc >= len(program) {
						return errors.New("truetype: hinting: unbalanced IF or ELSE")
					}
					pc += 2 * int(program[pc])
				case opPUSHB000, opPUSHB001, opPUSHB010, opPUSHB011, opPUSHB100, opPUSHB101, opPUSHB110, opPUSHB111:
					pc += int(program[pc] - (opPUSHB000 - 1))
				case opPUSHW000, opPUSHW001, opPUSHW010, opPUSHW011, opPUSHW100, opPUSHW101, opPUSHW110, opPUSHW111:
					pc += 2 * int(program[pc]-(opPUSHW000-1))
				default:
					// No-op.
				}
			}
			pc++
			continue
		}

	push:
		// Push n elements from the program to the stack, where n is the low 7 bits of
		// opcode. If the low 7 bits are zero, then n is the next byte from the program.
		// The high bit being 0 means that the elements are zero-extended bytes.
		// The high bit being 1 means that the elements are sign-extended words.
		{
			width := 1
			if opcode&0x80 != 0 {
				opcode &^= 0x80
				width = 2
			}
			if opcode == 0 {
				pc++
				if pc >= len(program) {
					return errors.New("truetype: hinting: insufficient data")
				}
				opcode = program[pc]
			}
			pc++
			if top+int(opcode) > len(h.stack) {
				return errors.New("truetype: hinting: stack overflow")
			}
			if pc+width*int(opcode) > len(program) {
				return errors.New("truetype: hinting: insufficient data")
			}
			for ; opcode > 0; opcode-- {
				if width == 1 {
					h.stack[top] = int32(program[pc])
				} else {
					h.stack[top] = int32(int8(program[pc]))<<8 | int32(program[pc+1])
				}
				top++
				pc += width
			}
			continue
		}
	}
	return nil
}

// f2dot14 is a 2.14 fixed point number.
type f2dot14 int16

// f26dot6 is a 26.6 fixed point number.
type f26dot6 int32

// div returns x/y in 26.6 fixed point arithmetic.
func (x f26dot6) div(y f26dot6) f26dot6 {
	return f26dot6((int64(x) << 6) / int64(y))
}

// mul returns x*y in 26.6 fixed point arithmetic.
func (x f26dot6) mul(y f26dot6) f26dot6 {
	return f26dot6(int64(x) * int64(y) >> 6)
}

// round rounds the given number. The rounding algorithm is described at
// https://developer.apple.com/fonts/TTRefMan/RM02/Chap2.html#rounding
func (h *Hinter) round(x f26dot6) f26dot6 {
	if h.roundPeriod == 0 {
		return x
	}
	neg := x < 0
	x -= h.roundPhase
	x += h.roundThreshold
	if x >= 0 {
		x = (x / h.roundPeriod) * h.roundPeriod
	} else {
		x -= h.roundPeriod
		x += 1
		x = (x / h.roundPeriod) * h.roundPeriod
	}
	x += h.roundPhase
	if neg {
		if x >= 0 {
			x = h.roundPhase - h.roundPeriod
		}
	} else if x < 0 {
		x = h.roundPhase
	}
	return x
}

func bool2int32(b bool) int32 {
	if b {
		return 1
	}
	return 0
}
