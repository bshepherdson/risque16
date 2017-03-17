# Instruction Encoding

Risque-16 instructions are encoded in one of 5 formats, which are detailed
below. Each format is identified with a unique prefix, ranging from 1 to 3 bits.

## Notation

- `ooo` represents an operation number.
- `XXXXX` is a literal value
- `ddd` is the "destination" register
- `aaa` and `bbb` are argument registers.
- `rrrrrrrr` is a bitmap of registers to load/store.


## Format Overview

| Format      | Prefix | Details            | Description                            |
| :---        | :---   | :---               | :---                                   |
| Immediate   | `0xx`  | `0oooodddXXXXXXXX` | 1 register and 8-bit immediate         |
| Registers   | `100`  | `100oooobbbaaaddd` | 0-3 registers                          |
| Branch      | `101`  | `101ooooXXXXXXXXX` | Conditional and unconditional branches |
| Memory      | `110`  | `110ooodddbbbXXXX` | Memory access                          |
| Multi-store | `111`  | `111oobbbrrrrrrrr` | `PUSH`, `POP`, `LDMIA`, `STMIA`        |



## Immediate Format

`0oooodddXXXXXXXX`

Prefixed with `0xx`, all of these operations have a register destination and an
8-bit unsigned immediate value.

If the opcode is 0, the destination register is special (eg. `SP`), and the real
opcode is in the `ddd` slot.

### Normal Immediate Instructions

| Op   | Assembly           | Cycles | Flags  | Meaning |
| :--- | :---               | :---   | :---   | :--- |
| `$0` | (Special format)   | -      | - | See the table below. |
| `$1` | `MOV Rd, #Imm`     | 1      | `NZ00` | `Rd := Imm` |
| `$2` | `NEG Rd, #Imm`     | 1      | `NZ00` | `Rd := -Imm` - 2's complement negative |
| `$3` | `CMP Rd, #Imm`     | 1      | `NZCV` | Sets flags on `Rd - Imm`; `Rd` unchanged |
| `$4` | `ADD Rd, #Imm`     | 1      | `NZCV` | `Rd := Rd +   Imm` |
| `$5` | `SUB Rd, #Imm`     | 1      | `NZCV` | `Rd := Rd -   Imm` |
| `$6` | `MUL Rd, #Imm`     | 4      | `NZCV` | `Rd := Rd *   Imm` |
| `$7` | `LSL Rd, #Imm`     | 1      | `NZCV` | `Rd := Rd <<  Imm` |
| `$8` | `LSR Rd, #Imm`     | 1      | `NZCV` | `Rd := Rd >>> Imm` |
| `$9` | `ASR Rd, #Imm`     | 1      | `NZCV` | `Rd := Rd >>  Imm` |
| `$a` | `AND Rd, #Imm`     | 1      | `NZ00` | `Rd := Rd & Imm` |
| `$b` | `ORR Rd, #Imm`     | 1      | `NZ00` | `Rd := Rd | Imm` |
| `$c` | `XOR Rd, #Imm`     | 1      | `NZ00` | `Rd := Rd ^ Imm` |
| `$d` | `ADD Rd, PC, #Imm` | 1      | `NZCV` | `Rd := PC + Imm` (PC points after this instruction) |
| `$e` | `ADD Rd, SP, #Imm` | 1      | `NZCV` | `Rd := SP + Imm` |
| `$f` | `MVH Rd, #Imm`     | 1      | `NZ00` | `Rd := (Rd & 0xff) | (Imm << 8)` - move high half-word |

Note that these take 1 cycle each, except for `MUL`, which takes 4 cycles.


### Special Immediate Instructions

These have `oooo` set to 0, and no `Rd`: the `ddd` slot gives the operation.

| Op   | Assembly       | Cycles | Flags  | Meaning |
| :--- | :---           | :---   | :---   | :--- |
| `$0` | `ADD SP, #Imm` | 1      | `----` | `SP := SP + Imm` |
| `$1` | `SUB SP, #Imm` | 1      | `----` | `SP := SP - Imm` |
| `$2` | `SWI #Imm`     | 4      | `----` | Pushes interrupt with message `Imm` |
| `$3-7` | (reserved) | | |



## Register Format

`100oooobbbaaaddd`

This format has a 4-bit opcode, and 3 3-bit register fields, in that order.

Each format can punt to the next, if it doesn't need 3 registers as operands. It
does that by settings its opcode to 0.

### 3-Register Instructions

