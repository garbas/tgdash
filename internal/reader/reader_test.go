package reader

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func TestReadLines(t *testing.T) {
	input := "line1\nline2\nline3\n"
	r := strings.NewReader(input)

	msgs := make([]tea.Msg, 0)
	ch := make(chan tea.Msg, 10)

	go func() {
		ReadLines(r, func(msg tea.Msg) {
			ch <- msg
		})
		close(ch)
	}()

	for msg := range ch {
		msgs = append(msgs, msg)
	}

	// 3 LineMsg + 1 InputDoneMsg
	lineCount := 0
	doneCount := 0
	for _, msg := range msgs {
		switch msg.(type) {
		case LineMsg:
			lineCount++
		case InputDoneMsg:
			doneCount++
		}
	}

	if lineCount != 3 {
		t.Fatalf("expected 3 LineMsg, got %d",
			lineCount)
	}
	if doneCount != 1 {
		t.Fatalf("expected 1 InputDoneMsg, got %d",
			doneCount)
	}
}

func TestReadLinesContent(t *testing.T) {
	input := "[vpc] hello\n[rds] world\n"
	r := strings.NewReader(input)

	var lines []string
	ch := make(chan tea.Msg, 10)

	go func() {
		ReadLines(r, func(msg tea.Msg) {
			ch <- msg
		})
		close(ch)
	}()

	timeout := time.After(2 * time.Second)
	for {
		select {
		case msg, ok := <-ch:
			if !ok {
				goto done
			}
			if lm, ok := msg.(LineMsg); ok {
				lines = append(lines, lm.Raw)
			}
		case <-timeout:
			t.Fatal("timeout waiting for messages")
		}
	}
done:
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(lines))
	}
	if lines[0] != "[vpc] hello" {
		t.Fatalf("expected [vpc] hello, got %s",
			lines[0])
	}
}
