package main

import (
	"strings"
	"testing"
)

func TestSplitCommands(t *testing.T) {
	tests := []struct {
		name    string
		command string
		want    []string
	}{
		{
			name:    "単一コマンド",
			command: "ls -la",
			want:    []string{"ls -la"},
		},
		{
			name:    "パイプ",
			command: "git log --oneline | wc -l",
			want:    []string{"git log --oneline", "wc -l"},
		},
		{
			name:    "AND",
			command: "make build && make test",
			want:    []string{"make build", "make test"},
		},
		{
			name:    "OR",
			command: "test -f file || echo missing",
			want:    []string{"test -f file", "echo missing"},
		},
		{
			name:    "セミコロン",
			command: "echo hello; echo world",
			want:    []string{"echo hello", "echo world"},
		},
		{
			name:    "複合チェーン",
			command: "git log --oneline | wc -l && echo done",
			want:    []string{"git log --oneline", "wc -l", "echo done"},
		},
		{
			name:    "3段パイプ",
			command: "cat file | grep foo | wc -l",
			want:    []string{"cat file", "grep foo", "wc -l"},
		},
		{
			name:    "クォート内のパイプは分割しない",
			command: `jq '[.[] | select(.name)]' file.json`,
			want:    []string{`jq '[.[] | select(.name)]' file.json`},
		},
		{
			name:    "パイプ+クォート内パイプ",
			command: `echo hello | jq '[.[] | select(.x)]'`,
			want:    []string{"echo hello", `jq '[.[] | select(.x)]'`},
		},
		{
			name:    "コマンド置換内のコマンドも展開",
			command: `echo "$(cat file)"`,
			want:    []string{`echo "$()"`, "cat file"},
		},
		{
			name:    "コマンド置換内のパイプも展開",
			command: `echo "$(echo foo | grep f)"`,
			want:    []string{`echo "$()"`, "echo foo", "grep f"},
		},
		{
			name:    "コマンド置換+チェーン",
			command: `echo "$(cat file)" && ls`,
			want:    []string{`echo "$()"`, "cat file", "ls"},
		},
		{
			name:    "空文字列",
			command: "",
			want:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := splitCommands(tt.command)
			if err != nil {
				t.Fatalf("splitCommands(%q) error: %v", tt.command, err)
			}
			if len(got) != len(tt.want) {
				t.Fatalf("splitCommands(%q) = %v (len=%d), want %v (len=%d)",
					tt.command, got, len(got), tt.want, len(tt.want))
			}
			for i := range got {
				if strings.TrimSpace(got[i]) != strings.TrimSpace(tt.want[i]) {
					t.Errorf("splitCommands(%q)[%d] = %q, want %q",
						tt.command, i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestCommandNames(t *testing.T) {
	tests := []struct {
		name    string
		command string
		want    []string
	}{
		{
			name:    "単一コマンド",
			command: "ls -la",
			want:    []string{"ls"},
		},
		{
			name:    "パイプ",
			command: "git log --oneline | wc -l",
			want:    []string{"git", "wc"},
		},
		{
			name:    "コマンド置換内も展開",
			command: `echo "$(cat file)" && ls`,
			want:    []string{"echo", "cat", "ls"},
		},
		{
			name:    "AND",
			command: "make build && make test",
			want:    []string{"make", "make"},
		},
		{
			name:    "空文字列",
			command: "",
			want:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := commandNames(tt.command)
			if err != nil {
				t.Fatalf("commandNames(%q) error: %v", tt.command, err)
			}
			if len(got) != len(tt.want) {
				t.Fatalf("commandNames(%q) = %v (len=%d), want %v (len=%d)",
					tt.command, got, len(got), tt.want, len(tt.want))
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("commandNames(%q)[%d] = %q, want %q",
						tt.command, i, got[i], tt.want[i])
				}
			}
		})
	}
}
