package models

type JWT struct {
	JWT struct {
		SecretKey string `json:"secret_key"`
		ExpTime   string `json:"exp_time"`
	} `json:"jwt"`
}
