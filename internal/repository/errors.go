package repository

import "errors"

var ErrNotFound = errors.New("not found")

func IsBadRequest(err error) bool {
	return err != nil && (errors.Is(err, ErrNotFound))
}
