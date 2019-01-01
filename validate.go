package go_tree

import (
	"errors"
	"github.com/urfave/cli"
)

func Validate(c *cli.Context) error {
	err := validateLevel(c.Int("L"))
	if err != nil {
		return err
	}
	return nil
}

func validateLevel(level int) error {
	if level <= 0 {
		return errors.New("invalid level, must be greater than 0")
	}
	return nil
}
