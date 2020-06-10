package menu

// Menu general menu struct
type Menu struct {
	Title      string `json:"title"`
	Url        string `json:"url"`
	Icon       string `json:"icon"`
	Attributes map[string]string
	Permission string `json:"permission"`
	Children   []Menu `json:"children"`
}
