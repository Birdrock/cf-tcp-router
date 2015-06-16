package handlers_test

import (
	"errors"
	"net/http"
	"net/http/httptest"

	cf_tcp_router "github.com/cloudfoundry-incubator/cf-tcp-router"
	"github.com/cloudfoundry-incubator/cf-tcp-router/configurer/fakes"
	"github.com/cloudfoundry-incubator/cf-tcp-router/handlers"
	"github.com/pivotal-golang/lager"
	"github.com/pivotal-golang/lager/lagertest"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
)

var _ = Describe("ExternalPortMapHandler", func() {
	var (
		handler          *handlers.ExternalPortMapHandler
		logger           lager.Logger
		responseRecorder *httptest.ResponseRecorder
		fakeConfigurer   *fakes.FakeRouterConfigurer
	)

	BeforeEach(func() {
		logger = lagertest.NewTestLogger("test")
		responseRecorder = httptest.NewRecorder()
		fakeConfigurer = new(fakes.FakeRouterConfigurer)
		handler = handlers.NewExternalPortMapHandler(logger, fakeConfigurer)
	})

	Describe("MapExternalPort", func() {
		var (
			mappingRequest cf_tcp_router.MappingRequests
		)
		BeforeEach(func() {
			backendHostInfo := cf_tcp_router.NewBackendHostInfo("1.2.3.4", 1234)
			backendHostInfos := cf_tcp_router.BackendHostInfos{backendHostInfo}
			mappingRequest = cf_tcp_router.MappingRequests{
				cf_tcp_router.NewMappingRequest(1234, backendHostInfos),
			}
		})

		JustBeforeEach(func() {
			handler.MapExternalPort(responseRecorder, newTestRequest(mappingRequest))
		})

		Context("when request is valid", func() {
			BeforeEach(func() {
				fakeConfigurer.CreateExternalPortMappingsReturns(nil)
			})

			It("responds with 200 CREATED", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusOK))
			})

			Context("when configurer returns an error", func() {
				BeforeEach(func() {
					fakeConfigurer.CreateExternalPortMappingsReturns(errors.New("Kabooom"))
				})

				It("responds with 500 INTERNAL_SERVER_ERROR", func() {
					Expect(responseRecorder.Code).To(Equal(http.StatusInternalServerError))
					Eventually(logger).Should(gbytes.Say("test.map_external_port.failed-to-configure"))
				})
			})
		})

		Context("when request is invalid", func() {
			Context("when payload is not a valid json", func() {
				BeforeEach(func() {
					handler.MapExternalPort(responseRecorder, newTestRequest(`{abcd`))
				})

				It("responds with 400 BAD REQUEST", func() {
					Expect(responseRecorder.Code).To(Equal(http.StatusBadRequest))
					Eventually(logger).Should(gbytes.Say("test.map_external_port.failed-to-unmarshal"))
				})
			})

			Context("when payload has invalid values", func() {
				BeforeEach(func() {
					backendHostInfo := cf_tcp_router.NewBackendHostInfo("1.2.3.4", 0)
					backendHostInfos := cf_tcp_router.BackendHostInfos{backendHostInfo}
					mappingRequest = cf_tcp_router.MappingRequests{
						cf_tcp_router.NewMappingRequest(1234, backendHostInfos),
					}
					fakeConfigurer.CreateExternalPortMappingsReturns(errors.New(cf_tcp_router.ErrInvalidMapingRequest))
				})

				It("responds with 400 BAD REQUEST", func() {
					Expect(responseRecorder.Code).To(Equal(http.StatusBadRequest))
					Eventually(logger).Should(gbytes.Say("test.map_external_port.invalid-payload"))
				})
			})
		})
	})
})
