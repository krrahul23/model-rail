package constant

// Role identifies the author of a message in a conversation.
type Role string

// Role constants
const (
	// RoleSystem carries instructions that apply to the whole conversation.
	RoleSystem Role = "system"
	// RoleUser carries input produced by the end user.
	RoleUser Role = "user"
	// RoleAssistant carries model output, including earlier turns.
	RoleAssistant Role = "assistant"
	// RoleTool carries the result of a tool invocation.
	RoleTool Role = "tool"
)

// ContentType identifies the kind of payload carried by a ContentPart.
type ContentType string

// ContentType constants
const (
	// ContentText is plain text content. Additional content types such as
	// images or tool calls are added when the corresponding capability ships.
	ContentText ContentType = "text"
)
