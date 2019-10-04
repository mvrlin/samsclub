package samsclub

// Cards represents cards from response.
type Cards []struct {
	ExpMonth string `json:"expMonth"`
	ExpYear  string `json:"expYear"`
	Expired  bool   `json:"expired"`
	Number   string `json:"lastFourDigits"`
	Type     string `json:"cardTypeName"`
}

// Member represents a member from response.
type Member struct {
	Address string `json:"addressLineOne"`
	City    string `json:"city"`
	State   string `json:"stateOrProvinceCode"`
	Zip     string `json:"postalCode"`

	FirstName      string `json:"firstName"`
	LastName       string `json:"lastName"`
	MembershipType string `json:"membershipType"`
}

// ResponseAuthenticate represents a response from Authenticate.
type ResponseAuthenticate struct {
	Member Member `json:"member"`

	Message string `json:"message"`
	Status  string `json:"status"`
}

// ResponseListCards respresents a response from ListCards.
type ResponseListCards struct {
	Cards Cards `json:"cards"`

	Message string `json:"message"`
	Status  string `json:"status"`
}
