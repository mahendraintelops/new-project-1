package models

type Account struct {
	ID string `json:"id,omitempty" bson:"_id,omitempty"`

	AccountNumber string `json:"accountNumber,omitempty" bson:"accountNumber,omitempty"`

	City string `json:"city,omitempty" bson:"city,omitempty"`

	Name string `json:"name,omitempty" bson:"name,omitempty"`
}
