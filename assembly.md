# Assembler

This is a guide to using the assembler to produce code for the Risque-16.

## Directives

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



## Labels

Labels are defined with a trailing colon:

```
foo:
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

Immediate values are given as `[US]length`, where `U7` means unsigned 7-bit, and
`S5` means signed (2's complement) 5-bit.

### Arithmetic and Moves

| Instruction | Cycles | Flags? | Meaning |
| :--- | :---: | :--- | :--- |
| `ADC Rd, Rs` | 1 | Yes | `Rd := Rd + Rs + C-bit` |
| `ADD Rd, #U8` | 1 | Yes | `Rd := Rd + U8` |
| `ADD Rd, PC, #U8` | 1 | **No** | `Rd := PC + U8`, `PC` unchanged. |
| `ADD Rd, Ra, #U3` | 1 | Yes | `Rd := Ra + U3` |
| `ADD Rd, Ra, Rb` | 1 | Yes | `Rd := Ra + Rb` |
| `ADD Rd, SP, #U8` | 1 | **No** | `Rd := SP + U8`, `SP` unchanged. |
| `ADD SP, #U7` | 1 | No | `SP := SP + U7` |
| `AND Rd, Rs` | 1 | Yes | `Rd := Rd AND Rs` |
| `ASR Rd, Rs, #U5` | 1 | Yes | Shift `Rs` right arithmetically by `U5`, store in `Rd`. |
| `ASR Rd, Rs` | 1 | Yes | `Rd := Rd ASR Rs` |
| `BIC Rd, Rs` | 1 | Yes | `Rd := Rd AND NOT Rs` (**bi**t **c**lear) |
| `CMN Rd, Rs` | 1 | Yes | Set condition codes based on `Rd + Rs` |
| `CMP Rd, #U8` | 1 | Yes | Set condition codes based on `Rd - U8` |
| `CMP Rd, Rs` | 1 | Yes | Set condition codes based on `Rd - Rs` |
| `EOR Rd, Rs` | 1 | Yes | `Rd := Rd EOR Rs` |
| `LSL Rd, Rs, #U5` | 1 | Yes | Shift `Rs` left by `U5`, store in `Rd`. |
| `LSL Rd, Rs` | 1 | Yes | `Rd := Rd << Rs` |
| `LSR Rd, Rs, #U5` | 1 | Yes | Shift `Rs` right logically by `U5`, store in `Rd`. |
| `LSR Rd, Rs` | 1 | Yes | `Rd := Rd >> Rs` |
| `MOV Rd, #U8` | 1 | Yes | `Rd := U8` |
| `MUL Rd, Rs` | 4 | Yes | `Rd := Rd * Rs` |
| `MVN Rd, Rs` | 1 | Yes | `Rd := NOT Rs` (bitwise negate, not integer negative) |
| `NEG Rd, Rs` | 1 | Yes | `Rd := -Rs` (arithmetic negation) |
| `ORR Rd, Rs` | 1 | Yes | `Rd := Rd OR Rs` |
| `ROR Rd, Rs` | 1 | Yes | `Rd := Rd ROR Rs` (rotate `Rd`'s bits rightward by `Rs`) |
| `SBC Rd, Rs` | 1 | Yes | `Rd := Rd - Rs - NOT C-bit` |
| `SUB Rd, #U8` | 1 | Yes | `Rd := Rd - U8` |
| `SUB Rd, Ra, #U3` | 1 | Yes | `Rd := Ra - U3` |
| `SUB Rd, Ra, Rb` | 1 | Yes | `Rd := Ra - Rb` |
| `SUB SP, #U7` | 1 | No | `SP := SP - U7` |
| `TST Rd, Rs` | 1 | Yes | Set condition codes based on `Rd AND Rs` |


### Unconditional Branches

| Instruction | Cycles | Flags? | Branch when |
| :--- | :---: | :--- | :--- |
| `B label` | 1 | No | `PC := PC + offset` (Range is 11-bit signed, +1023 to -1024) |
| `BL label` | 1 | No | `PC := label` (Unlimited range, maybe 2 instructions.) |
| `BX Ra` | 1 | No | `PC := Ra` |
| `BLX Ra` | 1 | No | `LR := PC` then `PC := Ra` |


### Conditional Branches

These are all PC-relative branches, by an 8-bit signed offset (+127 to -128).
Therefore they're for local labels, loops and so on, not for long jumps and
function calls.

None of these change the condition flags. Each takes 1 cycle on success, and 2
on failure.

| Instruction | Cycles | Branch when |
| :--- | :---: | :--- |
| `BEQ label` | 1/2 | `Z` set (equal) |
| `BNE label` | 1/2 | `Z` clear (not equal) |
| `BCS label` | 1/2 | `C` set (unsigned higher or same) |
| `BCC label` | 1/2 | `C` clear (unsigned lower) |
| `BMI label` | 1/2 | `N` set (negative) |
| `BPL label` | 1/2 | `N` clear (positive or zero) |
| `BVS label` | 1/2 | `V` set (overflow) |
| `BVC label` | 1/2 | `V` clear (no overflow) |
| `BHI label` | 1/2 | `C` set and `Z` clear (unsigned higher) |
| `BLS label` | 1/2 | `C` clear or `Z` set (unsigned lower or same) |
| `BGE label` | 1/2 | `N` and `V` match (signed greater or equal) |
| `BLT label` | 1/2 | `N` and `V` differ (signed less than) |
| `BGT label` | 1/2 | `Z` clear, and `N` and `V` match (signed greater than) |
| `BLE label` | 1/2 | `Z` set, or `N` and `V` differ (signed less than or equal) |

### Hardware

| Instruction | Cycles | Meaning |
| :--- | :---: | :--- |
| `HWN Rd` | 4 | `Rd := # of connected devices` |
| `HWQ Rd` | 4 | Sets `r0` - `r4` to the device info for device `Rd`. |
| | | (`r1:r0` = ID, `r2` = version, `r4:r3` = manufacturer) |
| `HWI Rd` | 4 | Sends a hardware interrupt to device `Rd`. |

### Interrupts

| Instruction | Cycles | Flags? | Meaning |
| :--- | :---: | :--- | :--- |
| `SWI #U8` | 2 | No | Triggers a `SWI` with code `U8`. |
| `RFI` | 1 | Special | Returns from an IRQ: `CPSR := SPSR_irq`, pop `r0` from `SP_irq`. |
| `RSI` | 1 | Special | Returns from a SWI: `CPSR := SPSR_swi`, `PC := LR_swi` |

### Status Register

| Instruction | Cycles | Flags? | Meaning |
| :--- | :---: | :--- | :--- |
| `IFS` | 1 | Special | Sets `I` in `CPSR`; doesn't change condition codes. |
| `IFC` | 1 | Special | Clears `I` in `CPSR`; doesn't change condition codes. |
| `MRS Rd` | 1 | No | Moves `Rd` to the `SPSR` for the current mode. Illegal in User mode! |
| `MSR Rd` | 1 | No | Moves the `SPSR` for the current mode to `Rd`. Illegal in User mode! |

### Load/Store

| Instruction | Cycles | Flags? | Meaning |
| :--- | :---: | :--- | :--- |
| `LDR Rd, [Rb, #U5]` | 1 | No | `Rd := [Rb + U5]`, `Rb` unchanged. |
| `LDR Rd, [Rb], #U5]` | 1 | No | `Rd := [Rb]`, `Rb := Rb + U5` |
| `LDR Rd, [Rb, Ra]` | 1 | No | `Rd := [Rb + Ra]`, `Rb` unchanged. |
| `LDR Rd, [Rb], Ra` | 1 | No | `Rd := [Rb]`, `Rb := Rb + Ra` |
| `STR Rd, [Rb, #U5]` | 1 | No | `[Rb + U5] := Rd`, `Rb` unchanged. |
| `STR Rd, [Rb], #U5]` | 1 | No | `[Rb] := Rd`, `Rb := Rb + U5` |
| `STR Rd, [Rb, Ra]` | 1 | No | `[Rb + Ra] := Rd`, `Rb` unchanged. |
| `STR Rd, [Rb], Ra` | 1 | No | `[Rb] := Rd`, `Rb := Rb + Ra` |
| `LDR Rd, [PC, #U8]` | 1 | No | `Rd := [PC + U8]`, `PC` unchanged. |
| `LDR Rd, [SP, #U8]` | 1 | No | `Rd := [SP + U8]`, `SP` unchanged. |
| `STR Rd, [SP, #U8]` | 1 | No | `[SP + U8] := Rd`, `SP` unchanged. |

### Multiple Load/Store

`Rlist` is a comma-separated list of general-purpose registers (`r0` to `r7`).

Their order in the list is irrelevant; they always get stored in ascending order.

| Instruction | Cycles | Flags? | Meaning |
| :--- | :---: | :--- | :--- |
| `PUSH { Rlist }` | 1 each | No | Writes registers ascending in memory, into the stack. |
| `PUSH { Rlist, LR }` | 1 each | No | Writes registers ascending in memory, into the stack. |
| `POP { Rlist }` | 1 each | No | Loads registers from the stack. |
| `POP { Rlist, PC }` | 1 + 1 each | No | Loads registers from the stack, including PC |
| `STMIA Rb!, { Rlist }` | 1 each | No | Store registers from `Rlist` starting at the address in `Rb`. Moves `Rb` to after the last one. |
| `LDMIA Rb!, { Rlist }` | 1 each | No | Loads registers from `Rlist` starting at the address in `Rb`. Moves `Rb` to after the last one. |





