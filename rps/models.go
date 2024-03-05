package rps

import "time"

type Resume struct {
	FirstName        string        `json:"first_name"`
	MiddleName       string        `json:"middle_name"`
	LastName         string        `json:"last_name"`
	Summary          string        `json:"summary"`
	Pdf              string        `json:"pdf"`
	Location         Location      `json:"location"`
	Emails           []string      `json:"emails"`
	Profession       string        `json:"profession"`
	Positions        []Position    `json:"positions"`
	Educations       []Education   `json:"educations"`
	SocialUrls       []SocialUrl   `json:"social_urls"`
	PhoneNumbers     []PhoneNumber `json:"phone_numbers"`
	Languages        []string      `json:"languages"`
	DetectedLanguage string        `json:"detected_language"`
	Skills           []Skill       `json:"skills"`
	RawText          string        `json:"raw_text"`
}

type Position struct {
	Title           string     `json:"title"`
	TitleNormalized string     `json:"title_normalized"`
	Organization    string     `json:"organization"`
	StartDate       *time.Time `json:"start_date"`
	EndDate         *time.Time `json:"end_date"`
	Description     string     `json:"description"`
	Location        Location   `json:"location"`
	ManagementLevel string     `json:"management_level"`
}

type Education struct {
	Organization   string     `json:"organization"`
	Degree         string     `json:"degree"`
	StartDate      *time.Time `json:"start_date"`
	EndDate        *time.Time `json:"end_date"`
	Location       Location   `json:"location"`
	EducationLevel string     `json:"education_level"`
}

type SocialUrl struct {
	Source   string `json:"source"`
	Url      string `json:"url"`
	Username string `json:"username"`
}

type PhoneNumber struct {
	CountryCode    string `json:"country_code"`
	CountryName    string `json:"country_name"`
	NationalNumber string `json:"national_number"`
}

type Skill struct {
	Name      string `json:"name"`
	NumMonths int    `json:"num_months"`
}

type Location struct {
	Formatted   string `json:"formatted"`
	Street      string `json:"street"`
	City        string `json:"city"`
	State       string `json:"state"`
	Country     string `json:"country"`
	CountryCode string `json:"countryCode"`
}

type parseDocumentRequest struct {
	Base64Data string `json:"base64_data"`
}
