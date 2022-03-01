package v1alpha1

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// FeishuNotifier defines the spec for integrating with Slack
// +k8s:openapi-gen=true
type FeishuNotifier struct {
	WebhookUrl string `json:"webhook_url"`
	Channel    string `json:"channel,omitempty"`
	Username   string `json:"username,omitempty"`
	IconEmoji  string `json:"icon_emoji,omitempty"`
}

func (n *FeishuNotifier) Send(message string) error {
	url := n.WebhookUrl
	channel := ``
	if n.Channel != `` {
		channel = fmt.Sprintf(`, "channel":"%s"`, n.Channel)
	}
	username := ``
	if n.Username != `` {
		username = fmt.Sprintf(`, "username":"%s"`, n.Username)
	}
	icon_emoji := ``
	if n.IconEmoji != `` {
		icon_emoji = fmt.Sprintf(`, "icon_emoji":"%s"`, n.IconEmoji)
	}

	var jsonStr = []byte(fmt.Sprintf(`{"text":"%s"%s%s%s}`, escapeString(message), channel, username, icon_emoji))
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(json.RawMessage(jsonStr)))
	req.Header.Set("Content-Type", "application/json")

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}
