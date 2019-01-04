package fdns

// entry represents each element of the dataset.
type entry struct {
	Timestamp string `json:"-"`
	Name      string `json:"name"`
	Type      string `json:"type"`
	Value     string `json:"value"`
}
