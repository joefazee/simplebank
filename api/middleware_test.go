package api

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joefazee/simplebank/token"
	"github.com/stretchr/testify/require"
)

func addAuthorization(
	t *testing.T,
	req *http.Request,
	tokenMaker token.Maker,
	authorizationType string,
	username string,
	duration time.Duration,
){

	token, payload, err := tokenMaker.CreateToken(username, duration)
	require.NoError(t, err)
	require.NotEmpty(t, payload)

	authorizationHeader := fmt.Sprintf("%s %s", authorizationType, token)
	req.Header.Set(authorizationHeaderKey, authorizationHeader)
}

func TestAuthMiddleware(t *testing.T) {
	testCases := []struct{
		name string
		setupAuth func (t *testing.T, req *http.Request, tokenMaker token.Maker)
		checkResponse  func(t *testing.T, rr *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, req, tokenMaker, authorizationTypeBearer,  "user", time.Minute)
			},
			checkResponse: func(t *testing.T, rr *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, rr.Code)
			},
		},

		{
			name: "No authorization",
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker token.Maker) {
			},
			checkResponse: func(t *testing.T, rr *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, rr.Code)
			},
		},

		{
			name: "Unsupported authorization",
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, req, tokenMaker, "unsupported",  "user", time.Minute)
			},
			checkResponse: func(t *testing.T, rr *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, rr.Code)
			},
		},

		{
			name: "Invalid authorization format",
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, req, tokenMaker, "",  "user", time.Minute)
			},
			checkResponse: func(t *testing.T, rr *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, rr.Code)
			},
		},

		{
			name: "token expired",
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, req, tokenMaker, authorizationTypeBearer,  "user", -time.Minute)
			},
			checkResponse: func(t *testing.T, rr *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, rr.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			server := newTestServer(t, nil)
			authPath := "/auth"
			server.router.GET(authPath, authMiddleware(server.tokenMaker), func(ctx *gin.Context) {
				ctx.JSON(http.StatusOK, gin.H{})
			})

			rr := httptest.NewRecorder()
			req, err := http.NewRequest(http.MethodGet, authPath, nil)
			require.NoError(t, err)

			tc.setupAuth(t, req, server.tokenMaker)
			server.router.ServeHTTP(rr, req)
			tc.checkResponse(t, rr)
		})

		
	}
}