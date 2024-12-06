.text
.balign 4
.globl _main
_main:
	stp	x29, x30, [sp, -16]!
	mov	x29, sp
	mov	x0, #16
	sub	sp, sp, x0
	mov	x0, #0
	add	x1, sp, x0
	mov	w0, #31416
	str	w0, [x1]
	adrp	x0, _fmt@page
	add	x0, x0, _fmt@pageoff
	bl	_printf
	mov	x0, #16
	add	sp, sp, x0
	mov	w0, #0
	ldp	x29, x30, [sp], 16
	ret
/* end function main */

.data
.balign 8
_fmt:
	.ascii "Hello World! %d\n"
	.byte 0
/* end data */

