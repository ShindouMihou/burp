package console

import (
	"fmt"
	"github.com/ttacon/chalk"
)

func Clear() {
	fmt.Println("\033[H\033[2J")
}

var Highlight = chalk.Black.NewStyle().WithForeground(chalk.ResetColor).WithBackground(chalk.Green)
