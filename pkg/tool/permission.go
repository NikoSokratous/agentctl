package tool

// Permission defines a scoped permission for a tool.
type Permission struct {
	Scope    string // e.g., "fs:read", "net:external", "email:send"
	Required bool
}
