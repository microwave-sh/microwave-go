package management

type SearchSort struct {
	Field     string `json:"field"`
	Direction string `json:"direction,omitempty"`
}

type SearchRequest struct {
	Query  string                    `json:"query,omitempty"`
	Filter map[string]map[string]any `json:"filter,omitempty"`
	Sort   []SearchSort              `json:"sort,omitempty"`
	Cursor string                    `json:"cursor,omitempty"`
	Offset *int                      `json:"offset,omitempty"`
	Limit  *int                      `json:"limit,omitempty"`
	Fields []string                  `json:"fields,omitempty"`
}

type SearchResponse[T any] struct {
	Data       []T    `json:"data"`
	NextCursor string `json:"next_cursor,omitempty"`
	HasMore    bool   `json:"has_more"`
	Limit      int    `json:"limit"`
}
