package model

import "fmt"

type ListArgument []string

func (i *ListArgument) String() string {
	return fmt.Sprintf("%v", *i)
}

func (i *ListArgument) Set(value string) error {
	*i = append(*i, value)
	return nil
}
