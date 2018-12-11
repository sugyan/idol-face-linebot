package recognizer

// Label type
type Label struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Twitter     string `json:"twitter"`
}
