package ui

import (
	"errors"
	"time"
)

type UserDataResponse struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	Username  string    `json:"username"`
	Avatar    struct {
		ID         string `json:"id"`
		Link       string `json:"link"`
		IsAnimated bool   `json:"is_animated"`
	} `json:"avatar"`
	AvatarDecoration struct {
		Asset     string `json:"asset"`
		SkuID     string `json:"sku_id"`
		ExpiresAt any    `json:"expires_at"`
	} `json:"avatar_decoration"`
	Badges      []string `json:"badges"`
	AccentColor int      `json:"accent_color"`
	GlobalName  string   `json:"global_name"`
	Banner      struct {
		ID         any    `json:"id"`
		Link       any    `json:"link"`
		IsAnimated bool   `json:"is_animated"`
		Color      string `json:"color"`
	} `json:"banner"`
	Raw struct {
		ID                   string `json:"id"`
		Username             string `json:"username"`
		Avatar               string `json:"avatar"`
		Discriminator        string `json:"discriminator"`
		PublicFlags          int    `json:"public_flags"`
		Flags                int    `json:"flags"`
		Banner               any    `json:"banner"`
		AccentColor          int    `json:"accent_color"`
		GlobalName           string `json:"global_name"`
		AvatarDecorationData struct {
			Asset     string `json:"asset"`
			SkuID     string `json:"sku_id"`
			ExpiresAt any    `json:"expires_at"`
		} `json:"avatar_decoration_data"`
		BannerColor string `json:"banner_color"`
		Clan        any    `json:"clan"`
	} `json:"raw"`
}

func GetUserData(userID string) (*UserDataResponse, error) {
	return nil, errors.New("service down")
	//var baseURL = "https://discordlookup.mesalytic.moe/v1/user/"
	//var url = baseURL + userID
	//
	////get request, load json into UserDataResponse
	//var userData UserDataResponse
	//resp, err := http.Get(url)
	//if resp == nil || resp.StatusCode != 200 {
	//	return nil, err
	//}
	//if resp.Body == nil {
	//	return nil, err
	//}
	//defer func(Body io.ReadCloser) {
	//	err := Body.Close()
	//	if err != nil {
	//		return
	//	}
	//}(resp.Body)
	//if err != nil {
	//	return nil, err
	//}
	//
	////parse json
	//err = json.NewDecoder(resp.Body).Decode(&userData)
	//if err != nil {
	//	return nil, err
	//}
	//return &userData, nil
}
