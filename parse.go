package main

import (
	"strings"

	"mvdan.cc/sh/v3/syntax"
)

// splitCommands はシェルコマンド文字列をパースし、
// チェーン（パイプ、&&、||、;）やコマンド置換内のコマンドも含めて
// フラットに個々のコマンドを返す。
func splitCommands(command string) ([]string, error) {
	_, commands, err := parseCommands(command)
	return commands, err
}

// commandNames はシェルコマンド文字列をパースし、
// 各コマンドの名前部分（サブコマンド含む、引数除く）を返す。
func commandNames(command string) ([]string, error) {
	calls, _, err := parseCommands(command)
	if err != nil {
		return nil, err
	}
	var names []string
	for _, call := range calls {
		names = append(names, extractName(call))
	}
	return names, err
}

func parseCommands(command string) ([]*syntax.CallExpr, []string, error) {
	parser := syntax.NewParser()
	prog, err := parser.Parse(strings.NewReader(command), "")
	if err != nil {
		return nil, nil, err
	}

	var calls []*syntax.CallExpr
	var commands []string
	for _, stmt := range prog.Stmts {
		collectFromStmt(stmt, prog, &calls, &commands)
	}
	return calls, commands, nil
}

func collectFromStmt(stmt *syntax.Stmt, prog *syntax.File, calls *[]*syntax.CallExpr, out *[]string) {
	if stmt == nil {
		return
	}
	collectFromCmd(stmt.Cmd, prog, calls, out)
}

func collectFromCmd(cmd syntax.Command, prog *syntax.File, calls *[]*syntax.CallExpr, out *[]string) {
	if cmd == nil {
		return
	}
	switch c := cmd.(type) {
	case *syntax.BinaryCmd:
		collectFromStmt(c.X, prog, calls, out)
		collectFromStmt(c.Y, prog, calls, out)
	case *syntax.Subshell:
		for _, stmt := range c.Stmts {
			collectFromStmt(stmt, prog, calls, out)
		}
	case *syntax.CallExpr:
		*calls = append(*calls, c)
		*out = append(*out, printRedacted(cmd, prog))
		collectSubstitutions(cmd, prog, calls, out)
	default:
		*calls = append(*calls, nil)
		*out = append(*out, printRedacted(cmd, prog))
		collectSubstitutions(cmd, prog, calls, out)
	}
}

// collectSubstitutions はノード内のコマンド置換 $() を探索し、
// 中のコマンドを収集する。
func collectSubstitutions(node syntax.Node, prog *syntax.File, calls *[]*syntax.CallExpr, out *[]string) {
	syntax.Walk(node, func(n syntax.Node) bool {
		cs, ok := n.(*syntax.CmdSubst)
		if !ok {
			return true
		}
		for _, stmt := range cs.Stmts {
			collectFromStmt(stmt, prog, calls, out)
		}
		return false
	})
}

// サブコマンドを持つことが既知のコマンド
var subcommandCommands = map[string]bool{
	"git":     true,
	"docker":  true,
	"kubectl": true,
	"go":      true,
	"npm":     true,
	"yarn":    true,
	"cargo":   true,
	"make":    true,
	"brew":    true,
	"asdf":    true,
	"gh":      true,
	"aws":     true,
	"gcloud":  true,
	"az":      true,
	"heroku":  true,
	"flyctl":  true,
	"terraform": true,
	"systemctl": true,
	"journalctl": true,
	"ip":      true,
}

// extractName はCallExprからコマンド名を抽出する。
// 既知のサブコマンドを持つコマンド（git, docker等）は2語目まで含める。
// それ以外は先頭1語のみ。
func extractName(call *syntax.CallExpr) string {
	if call == nil || len(call.Args) == 0 {
		return ""
	}

	first := wordToLiteral(call.Args[0])
	if first == "" {
		return ""
	}

	if subcommandCommands[first] && len(call.Args) > 1 {
		second := wordToLiteral(call.Args[1])
		if second != "" && !strings.HasPrefix(second, "-") {
			return first + " " + second
		}
	}

	return first
}

// wordToLiteral はWordノードからリテラル文字列を取得する。
// 変数展開やコマンド置換を含む場合は空文字を返す。
func wordToLiteral(word *syntax.Word) string {
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

// printRedacted はノードを出力する際、コマンド置換の中身を$(...)に置換する。
func printRedacted(node syntax.Node, prog *syntax.File) string {
	var sb strings.Builder
	printer := syntax.NewPrinter(syntax.Minify(true))

	// コマンド置換のStmtsを一時的に空にしてprint
	var saved []savedSubst
	syntax.Walk(node, func(n syntax.Node) bool {
		cs, ok := n.(*syntax.CmdSubst)
		if !ok {
			return true
		}
		saved = append(saved, savedSubst{cs: cs, stmts: cs.Stmts})
		cs.Stmts = nil
		return true
	})

	printer.Print(&sb, node)

	// Stmtsを復元
	for _, s := range saved {
		s.cs.Stmts = s.stmts
	}

	return strings.TrimSpace(sb.String())
}

type savedSubst struct {
	cs    *syntax.CmdSubst
	stmts []*syntax.Stmt
}
