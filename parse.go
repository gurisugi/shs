package main

import (
	"strings"

	"mvdan.cc/sh/v3/syntax"
)

// splitCommands はシェルコマンド文字列をパースし、
// チェーン（パイプ、&&、||、;）やコマンド置換内のコマンドも含めて
// フラットに個々のコマンドを返す。
func splitCommands(command string) ([]string, error) {
	parser := syntax.NewParser()
	prog, err := parser.Parse(strings.NewReader(command), "")
	if err != nil {
		return nil, err
	}

	var commands []string
	for _, stmt := range prog.Stmts {
		collectFromStmt(stmt, prog, &commands)
	}
	return commands, nil
}

func collectFromStmt(stmt *syntax.Stmt, prog *syntax.File, out *[]string) {
	if stmt == nil {
		return
	}
	collectFromCmd(stmt.Cmd, prog, out)
}

func collectFromCmd(cmd syntax.Command, prog *syntax.File, out *[]string) {
	if cmd == nil {
		return
	}
	switch c := cmd.(type) {
	case *syntax.BinaryCmd:
		collectFromStmt(c.X, prog, out)
		collectFromStmt(c.Y, prog, out)
	case *syntax.Subshell:
		for _, stmt := range c.Stmts {
			collectFromStmt(stmt, prog, out)
		}
	case *syntax.CallExpr:
		*out = append(*out, printRedacted(cmd, prog))
		collectSubstitutions(cmd, prog, out)
	default:
		*out = append(*out, printRedacted(cmd, prog))
		collectSubstitutions(cmd, prog, out)
	}
}

// collectSubstitutions はノード内のコマンド置換 $() を探索し、
// 中のコマンドを収集する。
func collectSubstitutions(node syntax.Node, prog *syntax.File, out *[]string) {
	syntax.Walk(node, func(n syntax.Node) bool {
		cs, ok := n.(*syntax.CmdSubst)
		if !ok {
			return true
		}
		for _, stmt := range cs.Stmts {
			collectFromStmt(stmt, prog, out)
		}
		return false
	})
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
