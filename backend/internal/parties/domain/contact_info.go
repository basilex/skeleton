package domain

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

var emailRegex = regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,}$`)
var phoneRegex = regexp.MustCompile(`^\+?[1-9]\d{6,14}$`)

type Address struct {
	Street     string `json:"street"`
	City       string `json:"city"`
	Region     string `json:"region"`
	PostalCode string `json:"postal_code"`
	Country    string `json:"country"`
}

type ContactInfo struct {
	Email       string            `json:"email"`
	Phone       string            `json:"phone"`
	Address     Address           `json:"address"`
	Website     string            `json:"website"`
	SocialMedia map[string]string `json:"social_media"`
}

func NewContactInfo(email, phone string, address Address) (ContactInfo, error) {
	ci := ContactInfo{
		Email:       strings.ToLower(strings.TrimSpace(email)),
		Phone:       strings.TrimSpace(phone),
		Address:     address,
		SocialMedia: make(map[string]string),
	}

	if err := ci.Validate(); err != nil {
		return ContactInfo{}, err
	}

	return ci, nil
}

func (ci ContactInfo) Validate() error {
	if ci.Email != "" && !emailRegex.MatchString(ci.Email) {
		return fmt.Errorf("invalid email format")
	}
	if ci.Phone != "" && !phoneRegex.MatchString(strings.ReplaceAll(ci.Phone, " ", "")) {
		return fmt.Errorf("invalid phone format")
	}
	return nil
}

func (ci ContactInfo) ToJSON() (json.RawMessage, error) {
	data, err := json.Marshal(ci)
	if err != nil {
		return nil, fmt.Errorf("marshal contact info: %w", err)
	}
	return json.RawMessage(data), nil
}

func ContactInfoFromJSON(data json.RawMessage) (ContactInfo, error) {
	var ci ContactInfo
	if err := json.Unmarshal(data, &ci); err != nil {
		return ContactInfo{}, fmt.Errorf("unmarshal contact info: %w", err)
	}
	return ci, nil
}
