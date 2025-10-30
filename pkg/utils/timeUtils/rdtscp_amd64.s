//go:build amd64

#include "textflag.h"

// func Rdtscp() (tsc uint64, aux uint32)
TEXT ·Rdtscp(SB), NOSPLIT, $0-12
    RDTSCP                 // DX:AX = TSC, CX = AUX
    SHLQ $32, DX           // DX <<= 32
    ORQ  AX, DX            // DX |= AX
    MOVQ DX, tsc+0(FP)     // write back tsc
    MOVL CX, aux+8(FP)     // write back aux
    RET
