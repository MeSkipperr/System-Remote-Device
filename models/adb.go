package models

type AdbConfigType struct {
	AdbPath           string            `json:"adbPath"`
	AdbPort           int               `json:"adbPort"`
	Package           map[string]string `json:"package"`
	VerificationSteps int               `json:"verificationSteps"`
	StatusMessage       map[string]string `json:"statusMessage"`
	AdbCommandTemplate map[string]string `json:"adbCommandTemplate"`
}

