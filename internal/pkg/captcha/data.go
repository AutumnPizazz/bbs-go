package captcha

type CaptchaData struct {
	Id          string `json:"id"`
	ImageBase64 string `json:"imageBase64"`
	ThumbBase64 string `json:"thumbBase64"`
	ThumbX      int    `json:"thumbX"`
	ThumbY      int    `json:"thumbY"`
	ThumbWidth  int    `json:"thumbWidth"`
	ThumbHeight int    `json:"thumbHeight"`
}
