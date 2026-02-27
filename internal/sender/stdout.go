package sender

import (
	"context"
	"fmt"
)

type StdoutSender struct{}

func (s *StdoutSender) Send(_ context.Context, text string) error {
	fmt.Println("================================")
	fmt.Println(text)
	fmt.Println("================================")
	return nil
}
