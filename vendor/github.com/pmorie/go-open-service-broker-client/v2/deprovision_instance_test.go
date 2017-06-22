package v2

import (
	"fmt"
	"net/http"
	"testing"
)

func defaultDeprovisionRequest() *DeprovisionRequest {
	return &DeprovisionRequest{
		InstanceID: testInstanceID,
		ServiceID:  testServiceID,
		PlanID:     testPlanID,
	}
}

func defaultAsyncDeprovisionRequest() *DeprovisionRequest {
	r := defaultDeprovisionRequest()
	r.AcceptsIncomplete = true
	return r
}

const successDeprovisionResponseBody = `{}`

func successDeprovisionResponse() *DeprovisionResponse {
	return &DeprovisionResponse{}
}

const successAsyncDeprovisionResponseBody = `{
  "operation": "test-operation-key"
}`

func successDeprovisionResponseAsync() *DeprovisionResponse {
	r := successDeprovisionResponse()
	r.Async = true
	r.OperationKey = &testOperation
	return r
}

func TestDeprovisionInstance(t *testing.T) {
	cases := []struct {
		name               string
		enableAlpha        bool
		request            *DeprovisionRequest
		httpChecks         httpChecks
		httpReaction       httpReaction
		expectedResponse   *DeprovisionResponse
		expectedErrMessage string
		expectedErr        error
	}{
		{
			name: "invalid request",
			request: func() *DeprovisionRequest {
				r := defaultDeprovisionRequest()
				r.InstanceID = ""
				return r
			}(),
			expectedErrMessage: "instanceID is required",
		},
		{
			name: "success - ok",
			httpReaction: httpReaction{
				status: http.StatusOK,
				body:   successDeprovisionResponseBody,
			},
			expectedResponse: successDeprovisionResponse(),
		},
		{
			name: "success - gone",
			httpReaction: httpReaction{
				status: http.StatusGone,
				body:   successDeprovisionResponseBody,
			},
			expectedResponse: successDeprovisionResponse(),
		},
		{
			name:    "success - async",
			request: defaultAsyncDeprovisionRequest(),
			httpChecks: httpChecks{
				params: map[string]string{
					asyncQueryParamKey: "true",
				},
			},
			httpReaction: httpReaction{
				status: http.StatusAccepted,
				body:   successAsyncDeprovisionResponseBody,
			},
			expectedResponse: successDeprovisionResponseAsync(),
		},
		{
			name:    "accepted with malformed response",
			request: defaultAsyncDeprovisionRequest(),
			httpChecks: httpChecks{
				params: map[string]string{
					asyncQueryParamKey: "true",
				},
			},
			httpReaction: httpReaction{
				status: http.StatusAccepted,
				body:   malformedResponse,
			},
			expectedErrMessage: "unexpected end of JSON input",
		},
		{
			name: "http error",
			httpReaction: httpReaction{
				err: fmt.Errorf("http error"),
			},
			expectedErrMessage: "http error",
		},
		{
			name: "200 with malformed response",
			httpReaction: httpReaction{
				status: http.StatusOK,
				body:   malformedResponse,
			},
			expectedResponse: successDeprovisionResponse(),
		},
		{
			name: "500 with malformed response",
			httpReaction: httpReaction{
				status: http.StatusInternalServerError,
				body:   malformedResponse,
			},
			expectedErrMessage: "unexpected end of JSON input",
		},
		{
			name: "500 with conventional failure response",
			httpReaction: httpReaction{
				status: http.StatusInternalServerError,
				body:   conventionalFailureResponseBody,
			},
			expectedErr: testHttpStatusCodeError(),
		},
	}

	for _, tc := range cases {
		if tc.request == nil {
			tc.request = defaultDeprovisionRequest()
		}

		if tc.httpChecks.URL == "" {
			tc.httpChecks.URL = "/v2/service_instances/test-instance-id"
		}

		if tc.httpChecks.body == "" {
			tc.httpChecks.body = "{}"
		}

		klient := newTestClient(t, tc.name, tc.enableAlpha, tc.httpChecks, tc.httpReaction)

		response, err := klient.DeprovisionInstance(tc.request)

		doResponseChecks(t, tc.name, response, err, tc.expectedResponse, tc.expectedErrMessage, tc.expectedErr)
	}
}

func TestValidateDeprovisionRequest(t *testing.T) {
	cases := []struct {
		name    string
		request *DeprovisionRequest
		valid   bool
	}{
		{
			name:    "valid",
			request: defaultDeprovisionRequest(),
			valid:   true,
		},
		{
			name: "missing instance ID",
			request: func() *DeprovisionRequest {
				r := defaultDeprovisionRequest()
				r.InstanceID = ""
				return r
			}(),
			valid: false,
		},
		{
			name: "missing service ID",
			request: func() *DeprovisionRequest {
				r := defaultDeprovisionRequest()
				r.ServiceID = ""
				return r
			}(),
			valid: false,
		},
		{
			name: "missing plan ID",
			request: func() *DeprovisionRequest {
				r := defaultDeprovisionRequest()
				r.PlanID = ""
				return r
			}(),
			valid: false,
		},
	}

	for _, tc := range cases {
		err := validateDeprovisionRequest(tc.request)
		if err != nil {
			if tc.valid {
				t.Errorf("%v: expected valid, got error: %v", tc.name, err)
			}
		} else if !tc.valid {
			t.Errorf("%v: expected invalid, got valid", tc.name)
		}
	}
}