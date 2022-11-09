package metrics

type PageView struct {
	Domain       string `json:"domain"`
	IPAddress    string `json:"ipAddress"`
	URL          string `json:"url"`
	UserAgent    string `json:"userAgent"`
	ScreenHeight int    `json:"screenHeight"`
	ScreenWidth  int    `json:"screenWidth"`
}
