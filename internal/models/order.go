package models

type Order struct {
	Item        SimpleItem `json:"item"`
	Concurrency int        `json:"concurrency"`
	Action      string     `json:"action"`
}
