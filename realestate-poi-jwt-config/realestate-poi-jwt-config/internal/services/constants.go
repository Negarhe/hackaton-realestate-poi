package services

import "time"

const (
	GetPostUrl   = "https://api.divar.ir/v1/open-platform/finder/post/"
	AddWidgetUrl = "https://api.divar.ir/v2/open-platform/addons/post/"
)

type Payload struct {
	Widgets []Row `json:"widgets"`
}

type Widget struct {
	TitleRow       map[string]interface{} `json:"title_row,omitempty"`
	SubtitleRow    map[string]interface{} `json:"subtitle_row,omitempty"`
	DescriptionRow map[string]interface{} `json:"description_row,omitempty"`
}

type Row struct {
	Key  string
	Data map[string]interface{}
}

type propertyApiResponse struct {
	Data struct {
		Title     string  `json:"title,omitempty"`
		Latitude  float64 `json:"latitude,omitempty"`
		Longitude float64 `json:"longitude,omitempty"`
	} `json:"data"`
}

type userInfo struct {
	UserId string `json:"user_id"`
}

type propertyInfo struct {
	PostID    string  `json:"post_id"`
	Title     string  `json:"title"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}
type TokenInfo struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresIn    time.Time `json:"expires_in"`
}

type Transaction struct {
	IsBuyer        bool
	PropertyDetail *propertyInfo
	UserDetail     *userInfo
	TokenInfo      *TokenInfo
}
