package services_test

import (
	. "github.com/llun/analytics/services"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"sort"
	"strings"
)

var _ = Describe("Ga", func() {

	var (
		service     GA
		mockNetwork MockNetwork
	)

	BeforeEach(func() {
		mockNetwork = MockNetwork{}
		service = GA{&mockNetwork, "tracking-id", "name"}
	})

	Context("#GetName", func() {

		It("should return GA as name", func() {
			name := service.GetName()
			Expect(name).To(Equal("name"))
		})

	})

	Context("#Send", func() {
		It("should do a request to ga api with post data", func() {
			in := GetMockInput()

			var output Output = service.Send(in)
			Expect(output.Success).To(BeTrue())
			Expect(mockNetwork.Data).ToNot(BeNil())
			Expect(mockNetwork.Url).To(Equal("http://www.google-analytics.com/collect"))
		})
	})

	Context("#FormatGAInput", func() {
		It("should return a GA payload", func() {
			in := GetMockInput()
			var (
				output string
			)
			output = service.FormatGAInput(in)
			data := "v=1&tid=TID&cid=CID&t=event&ec=android&ea=view_reward&el=game_key"
			dataStringArray := strings.Split(data, "&")
			sort.Strings(dataStringArray)
			outputStringArray := strings.Split(output, "&")
			sort.Strings(outputStringArray)

			Expect(outputStringArray).To(Equal(dataStringArray))
		})

	})

	Context("#GetConfiguration", func() {

		It("should return configuration as json with token", func() {

			service := GA{nil, "tracking-id", "ga"}
			config := service.GetConfiguration()
			Expect(config).To(Equal(map[string]interface{}{
				"tracking-id": "tracking-id",
			}))

		})

	})

	Context("#LoadConfiguration", func() {

		It("should apply new configuration to service", func() {

			service := GA{nil, "newid", "ga"}
			service.LoadConfiguration(map[string]interface{}{
				"tracking-id": "newid",
			})
			Expect(service.TrackingID).To(Equal("newid"))

		})

	})
})
