package types

// Role represents the role of a message participant in the conversation.
type Role int

// Enumeration of available message roles.
const (
	_         Role = iota
	User           // User role
	Assistant      // Assistant role
	System         // System role
	Tool           // Tool role
)
