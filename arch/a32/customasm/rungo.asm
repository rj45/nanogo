
; code bank is the main program memory bank, this is presumed
; to be in RAM and writable. So pre-initialized global variables
; are also put here, as well as read-only strings.
#bankdef code
{
  #addr 0x0
  #size 0x1000
  ; todo: change to ram address once the emulator supports this
  ; #addr 0x01000000
  ; #size 0x00100000
  #outp 0
}

; bss is the main data memory bank for uninitialized variables
; this bank is not stored in the output file, and in hardware
; it should be zeroed out on initialization
#bankdef bss
{
  #addr 0x01000000
  #size 0x00100000
  ; #addr 0x01100000
  ; #size 0x00800000
}

#bank code


stackStartAddress = 0x01FFFFFC

; run go's main__main function
init:
  ; initialize the stack
  ld   sp, stackStartAddress

  ; TODO: add code to zero out the bss area

  ; initialize all the global variables
  call  main__init

  ; check that the stack is not corrupted
  cmp   sp, stackStartAddress
  br.eq .stackok
  brk
  err
  bra   .looperr
.stackok:

  ; run the main program
  call   main__main

  ; check that the stack is not corrupted
  cmp    sp, stackStartAddress
  br.eq  .stackok2
  brk
  err
  bra    .looperr
.stackok2:

  ; halt or loop forever
  hlt

.loophalt:
  bra    .loophalt

.looperr:
  ; if the ERR instruction does nothing, then the fact
  ; it's looping here can be used to see if stack
  ; corruption happened
  bra    .looperr

#bank code
