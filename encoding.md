# Instruction Encoding

Following the usual presentation of ARM7 Thumb, this is presented as a set of
"formats", sometimes comprising several instructions.

Conventions: `ooo` or similar specifies the exact operation, `____` is a
literal values, `sss`, `aaa`, `bbb` and `ddd` are 3-bit register numbers
corresponding to operands (`aaa` and `bbb`), shifts (`sss`) and destinations
(`ddd`). Others are format-specific. `___` is "don't care".

1. `000ooXXXXXsssddd` - Move shifted register
2. `00011IoXXXsssddd` - Add/subtract
3. `001oodddXXXXXXXX` - Move/compare/add/subtract immediate
4. `010000oooosssddd` - ALU operations
5. `010001100L___aaa` - `BX` and `BLX`
6. `010001101oo__ddd` - Hardware (`HWN`, `HWQ` and `HWI`)
7. `0100011100000000` - Manipulating `CPSR`
8. `01001dddXXXXXXXX` - `PC`-relative load
9. `0101LB0aaabbbddd` - Load/store with register offset
10. `0101HS1aaabbbddd` - Unused (formerly Load/Store sign-extended byte/halfword)
11. `011BLXXXXXbbbddd` - Load/store with immediate offset
12. `1000LXXXXXbbbddd` - Unused (formerly load/store halfwords)
13. `1001LdddXXXXXXXX` - `SP`-relative load/store
14. `1010SdddXXXXXXXX` - Load address
15. `10110000SXXXXXXX` - Add offset to `SP`
16. `1011L10Rrrrrrrrr` - Push/pop registers
17. `1100Lbbbrrrrrrrr` - Multiple load/store
18. `1101CCCCXXXXXXXX` - Conditional branch
19. `11011111XXXXXXXX` - Software interrupt
20. `11100XXXXXXXXXXX` - Unconditional branch
21. `11111XXXXXXXXXXX` - Immediate branch and link
22. `11110H__XXXXXXXX` - Long branch and link



### Format 1 - Move shifted register

`000ooXXXXXsssddd`

- `XXXXX` - shift amount (unsigned)
- `sss` - source register
- `ddd` - destination register

| `oo` | Assembly | Meaning |
| :--- | :--- | :--- |
| `00` | `LSL Rd, Rs, #Imm` | Shift `Rs` left by the 5-bit immediate, store in `Rd`. |
| `01` | `LSR Rd, Rs, #Imm` | Shift `Rs` right logically by the 5-bit immediate, store in `Rd`. |
| `10` | `ASR Rd, Rs, #Imm` | Shift `Rs` right arithmetically by the 5-bit immediate, store in `Rd`. |
| `11` | (illegal) | (This is format 2) |

1 cycle. Sets `CPSR` condition codes.

### Format 2 - Add/subtract immediate

`00011Iobbbaaaddd`

- `bbb` - The second operand. When `I` is 0, a register. When `I` is 1, a
  3-bit unsigned literal.
- `aaa` - The first operand, `Ra`.
- `ddd` - The destination register `Rd`.

| `o` | `I` | Assembly | Meaning |
| :---: | :---: | :--- | :--- |
| `0` | `0` | `ADD Rd, Ra, Rb` | `Rd := Ra + Rb` |
| `0` | `1` | `ADD Rd, Ra, #Imm` | `Rd := Ra + Imm` |
| `1` | `0` | `SUB Rd, Ra, Rb` | `Rd := Ra - Rb` |
| `1` | `1` | `SUB Rd, Ra, #Imm` | `Rd := Ra - Imm` |

1 cycle. Sets `CPSR` condition codes.


### Format 3 - Move/compare/add/subtract immediate

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


### Format 4 - ALU operations

`010000oooosssddd`

- `sss` - Source register `Rs`. Never updated.
- `ddd` - Destination register `Rd`. Usually set to a result, but not always.
    See below.

| `oooo` | Assembly | Meaning |
| :---: | :--- | :--- |
| `0000` | `AND Rd, Rs` | `Rd := Rd AND Rs` |
| `0001` | `EOR Rd, Rs` | `Rd := Rd EOR Rs` |
| `0010` | `LSL Rd, Rs` | `Rd := Rd << Rs` |
| `0011` | `LSR Rd, Rs` | `Rd := Rd >> Rs` |
| `0100` | `ASR Rd, Rs` | `Rd := Rd ASR Rs` |
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



