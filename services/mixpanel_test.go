package services_test

import (
	"encoding/json"
	. "github.com/llun/analytics/services"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Mixpanel", func() {

	var (
		service     Mixpanel
		mockNetwork MockNetwork
	)

	Context("#GetName", func() {

		BeforeEach(func() {
			mockNetwork = MockNetwork{}
			service = Mixpanel{&mockNetwork, "token"}
		})

		It("should return mixpanel as name", func() {
			name := service.GetName()
			Expect(name).To(Equal("MixPanel"))
		})
	})

	Context("#Send", func() {

		It("should do a get request mixpanel api with base64 encoding data", func() {

			mockNetwork = MockNetwork{
				MockStatus: 200,
				MockData:   "1",
				MockError:  nil,
			}
			service = Mixpanel{&mockNetwork, "token"}

			in := GetMockInput()
			output := service.Send(in)

			Expect(output.Success).To(Equal(true))

			Expect(mockNetwork.Url).To(Equal("https://api.mixpanel.com/track/"))
			Expect(mockNetwork.Data).To(ContainSubstring("data="))
		})

		It("should return error when Mixpanel returns 0", func() {
			mockNetwork = MockNetwork{
				MockStatus: 200,
				MockData:   "0",
				MockError:  nil,
			}
			service = Mixpanel{&mockNetwork, "token"}

			in := GetMockInput()
			output := service.Send(in)

			Expect(output.Success).To(Equal(false))
		})

		It("should return error when Mixpanel goes down", func() {
			mockNetwork = MockNetwork{
				MockStatus: 503,
				MockData:   "",
				MockError:  nil,
			}
			service = Mixpanel{&mockNetwork, "token"}

			in := GetMockInput()
			output := service.Send(in)

			Expect(output.Success).To(Equal(false))
		})

	})

	Context("#Parse", func() {

		BeforeEach(func() {
			mockNetwork = MockNetwork{}
			service = Mixpanel{&mockNetwork, "token"}
		})

		It("should return JSON string with mixpanel token and request ip", func() {

			in := GetMockInput()
			output, _ := service.Parse(in)

			expect := map[string]interface{}{
				"event": "view_reward",
				"properties": map[string]interface{}{
					"reward": "Nexus5",
					"token":  "token",
					"$ip":    in.IP,
				},
			}
			data, _ := json.Marshal(expect)
			Expect(output).To(Equal(data))

		})

	})

	Context("#GetConfiguration", func() {

		It("should return configuration as json with token", func() {

			service := Mixpanel{nil, "token"}
			config := service.GetConfiguration()
			Expect(config).To(Equal(map[string]interface{}{
				"token": "token",
			}))

		})

	})

	Context("#LoadConfiguration", func() {

		It("should apply new configuration to service", func() {

			service := Mixpanel{nil, "token"}
			service.LoadConfiguration(map[string]interface{}{
				"token": "newtoken",
			})
			Expect(service.Token).To(Equal("newtoken"))

		})

	})

})
