package app

import "regexp"

type Message struct {
	Email          string
	Password       string
	PasswordVerify string
	Username       string
	//Content      string
	Errors      map[string]interface{}
	Firstname   string
	Lastname    string
	ID          string
	RedirectURL string
}

func (msg *Message) Validate() bool {

	re := regexp.MustCompile(".+@.+\\..+")
	matched := re.Match([]byte(msg.Email))
	if matched == false {
		msg.Errors["Email"] = "Please enter a valid email address"
	}

	if msg.PasswordVerify != msg.Password {
		msg.Errors["Password"] = "Password not the same with verify Password"
	}

	// if strings.TrimSpace(msg.Content) == "" {
	// 	msg.Errors["Content"] = "Please write a message"
	// }

	return len(msg.Errors) == 0
}