### Format 5 - BX and BLX

`010001100L___aaa`

- `L` - 0 = no link, 1 = set `LR` to `PC`
- `aaa` - Target register `Ra`

| `L` | Assembly | Meaning |
| :---: | :--- | :--- |
| `0` | `BX Ra` | `PC := Ra` |
| `1` | `BLX Ra` | `LR := PC` then `PC := Ra` |

1 cycle. Does **not** change `CPSR` flags.


### Format 6 - Hardware

These are for interacting with DCPU-16 hardware devices. There can be up to
65535 (`$ffff`) connected devices, numbered from `$0000` to `$fffe`.

`010001101oo__ddd`

- `ddd` - Destination register `Rd`

| `oo` | Assembly | Meaning |
| :---: | :--- | :--- |
| `00` | `HWN Rd` | Sets `Rd` to the number of connected devices. |
| `01` | `HWQ Rd` | Queries the device whose number is in `Rd`, setting registers as below. |
| `10` | `HWI Rd` | Sends a hardware interrupt to the device whose number is in `Rd` |
| `11` | (illegal) | |

4 cycles, and more if the hardware blocks execution.

Does not change `CPSR` condition codes.



### Format 7 - Manipulating `CPSR`

`01000111ooo_____`

| `ooo` | Assembly | Meaning |
| :---: | :--- | :--- |
| `000` | `RFI` | Returns from an IRQ. (Swaps `SPSR_irq` into `CPSR`, pops `r0` from `SP_irq`.) |
| `001` | `RSI` | Returns from a SWI. (Swaps `SPSR_swi` into `CPSR`, `PC := LR_swi`) |
| `010` | `IFS` | Enable interrupts by setting the `I` flag in `CPSR`. |
| `011` | `IFC` | Disable interrupts by clearing the `I` flag in `CPSR`. |
| `100` | `MRS Rd` | Move `Rd` to the `SPSR` for this mode. (Illegal in User mode.) |
| `101` | `MSR Rd` | Move the `SPSR` for this mode to `Rd`. (Illegal in User mode.) |
| `11_` | (illegal) | (Reserved for the future.) |

1 cycle. Changes `CPSR` as documented above.


### Format 8 - `PC`-relative load

`01001dddXXXXXXXX`

- `ddd` - Destination register `Rd`
- `XXXXXXXX` - 8-bit unsigned offset

`LDR Rd, [PC, #Imm]`

Remember that `PC` points at the next instruction during execution of this one.
(Put another way, `LDR Rd, [PC, #0]` will load the next instruction into `Rd`.)

1 cycle. Does **not** set `CPSR` condition codes.



### Format 9 - Load/store with register offset

`0101LP0aaabbbddd`

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


### Format 10 - Unused (halfword and byte sign-extended load/store)

`0101__1`


### Format 11 - Load/store with immediate offset

`011LPXXXXXbbbddd`

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



### Format 12 - Unused (load/store halfword)

`1000`


### Format 13 - `SP`-relative load/store

`1001LdddXXXXXXXX`

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



### Format 14 - Load address

`1010SdddXXXXXXXX`

- `S` - Source: 0 = `PC`, 1 = `SP`
- `ddd` - Destination register `Rd`
- `XXXXXXXX` - 8-bit unsigned immediate

| `S` | Assembly | Meaning |
| :---: | :--- | :--- |
| `0` | `ADD Rd, PC, #Imm` | Add `Imm` to the current value of `PC`, and load the result into `Rd`. |
| `1` | `ADD Rd, SP, #Imm` | Add `Imm` to the current value of SP, and load the result into `Rd`. |

Note that this puts an address into `Rd`, not the value in memory at that spot!

Note that `PC` points to the instruction after this one.

1 cycle. Does **not** set `CPSR` condition codes.


### Format 15 - Adjust stack pointer

`10110000SXXXXXXX`

- `S` - 0 = add, 1 = subtract
- `XXXXXXX` - 7-bit unsigned immediate

