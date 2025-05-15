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

type TerminalState struct {
	headerState string
	inputState  []string
	keyhit      chan string
	refreshRate int
	startPad    int
	text        string
}

func (ts *TerminalState) updateTerminalHeader() {
	terminalState := &ts.headerState
	start := &ts.startPad
	text := ts.text
	for {
		padded := strings.Repeat(" ", *start) + text
		*terminalState = padded
		time.Sleep(time.Duration(ts.refreshRate) * time.Millisecond)
		*start++
	}

}

func (ts *TerminalState) repaint() {
	input := ts.inputState
	terminalState := &ts.headerState
	keyhit := ts.keyhit

	for {

		fmt.Print(*terminalState + "\n~" + strings.Join(input, "\n~"))
		select {

		case val := <-keyhit:

			if val == "\r" {
				if input[len(input)-1] == "clear" {
					input = make([]string, 0)
				}
				input = append(input, "")
			} else if val == "\b" && len(input[len(input)-1]) >= 1 {
				in := input[len(input)-1]
				input[len(input)-1] = in[:len(input[len(input)-1])-1]
				setCursor()
			} else {
				input[len(input)-1] += val
			}
			clear()

		case <-time.After(time.Duration(ts.refreshRate) * time.Millisecond):
			clear()
		}

	}
}

func (ts *TerminalState) setInputReader() {
	input := ts.keyhit
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

func main() {
	disableEcho()
	clear()

	ts := TerminalState{"CSOPESY", []string{""}, make(chan string), 100, 0, "CSOPESY"}

	go ts.updateTerminalHeader()
	go ts.setInputReader()
	go ts.repaint()
	//fmt.Println("Type something (you have 5 seconds):")

	<-time.After(20 * time.Second)
	fmt.Println("\ninput!")

}
