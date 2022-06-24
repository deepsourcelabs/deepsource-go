package types

type Issue struct {
	IssueCode   string `toml:"issue_code"`
	Category    string `toml:"category"`
	Title       string `toml:"title"`
	Description string `toml:"description"`
}
