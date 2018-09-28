package obsws

import (
	"fmt"
	"strings"
)

type response interface {
	error() error
	messageID() string
}

type ResponseBase struct {
	MessageID string `json:"message-id"`
	Status    string `json:"status"`
	Error     string `json:"error"`
}

func (r ResponseBase) error() error {
	if strings.ToLower(r.Status) == "ok" {
		return nil
	}
	return fmt.Errorf("obsws: status:%s error:%s", r.Status, r.Error)
}

func (r ResponseBase) messageID() string {
	return r.MessageID
}

type GetAuthRequiredResponse struct {
	AuthRequired bool   `json:"authRequired"`
	Challenge    string `json:"challenge"`
	Salt         string `json:"salt"`
	*ResponseBase
}

type Source struct {
	Type   string  `json:"type"`
	Volume float64 `json:"volume"`
}

type Scene struct {
	Name    string   `json:"name"`
	Sources []Source `json:"sources"`
}

type GetSceneListResponse struct {
	CurrentScene string  `json:"current-scene"`
	Scenes       []Scene `json:"scenes"`
	*ResponseBase
}

type GetCurrentScene struct {
	Scene
	*ResponseBase
}
