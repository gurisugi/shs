package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"mvdan.cc/sh/v3/syntax"
)

type hookInput struct {
	ToolInput struct {
		Command string `json:"command"`
	} `json:"tool_input"`
}

type hookOutput struct {
	HookSpecificOutput struct {
		HookEventName            string `json:"hookEventName"`
		PermissionDecision       string `json:"permissionDecision"`
		PermissionDecisionReason string `json:"permissionDecisionReason"`
	} `json:"hookSpecificOutput"`
}

func runCheck() int {
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading stdin: %v\n", err)
		return 1
	}

	var input hookInput
	if err := json.Unmarshal(data, &input); err != nil {
		fmt.Fprintf(os.Stderr, "error parsing JSON: %v\n", err)
		return 1
	}

	command := input.ToolInput.Command
	if command == "" {
		return 0
	}

	cfg, err := loadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error loading config: %v\n", err)
		return 1
	}

	disallowed := checkChainedCommands(command, cfg.AllowList)
	if disallowed == "" {
		return 0
	}

	out := hookOutput{}
	out.HookSpecificOutput.HookEventName = "PreToolUse"
	out.HookSpecificOutput.PermissionDecision = "ask"
	out.HookSpecificOutput.PermissionDecisionReason = fmt.Sprintf(
		"Chained command contains non-allowlisted command: %s. Approve to proceed.", disallowed,
	)

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(out); err != nil {
		fmt.Fprintf(os.Stderr, "error encoding JSON: %v\n", err)
		return 1
	}
	return 0
}

// checkChainedCommands はコマンド文字列をシェルパーサーで解析し、
// チェーンされたコマンドに許可リスト外のものがあれば最初のものを返す。
// チェーンでない場合やすべて許可されている場合は空文字を返す。
func checkChainedCommands(command string, allowList []string) string {
	parser := syntax.NewParser()
	prog, err := parser.Parse(strings.NewReader(command), "")
	if err != nil {
		// パースエラーの場合は安全側に倒してチェックしない
		return ""
	}

	// 単一のシンプルコマンドならチェーン検出不要
	if !hasChain(prog) {
		return ""
	}

	// ASTを走査して全コマンド名を収集
	var disallowed string
	syntax.Walk(prog, func(node syntax.Node) bool {
		if disallowed != "" {
			return false
		}
		call, ok := node.(*syntax.CallExpr)
		if !ok || len(call.Args) == 0 {
			return true
		}

		cmdParts := collectCommandWords(call)
		if !isAllowed(cmdParts, allowList) {
			disallowed = cmdParts[0]
		}
		return true
	})

	return disallowed
}

// hasChain はプログラムにチェーン（パイプ、&&、||、;）が含まれるか判定する。
func hasChain(prog *syntax.File) bool {
	for _, stmt := range prog.Stmts {
		// 複数のstmtがある = ; で区切られている
		if len(prog.Stmts) > 1 {
			return true
		}
		// BinaryCmd (&&, ||) やパイプのチェック
		if hasChainInCmd(stmt.Cmd) {
			return true
		}
	}
	return false
}

func hasChainInCmd(cmd syntax.Command) bool {
	switch c := cmd.(type) {
	case *syntax.BinaryCmd:
		return true
	case *syntax.Subshell:
		for _, s := range c.Stmts {
			if len(c.Stmts) > 1 {
				return true
			}
			if hasChainInCmd(s.Cmd) {
				return true
			}
		}
	}
	return false
}

// collectCommandWords はCallExprから先頭のコマンド名（とサブコマンド）の文字列を返す。
func collectCommandWords(call *syntax.CallExpr) []string {
	var parts []string
	for _, word := range call.Args {
		s := wordToString(word)
		if s == "" || strings.HasPrefix(s, "-") {
			break
		}
		// パス区切りを含む場合はコマンド名として扱わない（リダイレクト先など）
		if strings.Contains(s, "/") && len(parts) > 0 {
			break
		}
		parts = append(parts, s)
	}
	return parts
}

// wordToString はWordノードからリテラル文字列を取得する。
// 変数展開やコマンド置換を含む場合は空文字を返す。
func wordToString(word *syntax.Word) string {
	var sb strings.Builder
	for _, part := range word.Parts {
		lit, ok := part.(*syntax.Lit)
		if !ok {
			return ""
		}
		sb.WriteString(lit.Value)
	}
	return sb.String()
}

// isAllowed はコマンドのワード列がallow listに含まれるか判定する。
// "git log" のように複数語のエントリにも対応。
func isAllowed(cmdParts []string, allowList []string) bool {
	if len(cmdParts) == 0 {
		return true
	}

	for _, allowed := range allowList {
		allowedParts := strings.Fields(allowed)
		if len(allowedParts) > len(cmdParts) {
			continue
		}
		match := true
		for i, p := range allowedParts {
			if cmdParts[i] != p {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}
