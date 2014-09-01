package services

type MockNetwork struct {
	Url  string
	Data string

	MockStatus int
	MockData   string
	MockError  error
}

func (m *MockNetwork) Request(url string, data string) (status int, body string, err error) {
	m.Url = url
	m.Data = data
	return m.MockStatus, m.MockData, m.MockError
}

func GetMockInput() Input {
	in := Input{
		Event: "view_reward",
		Data: map[string]interface{}{
			"reward": "Nexus5",
		},
		IP: "127.0.0.0",
	}
	return in
}

type MockService struct {
	Name string
	Data Input
}

func (m *MockService) Send(in Input) Output {
	m.Data = in
	return Output{true}
}

func (m *MockService) GetName() string {
	return m.Name
}

func (m *MockService) GetConfiguration() map[string]interface{} {
	return map[string]interface{}{}
}

func (m *MockService) LoadConfiguration(configuration map[string]interface{}) {
}

func (m *MockService) SetNetwork(network Network) {
}
