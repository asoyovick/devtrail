package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"devtrail/cli/internal/config"
)

func Login(args []string) error {
	var email, password string
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--email":
			if i+1 < len(args) { email = args[i+1]; i++ }
		case "--password":
			if i+1 < len(args) { password = args[i+1]; i++ }
		case "--api":
			if i+1 < len(args) { i++ } // handled below
		}
	}
	if email == "" || password == "" { return fmt.Errorf("usage: devtrail login --email <email> --password <password> [--api <url>]") }
	cfg, err := config.Load()
	if err != nil { return err }
	for i := 0; i < len(args); i++ { if args[i] == "--api" && i+1 < len(args) { cfg.APIBaseURL = args[i+1] } }
	body, _ := json.Marshal(map[string]string{"email": email, "password": password})
	res, err := http.Post(cfg.APIBaseURL+"/api/v1/auth/login", "application/json", bytes.NewReader(body))
	if err != nil { return err }
	defer res.Body.Close()
	respBody, _ := io.ReadAll(res.Body)
	if res.StatusCode >= 300 { return fmt.Errorf("login failed: %s", string(respBody)) }
	var out struct { Token string `json:"token"`; User struct{ Email string `json:"email"` } `json:"user"` }
	if err := json.Unmarshal(respBody, &out); err != nil { return err }
	cfg.AuthToken = out.Token
	cfg.UserEmail = out.User.Email
	if err := config.Save(cfg); err != nil { return err }
	fmt.Println("Logged in as", out.User.Email)
	return nil
}
