package main

import (
	"strings"

	"mvdan.cc/sh/v3/syntax"
)

// splitCommands はシェルコマンド文字列をパースし、
// チェーン（パイプ、&&、||、;）で分割された個々のコマンドを返す。
func splitCommands(command string) ([]string, error) {
	parser := syntax.NewParser()
	prog, err := parser.Parse(strings.NewReader(command), "")
	if err != nil {
		return nil, err
	}

	var commands []string
	for _, stmt := range prog.Stmts {
		collectFromCmd(stmt.Cmd, prog, &commands)
	}
	return commands, nil
}

func collectFromCmd(cmd syntax.Command, prog *syntax.File, out *[]string) {
	switch c := cmd.(type) {
	case *syntax.BinaryCmd:
		collectFromStmt(c.X, prog, out)
		collectFromStmt(c.Y, prog, out)
	default:
		if cmd != nil {
			*out = append(*out, printNode(cmd, prog))
		}
	}
}

func collectFromStmt(stmt *syntax.Stmt, prog *syntax.File, out *[]string) {
	if stmt == nil {
		return
	}
	collectFromCmd(stmt.Cmd, prog, out)
}

func printNode(node syntax.Node, prog *syntax.File) string {
	var sb strings.Builder
	syntax.NewPrinter(syntax.Minify(true)).Print(&sb, node)
	return strings.TrimSpace(sb.String())
}