| `S` | Assembly | Meaning |
| :---: | :-- | :-- |
| `0` | `ADD SP, #Imm` | Add `Imm` to `SP` |
| `1` | `SUB SP, #Imm` | Subtract `Imm` from `SP` |

1 cycle. Does **not** set `CPSR` condition codes.


### Format 16 - Push/pop registers

`1011L10Rrrrrrrrr`

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

Note that popping `PC` constitutes a branch.

1 cycle per regular register stored/loaded. 1 extra for storing `LR`, and 2 extra
for loading `PC`.

Does **not** set the `CPSR` condition codes.


### Format 17 - Multiple load/store

`1100Lbbbrrrrrrrr`

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
that `Rb` ends up pointing after the last word written.

1 cycle per register loaded or stored. It is an illegal instruction for
`rrrrrrrr` to be empty.

Does **not** set the `CPSR` condition codes.

### Format 18 - Conditional branches

`1101ccccXXXXXXXX`

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
| `1111` | (illegal) | (This is format 19, software interrupt.) |

Branches take 2 cycles if they fail and 1 cycle to they succeed. (Prefetching
expects conditional branches to succeed.)

Branches examine the `CPSR` condition codes but don't change them.


### Format 19 - Software interrupts

`11011111XXXXXXXX`

`SWI Imm8`

A `SWI` is like a function call to a fixed location.

The immediate field is **ignored** by the processor, it's simply an annotation
or comment the `SWI` handler can use to determine which SWI is requested.

To execute a `SWI`, the processor sets `LR_swi` to the current `PC` (ie. the
instruction after the SWI), saves `CPSR` to `SPSR_swi`, and sets `CPSR` to
`$0011` (flags clear, interrupts disabled, SWI mode). Finally, it sets `PC` to
`$10`, the SWI vector.

Exiting a software interrupt handler is similar to exiting an IRQ: Set the mode
back to what it was before. This is done with `RSI`, which sets `CPSR` to
`SPSR_swi` and `PC` to `LR_swi`.

Note that SWIs share the `SP` and general purpose registers with User mode, and so
must restore them before returning.

Unlike IRQs, however, SWIs happen on demand and at predictable times, so it's
valid and safe for them to consume or return values to/from the user code in the
general-purpose registers and on the stack.

It's safe to nest from an IRQ to a SWI, or vice-versa. SWIs and IRQs are
nestable with extreme care to keep the `SPSR` and `LR` safe.

2 cycles. Does **not** change `CPSR` condition codes, though it does swap the
current `CPSR` for another.


### Format 20 - Unconditional branch

`11100XXXXXXXXXXX`

- `XXXXXXXXXXX` is an 11-bit 2's complement signed offset.

`B label` - Branches unconditionally to `PC + offset`.

Remember that `PC` points to the instruction after this one.

Does not set `LR` or the `CPSR` condition codes.



### Format 21 - Immediate branch and link

`11111XXXXXXXXXXX`

- `XXXXXXXXXXX` - 11-bit unsigned absolute address

`BL label` where the target address is small enough to fit into 11 bits gets
assembled like this. `LR` is set to `PC`, and `PC` is set to the immediate value.

1 cycle, does **not** set `CPSR` condition codes.


### Format 22 - Long branch and link

`11110H__XXXXXXXX`

When the absolute address is too big for 11 bits (and therefore can't use format
21), this two-instruction form is used.

The first instruction has `H=0` and gives the high 8 bits. Those 8 bits are
written into `LR` and shifted up by 8.

The second instruction has `H=1` and gives the low 8 bits. Those 8 bits are
ORed with `LR`, given the full 16-bit address in `LR`.

Then `LR` is exchanged with `PC`, effectively performing a branch-and-link to
the absolute address.

The assembly is `BL label`, and the assembler will choose format 21 or 22 based
on the size of the absolute address.

Each part of this two-part instruction takes 1 cycle, for a total of 2 cycles.

(Note that it's not obvious from the assembly code whether a `BL label` takes 1
cycle or 2. If that's important to your code, you'll have to work out which form
it is based on the value of `label`, or hand-assemble the two-instruction form.)


Does not change `CPSR` condition codes.

