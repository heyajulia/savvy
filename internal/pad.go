package internal

import "fmt"

func Pad(i int) string {
	return fmt.Sprintf("%02d", i)
}
