# Risque-16 CPU Architecture

This is a RISC architecture intended as an in-universe competitor to the
DCPU-16. It can interface with the same hardware devices as the DCPU, and is of
a similar speed and power.

The Risque-16 architecture and instruction set is inspired by the real-world
Thumb architecture, the 16-bit variant of ARM.

## Overview

The Risque-16 is a 16-bit, word-addressing RISC architecture CPU. The clock
speed is 200kHz, and most instructions require 1 cycle to execute. Since each
cycle generally does less than a DCPU cycle, the overall pace of operation is
comparable.

There are 8 16-bit general-purpose registers, named `r0` to `r7`. There are
five more special-purpose registers:

- `PC` is the program counter, which points at the next (not current)
  instruction to execute.
- `SP` is a stack pointer, usually used for first-in, first-out data storage.
- `LR` is the "link register", which is set by branch-and-link opcodes to
  enable returning to call sites.
- `CPSR` is the current program status register. It contains flags relating to
  the current state of the processor (see below).
- `SPSR` is the saved program status register. `CPSR` is copied here on an
  interrupt, and restored when returning from an interrupt.

See below for the details of interrupt handling.

`CPSR` has the form `________ I___NZCV`

| Flag  | Name             | Meaning                                                                |
| :---: | :---             | :---                                                                   |
| `I`   | Interrupt Enable | Set to permit interrupts to fire, clear to disallow them.              |
| `N`   | Negative         | Set to bit 15 of results, so `N` is 1 if the signed value is negative. |
| `Z`   | Zero             | Set if the result is zero (often denotes equality in a comparison).    |
| `C`   | Carry            | More complicated. See below.                                           |
| `V`   | Overflow         | Set is a signed overflow occurred. Otherwise left alone.               |

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
interrupts. While interrupts are disabled (status bit `I` is clear), new
interrupts are added to the queue. Once interrupts are re-enabled, interrupts
will fire from the queue until it empties.

Hardware IRQs and software interrupts triggered by `SWI` both go in the same
queue.

There is no guarantee of forward progress in the "real", non-interrupt program.
The same instruction can be repeatedly interrupted, forever.

Interrupts can be nested, with care. `SPSR` would get overwritten by the second
interrupt, but you can save and restore it using `XSR`.

### Interrupt Handling

When an interrupt fires (that is, leaves the queue and is being handled), the
Risque-16 does the following:

- Copies `CPSR` to `SPSR`.
- Writes 0 into `CPSR`, clearing all status and disabling interrupts.
- Pushes `PC` to the stack.
- Pushes `r0` to the stack.
- Sets `r0` to the interrupt message.
- Sets `PC` to the interrupt vector: `0x0008`

Then your interrupt handler can examine `r0`, respond to the interrupt, and
then return from the interrupt with `RFI`.


## Vectors and Reserved Space

There are two "vectors" on the Risque-16:

- `$0000` is the reset vector, loaded at startup.
- `$0008` is the IRQ vector, when handling interrupts.

Memory up to `$0020` is reserved; your code should start at or after `$0020`.

Since there's not a lot of space in these vectors, it's expected that they'll
contain either an immediate return, or a jump to more complex code.

## Startup State

On a reset, the processor performs the following operations:

- All general-purpose registers are set to 0.
- `CPSR` and `SPSR` start with all bits cleared (interrupts disabled).
- `PC` starts at `$0000`, the reset vector.
- `SP` starts at 0 as well. Since the stack is full-descending, that signals an
  empty stack. The first value goes in `$ffff` (if you don't change `SP`.)



## Differences from Thumb

The instruction set and functionality is inspired by the ARM architecture, and
especially the 16-bit Thumb subset.

- Heavily reworked instruction encoding.
    - This is slightly less efficient (eg. fewer bits available for literals)
    - But it's also much, much simpler - 5 formats with clear layout, instead of
      20 with a confused tangle.
- 16-bit addressing is used everywhere - no byte addressing.
- Registers and words are 16-bit, not 32-bit.
- There are no "high" registers, just `r0-7`, `sp`, `pc` and `lr`.
- No CPU modes, and interrupt handling has been made much simpler and
  DCPU-compatible.


## Timing and Pipelining

The processor and memory run at the same pace, allowing minimal delays for
memory operations.

The processor accordingly has a short pipeline. During the execution of one
instruction, `PC` points at the next instruction.

Most instructions take 1 or 2 cycles, as noted in their descriptions. `MUL`
takes 4, as do several that work with interrupts or hardware.

