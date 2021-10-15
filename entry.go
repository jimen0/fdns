package fdns

//go:generate easyjson -all entry.go

// entry represents each element of the dataset.
//
//easyjson:json
type entry struct {
	Timestamp string `json:"-"`
	Name      string `json:"name"`
	Type      string `json:"type"`
	Value     string `json:"value"`
}
