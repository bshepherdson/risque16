# Instruction Encoding

Every flavour of instruction is identified with a unique prefix, ranging from 3
to 8 bits. It is likely easiest to treat the top 3 or 4 bits as a table lookup,
and then use if-statements to guide the rest of the decoding.

Conventions: `ooo` or similar specifies the exact operation, `XXXX` is a
literal values, `sss`, `aaa`, `bbb` and `ddd` are 3-bit register numbers
corresponding to operands (`aaa` and `bbb`), shifts (`sss`) and destinations
(`ddd`). `___` is "don't care".

Don't cares must be set to 0s by assemblers; this enables future expansion.

These are presented in numerical order.

| Number | `ooo` | Format | Meaning |
| :--- | :--- | :--- | :--- |
|  1 | `000`      | `000ooXXXXXsssddd` | Move with shift |
|  2 | `001`      | `001oodddXXXXXXXX` | Math with immediate values |
|  3 | `01000`    | `01000_oooosssddd` | Math with 2 registers |
|  4 | `01001`    | `01001Mobbbaaaddd` | Math with 2 regs and 1 reg/imm |
|  5 | `0101`     | `0101SdddXXXXXXXX` | Load address near `SP` or `PC` |
|  6 | `0110`     | `0110ccccXXXXXXXX` | Conditional branches |
|  7 | `01110`    | `01110LXXXXXXXXXX` | Branch, 10-bit relative |
|  8 | `01111`    | `01111L__________` | Branch next-word absolute address |
|  9 | `10000000` | `10000000oo___ddd` | `HWN`, `HWQ`, and `HWI` |
| 10 | `10000001` | `10000001L____ddd` | `MRS` and `MSR` |
| 11 | `10000010` | `10000010______oo` | `RFI`, `IFC` and `IFS` |
| 12 | `10000011` | `10000011XXXXXXXX` | `SWI` with unsigned literal |
| 13 | `10000100` | `10000100_____ddd` | `SWI` with register |
| 14 | `10001`    | `10001L00_____ddd` | `BX` and `BLX` |
| 15 | `1001`     | ???                | Free space? |
| 16 | `10100`    | `10100_LRrrrrrrrr` | `PUSH` and `POP` |
| 17 | `10101`    | `10101___SXXXXXXX` | Adjust `SP` |
| 18 | `1011`     | `1011Lbbbrrrrrrrr` | Multiple load/store |
| 19 | `110`      | `110LPXXXXXbbbddd` | Load/store with immediate offset |
| 20 | `11100`    | `11100LPaaabbbddd` | Load/store with register offset |
| 21 | `11101`    | `11101dddXXXXXXXX` | PC-relative load |
| 22 | `1111`     | `1111LdddXXXXXXXX` | SP-relative load/store |


### Format 1 - Move with shift

`000ooXXXXXsssddd`

- `XXXXX` - shift amount (unsigned)
- `sss` - source register
- `ddd` - destination register

| `oo` | Assembly | Meaning |
| :--- | :--- | :--- |
| `00` | `LSL Rd, Rs, #Imm` | Shift `Rs` left by the 5-bit immediate, store in `Rd`. |
| `01` | `LSR Rd, Rs, #Imm` | Shift `Rs` right logically by the 5-bit immediate, store in `Rd`. |
| `10` | `ASR Rd, Rs, #Imm` | Shift `Rs` right arithmetically by the 5-bit immediate, store in `Rd`. |
| `11` | (unused) | |

1 cycle. Sets `CPSR` condition codes.


### Format 2 - Math with immediate values

`001oodddXXXXXXXX`

- `ddd` - `Rd`, the destination (and maybe left-hand operand) register.
- `XXXXXXXX` - 8-bit unsigned immediate operand

| `oo` | Assembly | Meaning |
| :---: | :--- | :--- |
| `00` | `MOV Rd, #Imm` | `Rd := Imm` |
| `01` | `CMP Rd, #Imm` | Set condition codes based on `Rd - Imm`. |
| `10` | `ADD Rd, #Imm` | `Rd := Rd + Imm` |
| `11` | `SUB Rd, #Imm` | `Rd := Rd - Imm` |

1 cycle. All operations sets the `CPSR` condition codes, not just `CMP`.


### Format 3 - Math with 2 registers

`01000_oooosssddd`

- `sss` - Source register `Rs`. Never updated.
- `ddd` - Destination register `Rd`. Usually set to a result, but not always.
    See below.