| Op   | Assembly            | Cycles | Flags  | Meaning                           |
| :--- | :---                | :---   | :---   | :---                              |
| `$0` | (2-register format) |        |        | See 2-Register Instructions below |
| `$1` | `ADD Rd, Ra, Rb`    | 1      | `NZCV` | `Rd := Ra + Rb`                   |
| `$2` | `ADC Rd, Ra, Rb`    | 1      | `NZCV` | `Rd := Ra + Rb + C-bit`           |
| `$3` | `SUB Rd, Ra, Rb`    | 1      | `NZCV` | `Rd := Ra - Rb`                   |
| `$4` | `SBC Rd, Ra, Rb`    | 1      | `NZCV` | `Rd := Ra - Rb - NOT C-bit`       |
| `$5` | `MUL Rd, Ra, Rb`    | 4      | `NZCV` | `Rd := Ra * Rb`                   |
| `$6` | `LSL Rd, Ra, Rb`    | 1      | `NZC0` | `Rd := Ra << Rb`                  |
| `$7` | `LSR Rd, Ra, Rb`    | 1      | `NZC0` | `Rd := Ra >>> Rb`                 |
| `$8` | `ASR Rd, Ra, Rb`    | 1      | `NZC0` | `Rd := Ra >>  Rb`                 |
| `$9` | `AND Rd, Ra, Rb`    | 1      | `NZ00` | `Rd := Ra & Rb`                   |
| `$a` | `ORR Rd, Ra, Rb`    | 1      | `NZ00` | `Rd := Ra | Rb`                   |
| `$b` | `XOR Rd, Ra, Rb`    | 1      | `NZ00` | `Rd := Ra ^ Rb`                   |


### 2-Register Instructions

| `bbb` | Assembly            | Cycles | Flags  | Meaning |
| :---  | :---                | :---   | :---   | :---    |
| `$0`  | (1-register format) |        |        |         |
| `$1`  | `MOV Rd, Rs`        | 1      | `NZ00` | `Rd := Rs` |
| `$2`  | `CMP Rd, Rs`        | 1      | `NZCV` | Flags set based on `Rd - Rs`; `Rd` unchanged |
| `$3`  | `CMN Rd, Rs`        | 1      | `NZCV` | Flags set based on `Rd + Rs`; `Rd` unchanged |
| `$4`  | `ROR Rd, Rs`        | 1      | `NZC0` | Rotates `Rd` right by `Rs` bits |
| `$5`  | `NEG Rd, Rs`        | 1      | `NZ00` | `Rd := -Rs` 2's complement negation |
| `$6`  | `TST Rd, Rs`        | 1      | `NZ00` | Flags set based on `Rd AND Rs`; `Rd` unchanged |
| `$7`  | `MVN Rd, Rs`        | 1      | `NZ00` | `Rd := NOT Rs` (bitwise negate) |


### 1-Register Instructions

| `aaa` | Assembly            | Cycles | Flags   | Meaning |
| :---  | :---                | :---   | :---    | :---    |
| `$0`  | (0-register format) |        |         |         |
| `$1`  | `BX Rd`             | 1      | `----`  | Branch to address in `Rd` |
| `$2`  | `BLX Rd`            | 1      | `----`  | Branch to address in `Rd`; `LR := PC` |
| `$3`  | `SWI Rd`            | 4      | `----`  | Push interrupt with `Rd` as message |
| `$4`  | `HWN Rd`            | 4      | `----`  | Sets `Rd` to the number of devices |
| `$5`  | `HWQ Rd`            | 4      | `----`  | Sets registers with details for device `Rd` (0-based) - see below |
| `$6`  | `HWI Rd`            | 4      | `----`  | Sends an interrupt to device `Rd` (0-based) |
| `$7`  | `XSR Rd`            | 2      | Special | Exchange `SPSR` and `Rd` |

`HWQ` sets `r1:r0` to the 32-bit device ID, `r2` to the version, `r4:r3` to the
32-bit manufacturer ID.



### 0-Register Instructions

| `ddd` | Assembly | Cycles | Flags   | Meaning |
| :---  | :---     | :---   | :---    | :---    |
| `$0`  | `RFI`    | 4      | Special | Return from interrupt (see below) |
| `$1`  | `IFS`    | 1      | `----`  | Sets the `I` status bit, enabling interrupts |
| `$2`  | `IFC`    | 1      | `----`  | Clear the `I` status bit, disabling interrupts |
| `$3`  | `RET`    | 1      | `----`  | `PC := LR`, returns from simple subroutines |
| `$4`  | `POPSP`  | 1      | `----`  | `SP := [SP]` |

Note that `RET` is not required for returning from subroutines; it's just a
helper for when you don't need to save `LR` to the stack.


## Branch Format

`101ooooXXXXXXXXX`

The literal is an 9-bit signed offset. It's relative to the next instruction.

