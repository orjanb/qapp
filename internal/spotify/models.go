package spotify

type Track struct {
	ID         string   `json:"id"`
	URI        string   `json:"uri"`
	Name       string   `json:"name"`
	Artists    []Artist `json:"artists"`
	Album      Album    `json:"album"`
	DurationMs int      `json:"duration_ms"`
}

type Artist struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Album struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Device struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}
