.text
.balign 4
_initVector:
	stp	x29, x30, [sp, -16]!
	mov	x29, sp
	str	w1, [x0]
	mov	x1, #4
	add	x0, x0, x1
	str	w2, [x0]
	ldp	x29, x30, [sp], 16
	ret
/* end function initVector */

.text
.balign 4
_VectorGetX:
	stp	x29, x30, [sp, -16]!
	mov	x29, sp
	ldr	w0, [x0]
	ldp	x29, x30, [sp], 16
	ret
/* end function VectorGetX */

.text
.balign 4
_VectorGetY:
	stp	x29, x30, [sp, -16]!
	mov	x29, sp
	mov	x1, #4
	add	x0, x0, x1
	ldr	w0, [x0]
	ldp	x29, x30, [sp], 16
	ret
/* end function VectorGetY */

.text
.balign 4
.globl _main
_main:
	stp	x29, x30, [sp, -48]!
	mov	x29, sp
	str	x19, [x29, 40]
	mov	w2, #20
	mov	w1, #10
	add	x0, x29, #28
	bl	_initVector
	add	x0, x29, #28
	bl	_VectorGetX
	mov	w19, w0
	add	x0, x29, #28
	bl	_VectorGetY
	mov	x1, #16
	sub	sp, sp, x1
	mov	x1, #8
	add	x1, sp, x1
	str	w0, [x1]
	mov	x0, #0
	add	x0, sp, x0
	str	w19, [x0]
	adrp	x0, _fmt@page
	add	x0, x0, _fmt@pageoff
	bl	_printf
	mov	x0, #16
	add	sp, sp, x0
	mov	w0, #0
	ldr	x19, [x29, 40]
	ldp	x29, x30, [sp], 48
	ret
/* end function main */

.data
.balign 8
_fmt:
	.ascii "x is %d\ny is %d\n"
	.byte 0
/* end data */