| `oooo` | Assembly | Meaning |
| :---: | :--- | :--- |
| `0000` | `AND Rd, Rs` | `Rd := Rd AND Rs` |
| `0001` | `EOR Rd, Rs` | `Rd := Rd EOR Rs` |
| `0010` | `LSL Rd, Rs` | `Rd := Rd << Rs` |
| `0011` | `LSR Rd, Rs` | `Rd := Rd >>> Rs` |
| `0100` | `ASR Rd, Rs` | `Rd := Rd >> Rs` |
| `0101` | `ADC Rd, Rs` | `Rd := Rd + Rs + C-bit` |
| `0110` | `SBC Rd, Rs` | `Rd := Rd - Rs - NOT C-bit` |
| `0111` | `ROR Rd, Rs` | `Rd := Rd ROR Rs` - rotate `Rd` rightward by `Rs` bits. |
| `1000` | `TST Rd, Rs` | Sets condition codes based on `Rd AND Rs`, but don't change `Rd` or `Rs`. |
| `1001` | `NEG Rd, Rs` | `Rd := -Rs` |
| `1010` | `CMP Rd, Rs` | Sets condition codes based on `Rd - Rs` |
| `1011` | `CMN Rd, Rs` | Sets condition codes based on `Rd + Rs` |
| `1100` | `ORR Rd, Rs` | `Rd := Rd OR Rs` |
| `1101` | `MUL Rd, Rs` | `Rd := Rs * Rd` |
| `1110` | `BIC Rd, Rs` | `Rd := Rd AND NOT Rs` ("BIt Clear") |
| `1111` | `MVN Rd, Rs` | `Rd := NOT Rs` (NB: bitwise negate, not integer negative) |

1 cycle, except for `MUL`, which needs 4. Note that all operations set the
`CPSR` condition codes, not just `TST`, `CMP`, etc.


### Format 4 - Add/subtract with 2 regs and 1 reg./imm.

`00011Mobbbaaaddd`

- `bbb` - The second operand. When `M` is 0, a register. When `M` is 1, a
  3-bit unsigned literal.
- `aaa` - The first operand, `Ra`.
- `ddd` - The destination register `Rd`.

| `o` | `M` | Assembly | Meaning |
| :---: | :---: | :--- | :--- |
| `0` | `0` | `ADD Rd, Ra, Rb` | `Rd := Ra + Rb` |
| `0` | `1` | `ADD Rd, Ra, #Imm` | `Rd := Ra + Imm` |
| `1` | `0` | `SUB Rd, Ra, Rb` | `Rd := Ra - Rb` |
| `1` | `1` | `SUB Rd, Ra, #Imm` | `Rd := Ra - Imm` |

1 cycle. Sets `CPSR` condition codes.


### Format 5 - Load address

`0101SdddXXXXXXXX`

- `S` - Source: 0 = `PC`, 1 = `SP`
- `ddd` - Destination register `Rd`
- `XXXXXXXX` - 8-bit unsigned immediate

| `S` | Assembly | Meaning |
| :---: | :--- | :--- |
| `0` | `ADD Rd, PC, #Imm` | Add `Imm` to the current value of `PC`, and load the result into `Rd`. |
| `1` | `ADD Rd, SP, #Imm` | Add `Imm` to the current value of SP, and load the result into `Rd`. |

Note that this puts an address into `Rd`, not the value in memory at that spot!

Note that `PC` points to the instruction after the current one.

1 cycle. Does **not** set `CPSR` condition codes.


### Format 6 - Conditional branches

`0110ccccXXXXXXXX`

These instructions all perform a conditional branch depending on the `CPSR`
condition codes (`N`, `Z`, `C` and `V`).

