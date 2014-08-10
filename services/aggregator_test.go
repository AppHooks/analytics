package services_test

import (
	"encoding/json"
	. "github.com/llun/analytics/services"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

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

var _ = Describe("Aggregator", func() {

	Context("#Send", func() {

		It("should create input and send to every services under it", func() {
			services := []Service{
				&MockService{"service1", Input{}},
				&MockService{"service2", Input{}},
			}
			aggregator := Aggregator{services}

			prepare := make(map[string]interface{})
			data := []byte(`{"event":"view","data":{"key1":"value1","key2":"value2"}}`)
			json.Unmarshal(data, &prepare)

			output := aggregator.Send(prepare, "127.0.0.1")

			expect := map[string]interface{}{
				"success": true,
				"services": map[string]interface{}{
					"service1": true,
					"service2": true,
				},
			}
			bytes, _ := json.Marshal(expect)
			Expect(string(bytes)).To(Equal(output))
		})

		It("should forward client ip to services", func() {

			mockService := MockService{Name: "service"}
			services := []Service{&mockService}
			aggregator := Aggregator{services}

			prepare := make(map[string]interface{})
			data := []byte(`{"event":"view","data":{"key1":"value1","key2":"value2"}}`)
			json.Unmarshal(data, &prepare)

			aggregator.Send(prepare, "127.0.0.1")

			Expect(mockService.Data.IP).To(Equal("127.0.0.1"))

		})

	})

})
