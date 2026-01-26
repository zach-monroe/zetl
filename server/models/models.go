package models

type Quote struct {
	QuoteID int      `json:"quote_id"`
	UserID  int      `json:"user_id"`
	Quote   string   `json:"quote"`
	Author  string   `json:"author"`
	Book    string   `json:"book"`
	Tags    []string `json:"tags"`
	Notes   string   `json:"notes"`
}

// ToMap converts a Quote to a map for JSON serialization
func (q Quote) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"quote_id": q.QuoteID,
		"user_id":  q.UserID,
		"quote":    q.Quote,
		"author":   q.Author,
		"book":     q.Book,
		"tags":     q.Tags,
		"notes":    q.Notes,
	}
}

type Quotes []Quote
