package models

type Model interface {
	ToString() (string, error)
}
