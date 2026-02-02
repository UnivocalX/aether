package actions

import (
	"context"
	"fmt"
	"os"
	"strings"
)

func AskForApproval(ctx context.Context, prompt string, assumeYes bool) (bool, error) {
	if assumeYes {
		return true, nil
	}

	stat, err := os.Stdin.Stat()
	if err != nil {
		return false, err
	}
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		return false, fmt.Errorf("approval required but stdin is not interactive")
	}

	fmt.Printf("%s [y/N]: ", prompt)

	inputCh := make(chan string, 1)
	go func() {
		var input string
		fmt.Scanln(&input)
		inputCh <- strings.TrimSpace(input)
	}()

	select {
	case <-ctx.Done():
		return false, ctx.Err()
	case input := <-inputCh:
		switch strings.ToLower(input) {
		case "y", "yes":
			return true, nil
		default:
			return false, nil
		}
	}
}
