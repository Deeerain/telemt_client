package telemtclient

type TelemtError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type TelemtResponse[T any] struct {
	Ok    bool         `json:"ok"`
	Data  *T           `json:"data"`
	Error *TelemtError `json:"error"`
}

type UserInfo struct {
	Username string    `json:"username"`
	Secret   string    `json:"secret"`
	Links    UserLinks `json:"links"`
}

type CreateUserRequest struct {
	Username string  `json:"username"`
	Secret   *string `json:"secret"`
}

type UserLinks struct {
	Tls []string `json:"tls"`
}

type CreateUserResponse struct {
	User   UserInfo
	Secret string `json:"secret"`
}
