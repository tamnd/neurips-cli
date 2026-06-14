package neurips

// Paper is a NeurIPS conference paper.
type Paper struct {
	Rank  int    `json:"rank"`
	Year  int    `json:"year"`
	Title string `json:"title"`
	URL   string `json:"url"`
}

// Year is an available NeurIPS proceedings year.
type Year struct {
	Year  int `json:"year"`
	Count int `json:"count"`
}
