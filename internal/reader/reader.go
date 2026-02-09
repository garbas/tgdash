package reader

import (
	"bufio"
	"io"

	tea "github.com/charmbracelet/bubbletea"
)

type LineMsg struct {
	Raw string
}

type InputDoneMsg struct{}

func ReadLines(r io.Reader, send func(tea.Msg)) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		send(LineMsg{Raw: scanner.Text()})
	}
	send(InputDoneMsg{})
}
