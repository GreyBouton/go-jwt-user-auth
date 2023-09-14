package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

/*
Retrieves variable from .env file by name
*/
func GetEnvVar(variable string) (string, error) {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
		return "", err
	}

	res := os.Getenv(variable)
	return res, nil
}
