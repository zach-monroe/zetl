package models

type Quote struct {
	QuoteID int      `json:"quote_id"`
	UserID  int      `json:"user_id"`
	Quote   string   `json:"quote"`
	Author  string   `json:"author"`
	Book    string   `json:"book"`
	Tags    []string `json:"tags"`
}

type Quotes []Quote
