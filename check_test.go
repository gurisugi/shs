package main

import (
	"testing"
)

func TestCheckChainedCommands(t *testing.T) {
	allowList := []string{"jq", "git log", "wc"}

	tests := []struct {
		name     string
		command  string
		wantCmd  string // 空なら許可
	}{
		{
			name:    "単一コマンド",
			command: "ls -la",
			wantCmd: "",
		},
		{
			name:    "jq | jq (全てallow list内)",
			command: "jq . file.json | jq .name",
			wantCmd: "",
		},
		{
			name:    "echo | jq (echoがallow list外)",
			command: "echo hello | jq .",
			wantCmd: "echo",
		},
		{
			name:    "git status && git diff (gitがallow list外)",
			command: "git status && git diff",
			wantCmd: "git",
		},
		{
			name:    "git log | wc (両方allow list内)",
			command: "git log --oneline | wc -l",
			wantCmd: "",
		},
		{
			name:    "jq . a.json; jq . b.json (セミコロン、全てallow list内)",
			command: "jq . a.json; jq . b.json",
			wantCmd: "",
		},
		{
			name:    "空コマンド",
			command: "",
			wantCmd: "",
		},
		{
			name:    "jqクエリ内にパイプあり（シングルクォート）",
			command: `gh api repos/o/r/pulls | jq '[.[] | select(.draft==false)]'`,
			wantCmd: "gh",
		},
		{
			name:    "jqクエリ内にパイプあり（allow list内のみ）",
			command: `jq '[.[] | select(.name)]' file.json | wc -l`,
			wantCmd: "",
		},
		{
			name:    "git log | jq (allow list内のみ)",
			command: `git log --format='%H' | jq -R .`,
			wantCmd: "",
		},
		{
			name:    "git push はallow list外",
			command: `git log --oneline | git push`,
			wantCmd: "git",
		},
		{
			name:    "コマンド置換内のパイプ（単一コマンドとして扱う）",
			command: `echo "$(echo foo | grep f)"`,
			wantCmd: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := checkChainedCommands(tt.command, allowList)
			if got != tt.wantCmd {
				t.Errorf("checkChainedCommands(%q) = %q, want %q", tt.command, got, tt.wantCmd)
			}
		})
	}
}
