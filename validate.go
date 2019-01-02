package go_tree

import (
	"errors"
	"github.com/urfave/cli"
	"strconv"
)

func ValidateFlag(c *cli.Context) error {
	var err error
	err = validateLevel(c.String("L"))
	if err != nil {
		return err
	}
	return nil
}

func validateLevel(levelStr string) error {
	if levelStr == "" {
		return nil
	}

	level, err := strconv.Atoi(levelStr)
	if err != nil || level <= 0 {
		return errors.New("invalid level, must be greater than 0")
	}
	return nil
}
