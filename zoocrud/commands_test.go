package zoocrud

import (
	"errors"
	"strings"
	"testing"
)

func TestExecuteSession(t *testing.T) {
	store := NewMemoryStore()
	steps := []struct {
		line string
		want string
		err  error
	}{
		{"CREATE /zoo", "Created /zoo", nil},
		{"create /zoo/lion the king", "Created /zoo/lion", nil},
		{"READ /zoo/lion", `"the king" (version 0, 0 children)`, nil},
		{"UPDATE /zoo/lion  still   the king", "Updated /zoo/lion to version 1", nil},
		{"READ /zoo/lion", `"still   the king" (version 1, 0 children)`, nil},
		{"UPDATE /zoo/lion again", "Updated /zoo/lion to version 2", nil},
		{"EXISTS /zoo/tiger", "/zoo/tiger does not exist", nil},
		{"CREATE /zoo/tiger", "Created /zoo/tiger", nil},
		{"LIST /zoo", "lion\ntiger", nil},
		{"DELETE /zoo", "Error: /zoo has children. Use DELETEC to delete it with its children.", ErrNotEmpty},
		{"DELETEC /zoo", "Deleted /zoo and 2 descendants", nil},
		{"LIST /", "/ has no children", nil},
		{"READ /zoo", "Error: node does not exist", ErrNoNode},
		{"CREATE zoo", `Error: invalid path "zoo": it must start with / and must not end with /`, ErrBadPath},
		{"READX /zoo", `Unknown command "READX". Type HELP to see the commands.`, ErrUsage},
		{"READ", "Usage: READ <path>", ErrUsage},
		{"UPDATE /zoo", "Usage: UPDATE <path> <value>", ErrUsage},
	}
	for _, step := range steps {
		result := Execute(store, step.line)
		if got := strings.Join(result.Lines, "\n"); got != step.want {
			t.Errorf("%q printed %q, want %q", step.line, got, step.want)
		}
		if !errors.Is(result.Err, step.err) {
			t.Errorf("%q error = %v, want %v", step.line, result.Err, step.err)
		}
	}
}

func TestExecuteEmptyLine(t *testing.T) {
	if result := Execute(NewMemoryStore(), "   "); len(result.Lines) != 0 || result.Err != nil {
		t.Errorf("empty line = %+v, want no output", result)
	}
}

func TestHelpListsEveryCommand(t *testing.T) {
	help := strings.Join(Execute(NewMemoryStore(), "help").Lines, "\n")
	for _, c := range Commands {
		if !strings.Contains(help, c.Usage) {
			t.Errorf("HELP does not show %q", c.Usage)
		}
	}
}
