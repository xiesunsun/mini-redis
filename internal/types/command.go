package types

// Command 表示一条解析后的客户端命令。
type Command struct {
	Name string   // 命令名称，如 "SET"、"GET"
	Args []string // 命令参数，如 ["name", "Alice"]
}