| Op   | Assembly    | Cycles | Meaning |
| :--- | :---        | :---   | :---    |
| `$0` | `B   label` | 1 or 2 | Branch always |
| `$1` | `BL  label` | 1 or 2 | Branch always; `LR := PC` |
| `$2` | `BEQ label` | 1 or 2 | Branch if `Z` set (equal) |
| `$3` | `BNE label` | 1 or 2 | Branch if `Z` clear (not equal) |
| `$4` | `BCS label` | 1 or 2 | Branch if `C` set (unsigned higher or same) |
| `$5` | `BCC label` | 1 or 2 | Branch if `C` clear (unsigned lower) |
| `$6` | `BMI label` | 1 or 2 | Branch if `N` set (negative) |
| `$7` | `BPL label` | 1 or 2 | Branch if `N` clear (positive or zero) |
| `$8` | `BVS label` | 1 or 2 | Branch if `V` set (overflow) |
| `$9` | `BVC label` | 1 or 2 | Branch if `V` clear (no overflow) |
| `$a` | `BHI label` | 1 or 2 | Branch if `C` set and `Z` clear (unsigned higher) |
| `$b` | `BLS label` | 1 or 2 | Branch if `C` clear or `Z` set (unsigned lower or same) |
| `$c` | `BGE label` | 1 or 2 | Branch if `N` and `V` match (signed greater or equal) |
| `$d` | `BLT label` | 1 or 2 | Branch if `N` and `V` differ (signed less than) |
| `$e` | `BGT label` | 1 or 2 | Branch if `Z` clear, and `N` and `V` match (signed greater than) |
| `$f` | `BLE label` | 1 or 2 | Branch if `Z` set, or `N` and `V` differ (signed less than or equal) |

A branch offset is needed that signals we're using the long form.

Using 0 doesn't work, it can cause an infinite loop. If the actual target is the
next word, the diff would be 0 but we need to use the long form. That makes the
offset 1, not 0. Then we can use the short form, etc.

So instead we use -1. That would loop back to the branch itself, but that can
be encoded as the long form without trouble.

Branches take 1 cycle on success in short form. Long-form success, or failure
in either form, takes 2 cycles.



## Memory-access Format

`110ooodddbbbXXXX`

Depending on the opcode (`ooo`) the lower 4 bits are either:

- `XXXX`, a 4-bit unsigned literal, or
- `0aaa`, a register

which is used for indexing or post-incrementing, as appropriate.

| Op   | Assembly             | Cycles | Meaning                                             |
| :--- | :---                 | :---   | :---                                                |
| `$0` | `LDR Rd, [Rb], #inc` | 1      | Load `Rd` from `[Rb]`, then increment `Rb` by `inc` |
| `$1` | `STR Rd, [Rb], #inc` | 1      | Store `Rd` at `[Rb]`, then increment `Rb` by `inc`  |
| `$2` | `LDR Rd, [Rb, #inc]` | 1      | Load `Rd` from `[Rb+inc]` (`Rb` unchanged)          |
| `$3` | `STR Rd, [Rb, #inc]` | 1      | Store `Rd` at `[Rb+inc]` (`Rb` unchanged)           |
| `$4` | `LDR Rd, [Rb, Ra]`   | 2      | Load `Rd` from `[Rb+Ra]` (`Rb`, `Ra` unchanged)     |
| `$5` | `STR Rd, [Rb, Ra]`   | 2      | Store `Rd` at `[Rb+Ra]` (`Rb`, `Ra` unchanged)      |
| `$6` | `LDR Rd, [SP, #inc]` | 1      | Load `Rd` from `[SP+inc]` (`SP` unchanged)          |
| `$7` | `STR Rd, [SP, #inc]` | 1      | Store `Rd` at `[SP+inc]` (`SP` unchanged)           |

Notes:

- Assemblers shall accept `LDR Rd, [Rb]` and encode it as `$0` with `inc=0`.
- Only the post-increment operations (`$0` and `$1`) change `Rb` or `SP`.
- Full-register indexing (`$4` and `$5`) takes 2 cycles, not 1.
- None of these modify the condition flags.


## Multiple Load/Store Format

`111oobbbrrrrrrrr`

Here `rrrrrrrr` is a bitmap of registers to save or load. `r0` corresponds to
the least significant bit, and `r7` to the most significant.

For `PUSH` and `POP`, `bbb` is actually just a single bit `00P`, which indicates
that `LR` should be saved on `PUSH`, or `PC` loaded on `POP`.

| Op   | Assembly             | Meaning                           |
| :--- | :---                 | :---                              |
| `$0` | `POP   {r0, r3, pc}` | Load `r0`, `r3`, etc. from `[SP]` |
| `$1` | `PUSH  {r0, r3, lr}` | Store `r0`, `r3`, etc. at `[SP]`  |
| `$2` | `LDMIA Rb, {r0, r3}` | Load `r0`, `r3`, etc. from `[Rb]` |
| `$3` | `STMIA Rb, {r0, r3}` | Store `r0`, `r3`, etc. at `[Rb]`  |

### Push and Pop

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

1 cycle per register stored/loaded (including LR and PC). Loading `PC` costs one
extra. (Eg. `push {r0, r3, lr}` is 3 cycles; `pop {r0, r3, pc}` is 4.)


### Load/Store Multiple

The registers are stored in ascending order, with the lowest-numbered register
at the original `[Rb]`. `Rb` is updated after this operation, with the result
that `Rb` ends up pointing after the last word written.

Note that because `Rb` moves, it cannot be reused to load the same values again.
`Rb` must be reset first.

1 cycle per register loaded or stored. It is an illegal instruction for
`rrrrrrrr` to be empty.

Does **not** set the `CPSR` condition codes.

