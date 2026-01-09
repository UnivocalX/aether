package pipelines

import (
	"context"
	"fmt"
	"log"
)

func main() {
	ctx := context.Background()

	squareStage :=
		NewStage("square", square).
			WithInput(1, 2, 3, 4, 5)

	toStringStage :=
		NewStage("to-string", toString).
			WithStream(squareStage.Output)

	printStage :=
		NewStage("print", printValue).
			WithStream(toStringStage.Output)

	go func() {
		if err := squareStage.Run(ctx); err != nil {
			log.Fatal(err)
		}
	}()

	go func() {
		if err := toStringStage.Run(ctx); err != nil {
			log.Fatal(err)
		}
	}()

	if err := printStage.Run(ctx); err != nil {
		log.Fatal(err)
	}
}

func square(n int) (int, error) {
	return n * n, nil
}

func toString(n int) (string, error) {
	return fmt.Sprintf("value=%d", n), nil
}

func printValue(s string) (struct{}, error) {
	fmt.Println(s)
	return struct{}{}, nil
}
