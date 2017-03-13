# Assembler

This is a guide to using the assembler to produce code for the Risque-16.

## Labels

Labels are defined with a leading colon:

```
:foo
  ldr r1, [r2]
```

Labels must begin with a letter or underscore, and are composed of letters,
underscores, and digits.

## Literals

Numeric literals are in decimal. Hex literals begin with `0x`. Binary literals
begin with `0b`.

Literals in instructions must be preceded with a `#`.

### Expressions

Labels and literals can be combined into compound expressions, using the usual
rules of parsing and precedence.

`+`, `-`, `*`, `/`, `&`, `|`, `>>` and `<<` are supported, as are parentheses.



## Instructions

Here are guides to all the different instruction families. This is a
programmer's view, with the instructions grouped by meaning and usage, not based
on their binary encoding.

Immediate values are given as `[US]len`, where `U7` means unsigned 7-bit, and
`S5` means signed (2's complement) 5-bit.

### Moving Data


| Instruction    | Cycles | Flags?  | Meaning                                |
| :---           | :---:  | :---    | :---                                   |
| `MOV Rd, #Imm` | 1      | `NZ00`  | `Rd := Imm`                            |
| `MOV Rd, Rs`   | 1      | `NZ00`  | `Rd := Rs`                             |
| `MVH Rd, #Imm` | 1      | `NZ00`  | `Rd := (Rd & 0xff) | (Imm << 8)`       |
| `MVN Rd, Rs`   | 1      | `NZ00`  | `Rd := NOT Rs` (bitwise negate)        |
| `NEG Rd, #Imm` | 1      | `NZ00`  | `Rd := -Imm` - 2's complement negative |
| `NEG Rd, Rs`   | 1      | `NZ00`  | `Rd := -Rs` 2's complement negation    |
| `XSR Rd`       | 2      | Special | Exchange `SPSR` and `Rd`               |

Note that the two-instruction sequence `MOV; MVH` suffices to load a full 16-bit
value into a register.

Assemblers should allow `MOV Rd, #imm` with any immediate, even one that's too
large for a single `MOV`, and assemble it as some combination of `MOV`, `NEG`,
`XOR` or `MVH`, whatever is most efficient.


### Arithmetic

| Instruction        | Cycles | Flags? | Meaning                     |
| :---               | :---:  | :---   | :---                        |
| `ADD Rd, #Imm`     | 1      | `NZCV` | `Rd := Rd + Imm`            |
| `ADD Rd, Ra, Rb`   | 1      | `NZCV` | `Rd := Ra + Rb`             |
| `ADC Rd, Ra, Rb`   | 1      | `NZCV` | `Rd := Ra + Rb + C-bit`     |
| `SUB Rd, #Imm`     | 1      | `NZCV` | `Rd := Rd - Imm`            |
| `SUB Rd, Ra, Rb`   | 1      | `NZCV` | `Rd := Ra - Rb`             |
| `SBC Rd, Ra, Rb`   | 1      | `NZCV` | `Rd := Ra - Rb - NOT C-bit` |
| `MUL Rd, #Imm`     | 4      | `NZCV` | `Rd := Rd * Imm`            |
| `MUL Rd, Ra, Rb`   | 1      | `NZCV` | `Rd := Ra * Rb`             |
| `ADD Rd, PC, #Imm` | 1      | `NZCV` | `Rd := PC + Imm`            |
| `ADD Rd, SP, #Imm` | 1      | `NZCV` | `Rd := SP + Imm`            |
| `ADD SP, #Imm`     | 1      | `----` | `SP := SP + Imm`            |
| `SUB SP, #Imm`     | 1      | `----` | `SP := SP - Imm`            |

Remember that `PC` points at the instruction after this one.


### Bitwise Arithmetic

| Instruction      | Cycles | Flags? | Meaning                        |
| :---             | :---:  | :---   | :---                           |
| `LSL Rd, Ra, Rb` | 1      | `NZC0` | `Rd := Ra <<  Rb`              |
| `LSL Rd, #Imm`   | 1      | `NZCV` | `Rd := Rd <<  Imm`             |
| `LSR Rd, Ra, Rb` | 1      | `NZC0` | `Rd := Ra >>> Rb`              |
| `LSR Rd, #Imm`   | 1      | `NZCV` | `Rd := Rd >>> Imm`             |
| `ASR Rd, Ra, Rb` | 1      | `NZC0` | `Rd := Ra >>  Rb`              |
| `ASR Rd, #Imm`   | 1      | `NZCV` | `Rd := Rd >>  Imm`             |
| `AND Rd, Ra, Rb` | 1      | `NZ00` | `Rd := Ra & Rb`                |
| `AND Rd, #Imm`   | 1      | `NZ00` | `Rd := Rd & Imm`               |
| `EOR Rd, Ra, Rb` | 1      | `NZ00` | `Rd := Ra | Rb`                |
| `EOR Rd, #Imm`   | 1      | `NZ00` | `Rd := Rd | Imm`               |
| `XOR Rd, Ra, Rb` | 1      | `NZ00` | `Rd := Ra ^ Rb`                |
| `XOR Rd, #Imm`   | 1      | `NZ00` | `Rd := Rd ^ Imm`               |
| `ROR Rd, Rs`     | 1      | `NZC0` | Rotate `Rd` right by `Rs` bits |


### Comparison

These operations perform arithmetic but don't save the result anywhere - they
just set the condition flags in `CPSR`.

| Instruction    | Cycles | Flags? | Meaning                                        |
| :---           | :---:  | :---   | :---                                           |
| `CMP Rd, Rs`   | 1      | `NZCV` | Flags set based on `Rd - Rs`; `Rd` unchanged   |
| `CMP Rd, #Imm` | 1      | `NZCV` | Sets flags on `Rd - Imm`; `Rd` unchanged       |
| `CMN Rd, Rs`   | 1      | `NZCV` | Flags set based on `Rd + Rs`; `Rd` unchanged   |
| `TST Rd, Rs`   | 1      | `NZ00` | Flags set based on `Rd AND Rs`; `Rd` unchanged |


### Control Flow

| Instruction | Cycles | Meaning                                                              |
| :---        | :---:  | :---                                                                 |
| `RET`       | 1      | `PC := LR`                                                           |
| `BX Rd`     | 1      | Branch to address in `Rd`                                            |
| `BLX Rd`    | 1      | `LR := PC`, branch to address in `Rd`                                |
| `B   label` | 1      | Branch always                                                        |
| `BL  label` | 1      | Branch always; `LR := PC`                                            |
| `BEQ label` | 1 or 2 | Branch if `Z` set (equal)                                            |
| `BNE label` | 1 or 2 | Branch if `Z` clear (not equal)                                      |
| `BCS label` | 1 or 2 | Branch if `C` set (unsigned higher or same)                          |
| `BCC label` | 1 or 2 | Branch if `C` clear (unsigned lower)                                 |
| `BMI label` | 1 or 2 | Branch if `N` set (negative)                                         |
| `BPL label` | 1 or 2 | Branch if `N` clear (positive or zero)                               |
| `BVS label` | 1 or 2 | Branch if `V` set (overflow)                                         |
| `BVC label` | 1 or 2 | Branch if `V` clear (no overflow)                                    |
| `BHI label` | 1 or 2 | Branch if `C` set and `Z` clear (unsigned higher)                    |
| `BLS label` | 1 or 2 | Branch if `C` clear or `Z` set (unsigned lower or same)              |
| `BGE label` | 1 or 2 | Branch if `N` and `V` match (signed greater or equal)                |
| `BLT label` | 1 or 2 | Branch if `N` and `V` differ (signed less than)                      |
| `BGT label` | 1 or 2 | Branch if `Z` clear, and `N` and `V` match (signed greater than)     |
| `BLE label` | 1 or 2 | Branch if `Z` set, or `N` and `V` differ (signed less than or equal) |

Instruction pre-fetching assumes branches will succeed, so they take 1 cycle on
success. They cost 2 cycles when they fail.

All of these (except `BX` and `BLX`) take either a relative offset or an
absolute address in the next word. Assemblers should take care of this on their
own, but programmers should be aware of this. (The encoding is to set the
relative branch to 0, and make the next word the absolute target address.)


#### On Returns

The usual flow for a subroutine call is to use `BL(X)` to enter it, which sets
`LR` to the return address. If the subroutine is saving registers to the stack,
it saves `LR` with them, and pops it into `PC` to return.

However, a simple subroutine that doesn't use `BL(X)` to make any calls can use
`RET` to load `PC` directly from `LR, which is faster and simpler.


### Load and Store

None of these modify the flags.

These come in three flavours: indexing, post-incrementing, and `SP`-relative.

| Instruction          | Cycles | Meaning                                             |
| :---                 | :---   | :---                                                |
| `LDR Rd, [Rb], #inc` | 1      | Load `Rd` from `[Rb]`, then increment `Rb` by `inc` |
| `STR Rd, [Rb], #inc` | 1      | Store `Rd` at `[Rb]`, then increment `Rb` by `inc`  |
| `LDR Rd, [Rb, #inc]` | 1      | Load `Rd` from `[Rb+inc]` (`Rb` unchanged)          |
| `STR Rd, [Rb, #inc]` | 1      | Store `Rd` at `[Rb+inc]` (`Rb` unchanged)           |
| `LDR Rd, [Rb, Ra]`   | 2      | Load `Rd` from `[Rb+Ra]` (`Rb`, `Ra` unchanged)     |
| `STR Rd, [Rb, Ra]`   | 2      | Store `Rd` at `[Rb+Ra]` (`Rb`, `Ra` unchanged)      |
| `LDR Rd, [SP, #inc]` | 1      | Load `Rd` from `[Rb+Ra]` (`Rb`, `Ra` unchanged)     |
| `STR Rd, [SP, #inc]` | 1      | Store `Rd` at `[Rb+Ra]` (`Rb`, `Ra` unchanged)      |


### Hardware

| Instruction | Cycles | Meaning                                                |
| :---        | :---:  | :---                                                   |
| `HWN Rd`    | 4      | `Rd := # of connected devices`                         |
| `HWQ Rd`    | 4      | Sets `r0` - `r4` to the device info for device `Rd`.   |
|             |        | (`r1:r0` = ID, `r2` = version, `r4:r3` = manufacturer) |
| `HWI Rd`    | 4      | Sends a hardware interrupt to device `Rd`.             |


### Interrupts

| Instruction | Cycles | Flags?  | Meaning                                             |
| :---        | :---:  | :---    | :---                                                |
| `SWI #U8`   | 4      | No      | Triggers a `SWI` with code `U8`.                    |
| `SWI Rd`    | 4      | No      | Triggers a `SWI` with code in `Rd`.                 |
| `RFI`       | 1      | Special | Returns from an interrupt: `CPSR := SPSR`, pop `r0` |

### Status Register

| Instruction | Cycles | Flags? | Meaning |
| :--- | :---: | :--- | :--- |
| `IFS` | 1 | Special | Sets `I` in `CPSR`; doesn't change condition codes. |
| `IFC` | 1 | Special | Clears `I` in `CPSR`; doesn't change condition codes. |
| `XSR Rd` | 2 | No | Exchanges `SPSR` with `Rd` |

Be careful with `XSR`, and in particular whether it might enable or disable
interrupts.


### Multiple Load/Store

`Rlist` is a comma-separated list of general-purpose registers (`r0` to `r7`).

Their order in the list is irrelevant; they always get stored in ascending order.

| Instruction           | Cycles     | Flags? | Meaning                                                                          |
| :---                  | :---:      | :---   | :---                                                                             |
| `PUSH { Rlist }`      | 1 each     | No     | Writes registers ascending in memory, into the stack.                            |
| `PUSH { Rlist, LR }`  | 1 each     | No     | Writes registers ascending in memory, into the stack.                            |
| `POP { Rlist }`       | 1 each     | No     | Loads registers from the stack.                                                  |
| `POP { Rlist, PC }`   | 1 + 1 each | No     | Loads registers from the stack, including PC                                     |
| `STMIA Rb, { Rlist }` | 1 each     | No     | Store registers from `Rlist` starting at `[Rb]`. `Rb` points after the last one. |
| `LDMIA Rb, { Rlist }` | 1 each     | No     | Loads registers from `Rlist` starting at `[Rb]`. `Rb` points after the last one. |

Note that `POP` with `PC` costs 1 extra cycle (due to prefetch failure).


### Miscellany

| Instruction      | Cycles | Flags? | Meaning      |
| :---             | :---:  | :---   | :---         |
| `POPSP`          | 1      | `----` | `SP := [SP]` |

## Assembler Directives

These directives aim to be compatible with
[DASM](https://github.com/techcompliant/DASM).

### DAT

Writes literal values (numbers and strings).

```
.dat 0xdead, 0xbeef, "also strings"
```

### ORG

Indicates that the following code should be assembled starting at the origin
given to `.org`.

Care must be taken to keep these segments from overlapping. The assembler will
report an error if adjacent segments are too big to fit.


### FILL

Puts a block of repeated data. Accepts two arguments: the value and the amount.

`.fill 0xdead, 20`

writes 20 copies of `0xdead`.

### RESERVE

`.reserve length` is a shorthand for `.fill 0, length`

### DEFINE

`.define` or `.def` (re)defines an assembly-time constant.

`.def symbol, value`

### MACRO

Defines a macro, which has syntax like an instruction.

```
.macro name=...
```

Replacements:

- `%n` is replaced by a newline.
- `%0`, `%1`, etc. are substituted for the arguments
- `%e0`, `%e1`, etc. are replaced with the assembly-time numerical value of the
  corresponding argument.

An example helps clarify:

```
.def foo, 8
.macro m1=.dat %1 %n mov %0, #%1
.macro m2=.dat %e1 %n mov %0, #%e1

m1 r0, 1+1
; .dat 1+1
; mov r0, #1+1

m2 r0, 1+1
; .dat 2
; mov r0, #2
```

Which allows redefinition of symbols in clever ways:

```
.def num0, 5
.def num1, 7
.def num2, 2
.def num3, 12
.def num_counter, 0

.macro dat_num=.dat num%e0
.macro do_num=dat_num num_counter %n .def num_counter, num_counter+1
```

Then

```
do_num
do_num
do_num
```

expands to

```
.dat num0
.def num_counter, num_counter+1
.dat num1
.def num_counter, num_counter+1
.dat num2
.def num_counter, num_counter+1
```

(leaving `num_counter` set to 3).


### ASCIIZ

`.asciiz "str"` is a macro for `.dat "str", 0`.



## Recommendations to Programmers

This section is "non-normative": it is composed of suggestions, not a strict
specification.

- Use `.org` directives to set up the reset and interrupt vectors:
    ```
    ; Reset vector
    b main

    .org 8  ; Interrupt vector
    b interrupt_handler

    .org 0x20 ; Space up to 0x1f is reserved for future vectors.
    ; Rest of the code goes here.
    ```
- Decide on a straightforward calling convention, and use it for all functions.
    One suggestion:
    - Use `r0`, `r1`, `r2`, ... to pass parameters.
    - Use `r0` for return values.
    - `r0`, `r1`, and `r2` can always be clobbered by the callee.
    - `r3` and higher must be preserved **unless** they were used as arguments
        - Arguments can always be clobbered.
- Use `PUSH` and `POP`, including for `LR` and `PC`! They make functions much easier to read.
- Take advantage of the multiple load/store instructions to read the fields of a
  structure. If you know you're going to access several fields, just load them
  all in one instruction.
    - Alternatively, use the immediate-indexing formats.
- Take advantage of post-increment on `LDR` and `STR` for fast loops!
- Remember that conditional branches are fast when they succeed. Each test has
  its inverse, so write loops that succeed N times and fail once.
- Read `SP` with `add r0, sp, #0`. Write `SP` by pushing the new value, then
  `POPSP`.
- Read `PC` with `add r0, pc, #0`. Write `PC` with `BX`.
