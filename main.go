package main

import (
	"bufio"
	"fmt"
	"os"
	"syscall"
	"unsafe"
)

type Tty struct {
	In      *os.File
	Out     *os.File
	Buf     *bufio.Reader
	Restore syscall.Termios
}

func SetTermios(fd uintptr, termios syscall.Termios) error {
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd), uintptr(syscall.TCSETS), uintptr(unsafe.Pointer(&termios)))
	if errno != 0 {
		return os.NewSyscallError("SYS_IOCTL", errno)
	}

	return nil
}

func GetTermios(fd uintptr) (syscall.Termios, error) {
	var termios syscall.Termios
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd), uintptr(syscall.TCGETS), uintptr(unsafe.Pointer(&termios)))
	if errno != 0 {
		return termios, os.NewSyscallError("SYS_IOCTL", errno)
	}

	return termios, nil
}

func GetTty() (Tty, error) {
	var tty Tty
	in, err := os.OpenFile("/dev/tty", os.O_RDONLY, 0)
	if err != nil {
		return tty, err
	}

	out, err := os.OpenFile("/dev/tty", os.O_WRONLY, 0)
	if err != nil {
		return tty, err
	}

	termios, err := GetTermios(in.Fd())
	if err != nil {
		return tty, err
	}
	tty.Restore = termios

	termios.Iflag &^= syscall.IGNBRK | syscall.BRKINT | syscall.PARMRK |
		syscall.ISTRIP | syscall.INLCR | syscall.IGNCR |
		syscall.ICRNL | syscall.IXON
	termios.Lflag &^= syscall.ECHO | syscall.ECHONL | syscall.ICANON |
		syscall.ISIG | syscall.IEXTEN
	termios.Oflag &^= syscall.OPOST
	termios.Cc[syscall.VMIN] = 1
	termios.Cc[syscall.VTIME] = 0

	if err := SetTermios(in.Fd(), termios); err != nil {
		return tty, err
	}

	tty.In = in
	tty.Out = out
	tty.Buf = bufio.NewReader(in)
	return tty, nil
}

func main() {
	tty, err := GetTty()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	for {
		r, _, _ := tty.Buf.ReadRune()
		if r == 'q' {
			SetTermios(tty.In.Fd(), tty.Restore)
			os.Exit(0)
		}
		fmt.Printf("rune: %c", r)
	}
}
