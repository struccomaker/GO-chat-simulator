package models

type MessageType int

const (
	ChatMessage MessageType = iota
	SystemMessage
)

type Message struct {
	Username string
	Content  string
	Type     MessageType
}
