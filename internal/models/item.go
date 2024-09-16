package models

import "github.com/promiseofcake/artifactsmmo-go-client/client"

type SimpleItems []SimpleItem

type SimpleItem struct {
	Code     string `json:"code"`
	Quantity int    `json:"quantity"`
}

type Items []*Item
type Item struct {
	client.ItemSchema
	Quantity       int `json:"quantity"`
	CraftMaterials []*CraftResource
	Skill          string
}

type CraftResource struct {
	RequiredCode    string
	CostPerResource int
	Available       int
}

type Order struct {
	Item        SimpleItem `json:"item"`
	Concurrency int        `json:"concurrency"`
}
