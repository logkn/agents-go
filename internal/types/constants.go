package types

type Role int

const (
	_         Role = iota
	User           // User role
	Assistant      // Assistant role
	System         // System role
	Tool           // Tool role
)
