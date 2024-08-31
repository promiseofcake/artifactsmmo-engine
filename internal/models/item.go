package models

import "github.com/promiseofcake/artifactsmmo-go-client/client"

type BankItems []BankItem

type BankItem struct {
	Code     string `json:"code"`
	Quantity int    `json:"quantity"`
}

type Items []Item
type Item struct {
	client.ItemSchema
	Quantity int `json:"quantity"`
	RawCode  string
}
