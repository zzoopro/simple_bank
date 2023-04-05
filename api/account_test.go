package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	mockdb "github.com/zzoopro/simple_bank/db/mock"
	db "github.com/zzoopro/simple_bank/db/sqlc"
	"github.com/zzoopro/simple_bank/util"
)

func TestGetAccount(t *testing.T) { 
	accout := randomAccount()

	testCases := []struct{
		name string
		accountID int64
		buildStubs func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			accountID: accout.ID,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
				GetAccount(gomock.Any(), gomock.Eq(accout.ID)).
				Times(1).
				Return(accout, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchAccount(t, recorder.Body, accout)
			},
		},
		{
			name: "NotFound",
			accountID: accout.ID,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
				GetAccount(gomock.Any(), gomock.Eq(accout.ID)).
				Times(1).
				Return(db.Account{}, sql.ErrNoRows)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)				
			},
		},
		{
			name: "InternalError",
			accountID: accout.ID,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
				GetAccount(gomock.Any(), gomock.Eq(accout.ID)).
				Times(1).
				Return(db.Account{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)				
			},
		},
		{
			name: "InvalidID",
			accountID: 0,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
				GetAccount(gomock.Any(), gomock.Any()).
				Times(0).
				Return(db.Account{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)				
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			tc.buildStubs(store)

			// start test server and send request
			server := NewServer(store)
			recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/accounts/%d", tc.accountID)
			request, err :=http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)
			
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}	
}

func randomAccount() db.Account {
	return db.Account{
		ID: util.RandomInt(1, 1000),
		Owner: util.RandomOwner(),
		Balance: util.RandomMoney(),
		Currency: util.RandomCurrency(),
	}
}

func requireBodyMatchAccount(t *testing.T, body *bytes.Buffer, account db.Account) {
	data, err := ioutil.ReadAll(body)
	require.NoError(t, err)

	var gotAccount db.Account
	err = json.Unmarshal(data, &gotAccount)
	require.NoError(t, err)
	require.Equal(t, account, gotAccount)
}