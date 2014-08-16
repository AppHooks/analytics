package services

import (
	"encoding/base64"
	"encoding/json"
)

type Mixpanel struct {
	Network Network
	Token   string
}

func (m *Mixpanel) Parse(in Input) ([]byte, error) {
	in.Data["token"] = m.Token
	in.Data["$ip"] = in.IP
	prepare := map[string]interface{}{
		"event":      in.Event,
		"properties": in.Data,
	}

	bytes, err := json.Marshal(prepare)
	return bytes, err
}

func (m *Mixpanel) GetConfiguration() map[string]interface{} {
	return map[string]interface{}{
		"token": m.Token,
	}
}

func (m *Mixpanel) LoadConfiguration(configuration map[string]interface{}) {
	m.Token = configuration["token"].(string)
}

func (m Mixpanel) GetName() string {
	return "MixPanel"
}

func (m Mixpanel) Send(input Input) Output {
	data, _ := m.Parse(input)
	status, resp, _ := m.Network.Request("https://api.mixpanel.com/track/", "data="+base64.StdEncoding.EncodeToString(data))
	return Output{Success: resp == "1" && status == 200}
}
