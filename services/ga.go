package services

import (
	"bytes"
	"fmt"
)

type GA struct {
	Network    Network
	TrackingID string
	Name       string
}

func (i *GA) GetName() string {
	return i.Name
}

func (i GA) Send(in Input) Output {
	payload := i.FormatGAInput(in)
	i.Network.Request("http://www.google-analytics.com/collect", payload)
	return Output{true}
}

func (i GA) FormatGAInput(in Input) string {
	prepare := map[string]interface{}{
		"cid": "CID",
		"ea":  in.Event,
		"ec":  "android",
		"el":  "game_key",
		"t":   "event",
		"tid": "TID",
		"v":   "1",
	}

	var buffer bytes.Buffer
	for key, value := range prepare {
		buffer.WriteString(fmt.Sprintf("%s=%s&", key, value))
	}

	var output string = buffer.String()
	return output[0 : len(output)-1]
}

func (i *GA) SetNetwork(network Network) {
	i.Network = network
}

func (i *GA) GetConfiguration() map[string]interface{} {
	return map[string]interface{}{
		"tracking-id": i.TrackingID,
	}
}

func (i *GA) LoadConfiguration(configuration map[string]interface{}) {
	i.TrackingID = configuration["tracking-id"].(string)
}
