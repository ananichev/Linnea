package config

type Config struct {
	Twitch struct {
		ClientID     string `json:"client_id"`
		ClientSecret string `json:"client_secret"`
	} `json:"twitch"`

	S3 struct {
		AccessToken string `json:"access_token"`
		SecretKey   string `json:"secret_key"`
		Region      string `json:"region"`
		Bucket      string `json:"bucket"`
		Endpoint    string `json:"endpoint"`
		Namespace   string `json:"namespace"`
	} `json:"s3"`

	DocDB struct {
		Endpoint    string `json:"endpoint"`
		Region      string `json:"region"`
		AccessToken string `json:"access_token"`
		SecretKey   string `json:"secret_key"`
	} `json:"doc_db"`

	Redis struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Database int    `json:"database"`
		Address  string `json:"address"`
	} `json:"redis"`

	Http struct {
		PublicAddr string `json:"public_addr"`
		Port       int    `json:"port"`
		Cookie     struct {
			Domain string `json:"domain"`
			Secure bool   `json:"secure"`
		} `json:"cookie"`
		Jwt struct {
			Secret string `json:"secret"`
		} `json:"jwt"`
	} `json:"http"`
}
