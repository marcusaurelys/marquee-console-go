package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"golang.org/x/sys/windows"
)

func disableEcho() {
	// Get the handle to standard input
	handle := windows.Handle(windows.Stdin)

	// Get current console mode
	var originalMode uint32
	err := windows.GetConsoleMode(handle, &originalMode)
	if err != nil {
		panic(err)
	}

	// Disable echo input
	newMode := originalMode &^ (windows.ENABLE_ECHO_INPUT | windows.ENABLE_LINE_INPUT)
	err = windows.SetConsoleMode(handle, newMode)
	if err != nil {
		panic(err)
	}
}

func setCursor() {
	handle := windows.Handle(os.Stdout.Fd())

	// Get console screen buffer info
	var info windows.ConsoleScreenBufferInfo
	err := windows.GetConsoleScreenBufferInfo(handle, &info)
	if err != nil {
		panic(err)
	}

	// Move cursor to bottom right
	columns := info.Size.X
	rows := info.Size.Y
	pos := windows.Coord{
		X: columns - 1,
		Y: rows - 1,
	}

	err = windows.SetConsoleCursorPosition(handle, pos)
	if err != nil {
		panic(err)
	}
}

func clear() {
	cmd := exec.Command("cmd", "/c", "cls") //Windows example, its tested
	cmd.Stdout = os.Stdout
	cmd.Run()
}

func repaint(terminalState *string, input []string, keyhit chan string) {
	for {

		fmt.Print(*terminalState + "\n~" + strings.Join(input, "\n~"))
		select {

		case val := <-keyhit:

			if val == "\r" {
				if input[len(input)-1] == "clear" {
					input = input[:0]
				}
				input = append(input, "")
			} else if val == "\b" && len(input[len(input)-1]) != 0 {
				in := input[len(input)-1]
				input[len(input)-1] = in[:len(input[len(input)-1])-1]
			} else {
				input[len(input)-1] += val
			}
			clear()

		case <-time.After(500 * time.Millisecond):
			clear()
		}

	}
}

func updateTerminalHeader(start *int, terminalState *string, text string) {
	for {
		padded := strings.Repeat(" ", *start) + text
		*terminalState = padded
		time.Sleep(500 * time.Millisecond)
		*start++
	}

}

func inputReader(input chan string) {
	reader := bufio.NewReader(os.Stdin)
	var buf [1]byte
	for {
		n, err := reader.Read(buf[:])
		if err != nil {
			continue
		}
		input <- string(buf[:n])
	}

}

func main() {
	disableEcho()
	headerState := "CSOPESY"
	input := []string{""}
	clear()

	keyhit := make(chan string)

	start := 0

	go updateTerminalHeader(&start, &headerState, "CSOPESY")
	go inputReader(keyhit)
	go repaint(&headerState, input, keyhit)
	//fmt.Println("Type something (you have 5 seconds):")

	<-time.After(20 * time.Second)
	fmt.Println("\ninput!")
	fmt.Println(strings.Join(input, "\n"))

}