The offset `XXXXXXXX` is a **signed** (2's complement) 8-bit offset from the
current `PC`. Recall that `PC` points at the next instruction, not at this one.

| `cccc` | Assembly | Meaning |
| :----: | :--- | :--- |
| `0000` | `BEQ label` | Branch if `Z` set (equal) |
| `0001` | `BNE label` | Branch if `Z` clear (not equal) |
| `0010` | `BCS label` | Branch if `C` set (unsigned higher or same) |
| `0011` | `BCC label` | Branch if `C` clear (unsigned lower) |
| `0100` | `BMI label` | Branch if `N` set (negative) |
| `0101` | `BPL label` | Branch if `N` clear (positive or zero) |
| `0110` | `BVS label` | Branch if `V` set (overflow) |
| `0111` | `BVC label` | Branch if `V` clear (no overflow) |
| `1000` | `BHI label` | Branch if `C` set and `Z` clear (unsigned higher) |
| `1001` | `BLS label` | Branch if `C` clear or `Z` set (unsigned lower or same) |
| `1010` | `BGE label` | Branch if `N` and `V` match (signed greater or equal) |
| `1011` | `BLT label` | Branch if `N` and `V` differ (signed less than) |
| `1100` | `BGT label` | Branch if `Z` clear, and `N` and `V` match (signed greater than) |
| `1101` | `BLE label` | Branch if `Z` set, or `N` and `V` differ (signed less than or equal) |
| `1110` | (illegal) | (Undefined opcode.) |
| `1111` | (illegal) | (Undefined opcode.) |

Branches take 2 cycles if they fail and 1 cycle to they succeed. (Prefetching
expects conditional branches to succeed.)

Branches examine the `CPSR` condition codes but don't change them. This allows
chaining multiple branches based on the same results.


### Format 7 - Branch, 10-bit relative

`01110LXXXXXXXXXX`

Unconditionally branches to a 10-bit signed (2's complement) offset from the
current `PC`.

`L` is the *link flag*, when 1 `LR` is set to `PC` (the address of the next
instruction).

1 cycle. Does **not** change `CPSR`.


### Format 8 - Branch to absolute address

`01111L__________`

Reads the next word, and branches to that absolute, 16-bit address.

When `L` is 1, `LR` is set to `PC` (the address of the next instruction).


### Format 9 - Hardware

`10000000oo___ddd`

- `oo` - Operation.
- `ddd` - Source or destination register.

| `oo` | Assembly | Meaning |
| :---: | :--- | :--- |
| `00` | `HWN Rd` | Sets `Rd` to the number of attached hardware devices. |
| `01` | `HWQ Rd` | Queries the device whose number is in `Rd`, setting registers as below. |
| `10` | `HWI Rd` | Sends a hardware interrupt to the device whose number is in `Rd` |
| `11` | (illegal) | |

4 cycles, and more if the hardware blocks execution.

Does not change `CPSR` condition codes.

`r1:r0` are set to the 32-bit device ID, `r2` to the version, `r4:r3` to the
32-bit manufacturer ID.


### Format 10 - `SPSR`

`10000001L____ddd`

| `L` | Assembly | Meaning |
| :---: | :--- | :--- |
| `0` | `MSR Rd` | Move `SPSR` to `Rd`. |
| `1` | `MRS Rd` | Move `Rd` to `SPSR`. |

1 cycle. Does **not** change `CPSR` (only `SPSR`).


### Format 11 - Interrupt handling

`10000010______oo`

| `oo` | Assembly | Meaning |
| :---: | :--- | :--- |
| `00` | `RFI` | Return from interrupt: pop `r0`, pop `pc`, move `SPSR` to `CPSR`. |
| `01` | `IFC` | Clear the interrupt bit in `CPSR`. |
| `10` | `IFS` | Set the interrupt bit in `CPSR`. |
| `11` | (illegal) | |

`RFI` is 4 cycles. `IFC` and `IFS` are 1.


### Format 12 - SWI with immediate

`10000011XXXXXXXX`

Adds an interrupt with the 8-bit immediate message.

1 cycle. Does **not** change `CPSR`.


### Format 13 - SWI with register

`10000100_____ddd`

Adds an interrupt with the value of `Rd` as the message.

1 cycle. Does **not** change `CPSR`.


### Format 14 - BX and BLX

`10001L00_____ddd`

If `L` is 1, sets `LR` to `PC`. Then sets `PC` to the value of `Rd`.

1 cycle. Does **not** change `CPSR`.


### Format 15 - Unused

Free space.

### Format 16 - Push and pop

`10100_LRrrrrrrrr`

- `L` - 0 = store/push, 1 = load/pop
- `R` - 1 = store `LR`/load `PC`
- `rrrrrrrr` - Register list, `r0` = LSB, `r7` = MSB, set to load/store that
  register.

| `L` | `R` | Assembly | Meaning |
| :---: | :---: | :-- | :-- |
| `0` | `0` | `PUSH { Rlist }` | Push the registers onto the stack, update `SP`. |
| `0` | `1` | `PUSH { Rlist, LR }` | Push the registers and `LR` onto the stack, update `SP`. |
| `1` | `0` | `POP { Rlist }` | Pop the registers from the stack, update `SP`. |
| `1` | `1` | `POP { Rlist, PC }` | Pop the registers and `PC` from the stack, update `SP`. |

Registers are pushed in "ascending" order in memory. Pushing `r0`, `r3`, and
`LR` looks like:

```
sp+3: ...
sp+2: LR
sp+1: r3
sp+0: r0
```

Popping expects them in the same order, of course.

Note that popping `PC` constitutes a branch (usually a return to a previous `BL`
or `BLX`).

1 cycle per regular register stored/loaded. 1 extra for storing `LR`, and 2 extra
for loading `PC`.

Does **not** set the `CPSR` condition codes.


### Format 17 - Adjust `SP`

`10101___SXXXXXXXX`

- `S` - 0 = add, 1 = subtract
- `XXXXXXX` - 7-bit unsigned immediate

| `S` | Assembly | Meaning |
| :---: | :-- | :-- |
| `0` | `ADD SP, #Imm` | Add `Imm` to `SP` |
| `1` | `SUB SP, #Imm` | Subtract `Imm` from `SP` |

1 cycle. Does **not** set `CPSR` condition codes.


### Format 18 - Multiple load/store

`1011Lbbbrrrrrrrr`

- `L` - 0 = store, 1 = load
- `bbb` - Base register `Rb`
- `rrrrrrrr` - Register list. `r0` = LSB, `r7` = MSB. Set to load/store that
  register.

| `L` | Assembly | Meaning |
| :---: | :-- | :-- |
| `0` | `STMIA Rb!, { Rlist }` | Store the registers from Rlist starting at the address given by `Rb`. |
| `1` | `LDMIA Rb!, { Rlist }` | Load the registers in Rlist starting at the address given by `Rb`. |

The registers are stored in ascending order, with the lowest-numbered register
at the original `[Rb]`. `Rb` is updated after this operation, with the result
that `Rb` ends up pointing after the last word written. See Format 16 for a
detailed example.

1 cycle per register loaded or stored. It is an illegal instruction for
`rrrrrrrr` to be empty.

Does **not** set the `CPSR` condition codes.


### Format 19 - Load/store with immediate offset

`110LPXXXXXbbbddd`

- `L` - 0 = store, 1 = load
- `P` - 0 = pre-indexed load/store (no writeback), 1 = post-increment (writeback)
- `XXXXX` - 5-bit unsigned immediate offset
- `bbb` - Base register `Rb`
- `ddd` - Source/destination register `Rd`

| `L` | `P` | Assembly | Meaning |
| :---: | :---: | :--- | :--- |
| `0` | `0` | `STR Rd, [Rb, #Imm]` | Store `Rd` to memory at `Rb + Imm`. |
| `0` | `1` | `STR Rd, [Rb], #Imm` | Store `Rd` to memory at `Rb`, then `Rb := Rb + Imm`. |
| `1` | `0` | `LDR Rd, [Rb, #Imm]` | Read memory at `Rb + Imm`, store in `Rd`. |
| `1` | `1` | `LDR Rd, [Rb], #Imm` | Read memory at `Rb`, store in `Rd`, then `Rb := Rb + Imm`. |

1 cycle. Does **not** set `CPSR` condition codes.


### Format 20 - Load/store with register offset

`11110LPaaabbbddd`

- `L` - 0 = store, 1 = load
- `P` - 0 = pre-indexed load/store (no writeback), 1 = post-increment (writeback)
- `aaa` - Offset register `Ra`
- `bbb` - Base register `Rb`
- `ddd` - Source/Destination register `Rd`

| `L` | `P` | Assembly | Meaning |
| :---: | :---: | :--- | :--- |
| `0` | `0` | `STR Rd, [Rb, Ra]` | Store `Rd` to memory at `Rb + Ra`. |
| `0` | `1` | `STR Rd, [Rb], Ra` | Store `Rd` to memory at `Rb`, then `Rb := Rb + Ra`. |
| `1` | `0` | `LDR Rd, [Rb, Ra]` | Store `Rd` to memory at `Rb + Ra`. |
| `1` | `1` | `LDR Rd, [Rb], Ra` | Store `Rd` to memory at `Rb`, then `Rb := Rb + Ra`. |

1 cycle. Does **not** set `CPSR` condition codes.


### Format 21 - PC-relative load

`11101dddXXXXXXXX`

- `ddd` - Destination register `Rd`
- `XXXXXXXX` - 8-bit unsigned offset

`LDR Rd, [PC, #Imm]`

Remember that `PC` points at the next instruction during execution of this one.
(Put another way, `LDR Rd, [PC, #0]` will load the next instruction into `Rd`.)

1 cycle. Does **not** set `CPSR` condition codes.


### Format 22 - SP-relative load/store

`1111LdddXXXXXXXX`

- `L` - 0 = store, 1 = load
- `ddd` - Source/destination register `Rd`
- `XXXXXXXX` - 8-bit unsigned offset from `SP`

Note that `SP` is expected to be a "full-descending" stack, that is `SP`
points at the value on top of the stack.

| `L` | Assembly | Meaning |
| :---: | :--- | :--- |
| `0` | `STR Rd, [SP, #Imm]` | Store `Rd` to memory at `SP + Imm`. |
| `1` | `LDR Rd, [Rb, #Imm]` | Read memory at `SP + Imm`, write to `Rd`. |

1 cycle. Does **not** set `CPSR` condition codes.


