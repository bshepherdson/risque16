# Risque-16 CPU Architecture

This is a RISC architecture intended as an in-universe competitor to the
DCPU-16. It can interface with the same hardware devices as the DCPU, and is of
a similar speed and power.

The Risque-16 architecture and instruction set is a modified version of the
real-world Thumb architecture supported as a variant of ARM.

## Overview

The Risque-16 is a 16-bit, word-addressing RISC architecture CPU. The clock
speed is 100kHz, and most instructions require 1 cycle to execute.

There are 8 16-bit general-purpose registers, named `r0` to `r7`. There are
four more special-purpose registers (some of them with doubles for other modes):

- `PC` is the program counter, which points at the next (not current)
  instruction to execute.
- `SP` is a stack pointer, usually used for first-in, first-out data storage.
- `LR` is the "link register", which is set by branch-and-link opcodes to
  enable returning to call sites.
- `CPSR` is the current program status register. It contains flags relating to
  the current state of the processor (see below).

Each of these also has an interrupt-mode double, referred to here respectively
as `PC_irq`, `SP_irq`, `LR_irq` and `SPSR_irq`.

There is an additional `LR_swi` and `SPSR_swi` used during software interrupts.

See below for the details of interrupt handling.

`CPSR` has the form `NZCV____ I__mmmmm`

| Flag | Name | Meaning |
| :---: | :--- | :--- |
| `N` | Negative | Set to bit 15 of results, so `N` is 1 if the signed value is negative. |
| `Z` | Zero | Set if the result is zero (often denotes equality in a comparison). |
| `C` | Carry | More complicated. See below. |
| `V` | Overflow | Set is a signed overflow occurred. Otherwise left alone. |
| `I` | Interrupt Enable | Set to permit IRQs from hardware, clear to disallow them. |

Processor modes:

| `mmmmm` | Name | Vector | Meaning |
| :---: | :--- | :---: | :--- |
| `10000` | User | `$00` | The main mode of execution. |
| `10001` | SWI | `$08` | While handling software interrupts. |
| `10010` | IRQ | `$10` | While handling hardware interrupts. |
| `11000` | Abort | `$18` | Signaling a well-defined error state like illegal opcode) |
| `11100` | Undefined | `$20` | Signaling something has gone very wrong. |

Carry has some tricky interactions, summarized here:
- For `ADC`, `ADD` and `CMN`, set if there's an *unsigned* overflow.
- For `CMP`, `SBC` and `SUB`, set if the result is an unsigned *underflow*.
- For shifting instructions, set to the last bit shifted out.
- Others usually leave this flag alone.

## Interacting with Hardware

Risque-16 is compatible with the same hardware as the DCPU-16.

DCPU-16 registers `A`, `B`, `C`, `X`, `Y`, `Z`, `I` and `J` correspond to
Risque-16 `r0-r7` in that order.

There are corresponding instructions to query the number of hardware devices,
and collect information about the hardware.

There is an interrupt queue like the DCPU-16, holding a maximum of 256
interrupts. While interrupts are disabled (`CPSR` `I` is clear), new interrupts
are added to the queue. Once interrupts are re-enabled, interrupts will fire
from the queue until it empties.

There is no guarantee of forward progress in the "real", non-interrupt program.
The same instruction can be repeatedly interrupted, forever.

Interrupts can be nested, with care. A second incoming interrupt would destroy
the values of the `PC_irq`, `LR_irq` and `SPSR_irq`. But you can contrive to
save and restore all of those values from memory, and thereby nest interrupts
safely.

### Interrupt Handling

When an interrupt fires (that is, leaves the queue and is being handled), the
Risque-16 does the following:

- Sets the mode to `10010`, signaling interrupts. This swaps in the `*_irq`
  registers, preserving the original user mode ones.
- Sets `PC_irq` to the interrupt vector: `0x0008`.
- Pushes `r0` to the `SP_irq` stack, and then sets `r0` to the interrupt
  message.

Then your interrupt handler can examine `r0`, respond to the interrupt, and
then return from interrupts with `RFI`.


## Vectors and Reserved Space

There are five vectors on the Risque-16:

- `$0000` is the reset vector, loaded at startup.
- `$0008` is the IRQ vector, when handling interrupts.
- `$0010` is the SWI vector, called by software interrupts.
- `$0018` is the Abort vector, called on error states.
- `$0020` is the Undefined vector, called when crazy things happen.

Memory up to `$0030` is reserved; your code should start at or after `$0030`.

Since there's not a lot of space in any of them, it's expected that they'll
contain either an immediate return, or a jump to more complex code.

## Startup State

On a reset, the processor performs the following operations:

- All general-purpose registers are set to 0.
- `CPSR`, `SPSR_irq` and `SPSR_swi` have interrupts disabled (`I=0`) and all
  condition flags cleared.
- `PC` starts at `$0000`, the reset vector.
- `SP` starts at 0 as well. Since the stack is full-descending, that signals an
  empty stack. The first value goes in `$ffff` (if you don't change `SP`.)
- The processor mode is User mode.



## Differences from Thumb

- 16-bit addressing is used everywhere - no byte addressing.
- Registers and words are 16-bit, not 32-bit.
- There are no "high" registers, just `r0-7`, `sp`, `pc` and `lr`.
- Modes and interrupts are different.
- Format 21 (long branch and link) is split into two, since with only 16-bit
  addressing, many absolute addresses will fit into 11 bits.
    - Format 21 is now a single-instruction branch-and-link to an absolute
      11-bit address, while format 22 is a modified form of Thumb's
      two-instruction format 19.
- Thumb's format 5 is repurposed into my formats 5, 6, and 7.
- Thumb's formats 8 and 10 (my formats 10 and 12) are unused.

### Available instruction set space

Here are the gaps in the instruction set where new things can go:

- `0101__1_________` - format 10 (originally 8)
- `10000___________` - format 12 (originally 10)


## Timing and Pipelining

The processor and memory run at the same pace, allowing minimal delays for
memory operations.

The processor accordingly has a short pipeline. During the execution of one
instruction, `PC` points at the next instruction.

Most instructions take 1 or 2 cycles, as noted in their descriptions.

