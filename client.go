package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
)

type Client struct {
	host string
	port int
}

func CreateClient(host string, port int) *Client {
	return &Client{
		host: host,
		port: port,
	}
}

func (c *Client) doRequest(method, endpoint string, body any) ([]byte, error) {
	url := fmt.Sprintf("http://%s:%v/v1/%s", c.host, c.port, endpoint)

	log.Println(url)

	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("Marshalling request body failed: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("Creating request failed: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Performing request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Reading response body failed: %w", err)
	}

	return respBody, nil
}

func (s *Client) CreateUser(req *CreateUserRequest) (*CreateUserResponse, error) {
	respBody, err := s.doRequest("POST", "users", req)
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	var resp TelemtResponse[CreateUserResponse]

	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("marshal response: %w", err)
	}

	if !resp.Ok {
		return nil, fmt.Errorf("Error for create user: %s", resp.Error.Message)
	}

	return resp.Data, nil
}

func (s *Client) IsHasUser(username string) (bool, error) {
	respBody, err := s.doRequest("GET", "users/"+username, nil)
	if err != nil {
		return false, fmt.Errorf("Faild to create user: %s", err)
	}

	var resp struct {
		Ok bool `json:"ok"`
	}

	if !resp.Ok {
		return false, fmt.Errorf("API returned unsuccessful response: %s", string(respBody))
	}

	return true, nil
}

func (s *Client) getByUsername(username string) (*UserInfo, error) {
	respBody, err := s.doRequest("GET", fmt.Sprintf("users/%s", username), nil)
	if err != nil {
		return nil, fmt.Errorf("Failed to get users: %w", err)
	}

	var response TelemtResponse[UserInfo]

	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("Failed to unmarshal response: %w", err)
	}

	if !response.Ok {
		switch response.Error.Code {
		case "not_found":
			return nil, ErrUserNotFound
		}
	}

	return response.Data, nil
}

func (s *Client) GetOrCreate(username string) (*UserInfo, error) {
	user, err := s.getByUsername(username)
	if err != nil {
		switch {
		case errors.Is(err, ErrUserNotFound):
			var resp *CreateUserResponse
			resp, err = s.CreateUser(&CreateUserRequest{
				Username: username,
			})
			if err != nil {
				return nil, err
			}

			user = &resp.User
		default:
			return nil, err
		}
	}

	return user, nil
}

func (s *Client) GetConfig(username string) (*string, error) {
	user, err := s.GetOrCreate(username)
	if err != nil {
		return nil, err
	}

	if len := len(user.Links.Tls); len <= 0 {
		return nil, fmt.Errorf("Config not found; coount = %v", len)
	}

	return &user.Links.Tls[0], nil
}
