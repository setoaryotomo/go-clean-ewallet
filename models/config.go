package models

type EnvironmentConfig struct {
	AppEnv		string	`json:"app_env"`
	AppName		string	`json:"app_name"`
	AppVersion	string	`json:"app_version"`
}

type Config struct {
	DriverName	string	 `json:"driver"`
	Database	Database `json:"database"`
}

type Database struct {
	DBName		string	`json:"db_name"`
	DBPort		string	`json:"db_port"`
	DBUser		string	`json:"db_user"`
	DBPass		string	`json:"db_pass"`
	SSLMode		string	`json:"ssl_mode"`
}