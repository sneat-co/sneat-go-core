package dbmodels

// IDTitle record
type IDTitle struct {
	ID    string `json:"id" firestore:"id"`
	Title string `json:"title" firestore:"title"`
}
