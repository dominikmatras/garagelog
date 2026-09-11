package vehicle

type Vehicle struct {
	ID      int64  `json:"id"`
	Brand   string `json:"brand"`
	Model   string `json:"model"`
	Year    int    `json:"year"`
	Mileage int    `json:"mileage"`
}