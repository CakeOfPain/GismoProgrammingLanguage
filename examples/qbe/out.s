.text
.balign 4
.globl _main
_main:
	stp	x29, x30, [sp, -16]!
	mov	x29, sp
	adrp	x0, _str0@page
	add	x0, x0, _str0@pageoff
	bl	_puts
	mov	w0, #0
	ldp	x29, x30, [sp], 16
	ret
/* end function main */

.data
.balign 8
_str0:
	.ascii "Hello World!"
	.int 0
/* end data */

