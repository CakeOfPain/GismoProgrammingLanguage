.text
.balign 4
.globl _main
_main:
	stp	x29, x30, [sp, -16]!
	mov	x29, sp
	bl	_greetUser
	bl	_showArt
	bl	_showComputedMessage
	mov	w0, #0
	ldp	x29, x30, [sp], 16
	ret
/* end function main */

.text
.balign 4
_greetUser:
	stp	x29, x30, [sp, -16]!
	mov	x29, sp
	adrp	x0, _str0@page
	add	x0, x0, _str0@pageoff
	bl	_puts
	ldp	x29, x30, [sp], 16
	ret
/* end function greetUser */

.text
.balign 4
_showArt:
	stp	x29, x30, [sp, -16]!
	mov	x29, sp
	adrp	x0, _str1@page
	add	x0, x0, _str1@pageoff
	bl	_puts
	adrp	x0, _str2@page
	add	x0, x0, _str2@pageoff
	bl	_puts
	adrp	x0, _str3@page
	add	x0, x0, _str3@pageoff
	bl	_puts
	adrp	x0, _str4@page
	add	x0, x0, _str4@pageoff
	bl	_puts
	ldp	x29, x30, [sp], 16
	ret
/* end function showArt */

.text
.balign 4
_showComputedMessage:
	stp	x29, x30, [sp, -16]!
	mov	x29, sp
	adrp	x0, _str5@page
	add	x0, x0, _str5@pageoff
	bl	_puts
	adrp	x0, _str6@page
	add	x0, x0, _str6@pageoff
	bl	_puts
	adrp	x0, _str7@page
	add	x0, x0, _str7@pageoff
	bl	_puts
	ldp	x29, x30, [sp], 16
	ret
/* end function showComputedMessage */

.data
.balign 8
_str0:
	.ascii "Hello, user! Welcome to this generated code!"
	.byte 0
/* end data */

.data
.balign 8
_str1:
	.ascii "Here is some ASCII art:"
	.byte 0
/* end data */

.data
.balign 8
_str2:
	.ascii "  /\\_/\\"
	.byte 0
/* end data */

.data
.balign 8
_str3:
	.ascii " ( o.o )"
	.byte 0
/* end data */

.data
.balign 8
_str4:
	.ascii "  > ^ <"
	.byte 0
/* end data */

.data
.balign 8
_str5:
	.ascii "Performing some computations:"
	.byte 0
/* end data */

.data
.balign 8
_str6:
	.ascii "42 + 13 = 55"
	.byte 0
/* end data */

.data
.balign 8
_str7:
	.ascii "'Hello' + 'World' = HelloWorld"
	.byte 0
/* end data */

