package hivesvc

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tanookiai/hive-svc/entity"
	accountsMock "github.com/tanookiai/hive-svc/mock/repository/authsvc/accounts"
	flathiveMock "github.com/tanookiai/hive-svc/mock/repository/db/flathive"
	hiveMock "github.com/tanookiai/hive-svc/mock/repository/db/hive"
	transactionsMock "github.com/tanookiai/hive-svc/mock/repository/db/transactions"
	producttokensMock "github.com/tanookiai/hive-svc/mock/repository/web3svc/producttokens"
	"github.com/tanookiai/hive-svc/repository/authsvc/accounts"
	"github.com/tanookiai/hive-svc/repository/db/flathive"
	"github.com/tanookiai/hive-svc/repository/db/hive"
	"github.com/tanookiai/hive-svc/repository/db/transactions"
	"github.com/tanookiai/hive-svc/repository/web3svc/producttokens"
)

func TestServiceHive_GetHiveMemberDetail(t *testing.T) {
	mockCtx := context.Background()
	mockTime := time.Date(2024, 10, 1, 0, 0, 0, 0, time.UTC)
	mockHiveRepo := new(hiveMock.IHiveRepository)
	mockFlatHiveRepo := new(flathiveMock.IFlatHiveRepository)
	mockAccountRepo := new(accountsMock.IAccountRepository)
	mockProductTokensRepo := new(producttokensMock.IProductTokensRepository)
	mockTransactionsRepo := new(transactionsMock.ITransactionRepository)

	tests := []struct {
		name    string
		request entity.GetHiveMemberDetailRequest
		mock    func()
		want    entity.GetHiveMemberDetailResponse
		wantErr bool
	}{
		{
			name: "Success",
			request: entity.GetHiveMemberDetailRequest{
				HiveID:        1,
				CurrentUserID: 2,
				IsTesting:     true,
			},
			mock: func() {
				beaconRules = func() map[string]string {
					return map[string]string{
						"bythen-pod": "10",
					}
				}
				getTierByBeaconPts = func(beaconPts int64, _ bool) (entity.Tier, error) {
					if beaconPts == 0 {
						return entity.Tier{}, nil
					}
					return entity.Tier{
						Tier: 1,
						Name: "New Bee",
					}, nil
				}
				mockHiveRepo.On("Get", mockCtx, hive.FilterHive{
					ID:        sql.NullInt64{Int64: 1, Valid: true},
					IsTesting: sql.NullBool{Bool: true, Valid: true},
				}).Return([]hive.Hive{
					{ID: 1, AccountID: 3, ReferrerAccountID: 4, BeaconPoints: 10, ActiveStatus: true},
				}, nil).Once()

				mockFlatHiveRepo.On("Get", mockCtx, flathive.FilterFlatHive{
					AccountID:         sql.NullInt64{Int64: 3, Valid: true},
					AncestorAccountID: sql.NullInt64{Int64: 2, Valid: true},
					IsTesting:         sql.NullBool{Bool: true, Valid: true},
				}).Return([]entity.FlatHive{
					{
						AccountID:         sql.NullInt64{Int64: 3, Valid: true},
						AncestorAccountID: sql.NullInt64{Int64: 2, Valid: true},
					},
				}, nil).Once()

				mockAccountRepo.On("GetMultipleAccounts", mockCtx, []uint64{3, 4}).Return([]accounts.Account{
					{Id: 3, WalletPublicKey: "0x742d35Cc6634C0532925a3b844Bc454e4438f44e", Username: "user3"},
					{Id: 4, WalletPublicKey: "0xFE3B557E8Fb62b89F4916B721be55cEb828dBd73", Username: "user4"},
				}, nil).Once()

				mockProductTokensRepo.On("GetProductTokens", mockCtx, producttokens.ProductTokenFields{
					OwnerAddress: "0x742d35Cc6634C0532925a3b844Bc454e4438f44e", // #gitleaks:allow
					Products:     []string{"bythen-chip", "bythen-pod", "bythen-card", "bythen-type01"},
					IsTesting:    true,
				}).Return([]producttokens.ProductToken{
					{
						TokenId: 2,
						Product: producttokens.ProductTokenProduct{
							Id:   1,
							Slug: "bythen-pod",
						},
					},
					{
						TokenId: 3,
						Product: producttokens.ProductTokenProduct{
							Id:   1,
							Slug: "bythen-pod",
						},
					},
					{
						TokenId: 3,
						Product: producttokens.ProductTokenProduct{
							Id:   6,
							Slug: "mystery-pod",
						},
					},
				}, nil).Once()

				mockTransactionsRepo.On("List", mockCtx, transactions.FilterTransaction{
					AccountId: sql.NullInt64{Int64: 3, Valid: true},
					Address:   sql.NullString{String: "0x742d35Cc6634C0532925a3b844Bc454e4438f44e", Valid: true}, // #gitleaks:allow
					SortBy:    "trx_at",
					SortMode:  "DESC",
					Limit:     1,
				}).Return([]transactions.Transaction{
					{Id: 1, AccountId: 3, Address: "0x742d35Cc6634C0532925a3b844Bc454e4438f44e", TrxAt: mockTime}, // #gitleaks:allow
				}, nil).Once()
			},
			want: entity.GetHiveMemberDetailResponse{
				HiveID:                 1,
				AccountID:              3,
				AccountWalletPublicKey: "0x742d...f44e",
				Username:               "user3",
				ReferrerAccountID:      4,
				ReferrerUsername:       "user4",
				BeaconPoints:           20,
				Tier:                   1,
				TierName:               "New Bee",
				ActiveStatus:           true,
				LastPurchaseAt:         &mockTime,
			},
			wantErr: false,
		},
		{
			name: "Success_get_detail_current_user",
			request: entity.GetHiveMemberDetailRequest{
				HiveID:        1,
				CurrentUserID: 3,
				IsTesting:     true,
			},
			mock: func() {
				beaconRules = func() map[string]string {
					return map[string]string{
						"bythen-pod": "10",
					}
				}
				getTierByBeaconPts = func(beaconPts int64, _ bool) (entity.Tier, error) {
					if beaconPts == 0 {
						return entity.Tier{}, nil
					}
					return entity.Tier{
						Tier: 1,
						Name: "New Bee",
					}, nil
				}
				mockHiveRepo.On("Get", mockCtx, hive.FilterHive{
					ID:        sql.NullInt64{Int64: 1, Valid: true},
					IsTesting: sql.NullBool{Bool: true, Valid: true},
				}).Return([]hive.Hive{
					{ID: 1, AccountID: 3, ReferrerAccountID: 4, BeaconPoints: 10, ActiveStatus: true},
				}, nil).Once()

				mockAccountRepo.On("GetMultipleAccounts", mockCtx, []uint64{3, 4}).Return([]accounts.Account{
					{Id: 3, WalletPublicKey: "0x742d35Cc6634C0532925a3b844Bc454e4438f44e", Username: "user3"}, // #gitleaks:allow
					{Id: 4, WalletPublicKey: "0xFE3B557E8Fb62b89F4916B721be55cEb828dBd73", Username: "user4"}, // #gitleaks:allow
				}, nil).Once()

				mockProductTokensRepo.On("GetProductTokens", mockCtx, producttokens.ProductTokenFields{
					OwnerAddress: "0x742d35Cc6634C0532925a3b844Bc454e4438f44e", // #gitleaks:allow
					Products:     []string{"bythen-chip", "bythen-pod", "bythen-card", "bythen-type01"},
					IsTesting:    true,
				}).Return([]producttokens.ProductToken{
					{
						TokenId: 2,
						Product: producttokens.ProductTokenProduct{
							Id:   1,
							Slug: "bythen-pod",
						},
					},
					{
						TokenId: 3,
						Product: producttokens.ProductTokenProduct{
							Id:   1,
							Slug: "bythen-pod",
						},
					},
					{
						TokenId: 3,
						Product: producttokens.ProductTokenProduct{
							Id:   6,
							Slug: "mystery-pod",
						},
					},
				}, nil).Once()

				mockTransactionsRepo.On("List", mockCtx, transactions.FilterTransaction{
					AccountId: sql.NullInt64{Int64: 3, Valid: true},
					Address:   sql.NullString{String: "0x742d35Cc6634C0532925a3b844Bc454e4438f44e", Valid: true}, // #gitleaks:allow
					SortBy:    "trx_at",
					SortMode:  "DESC",
					Limit:     1,
				}).Return([]transactions.Transaction{}, nil).Once()
			},
			want: entity.GetHiveMemberDetailResponse{
				HiveID:                 1,
				AccountID:              3,
				AccountWalletPublicKey: "0x742d...f44e",
				Username:               "user3",
				ReferrerAccountID:      4,
				ReferrerUsername:       "user4",
				BeaconPoints:           20,
				Tier:                   1,
				TierName:               "New Bee",
				ActiveStatus:           true,
			},
			wantErr: false,
		},
		{
			name: "Success_trial_user",
			request: entity.GetHiveMemberDetailRequest{
				HiveID:        1,
				CurrentUserID: 3,
				IsTesting:     true,
			},
			mock: func() {
				timeNow = func() time.Time {
					return mockTime
				}
				beaconRules = func() map[string]string {
					return map[string]string{
						"bythen-pod": "10",
					}
				}
				getTierByBeaconPts = func(beaconPts int64, isTrial bool) (entity.Tier, error) {
					if beaconPts == 0 && !isTrial {
						return entity.Tier{}, nil
					}
					return entity.Tier{
						Tier: 1,
						Name: "New Bee",
					}, nil
				}
				mockHiveRepo.On("Get", mockCtx, hive.FilterHive{
					ID:        sql.NullInt64{Int64: 1, Valid: true},
					IsTesting: sql.NullBool{Bool: true, Valid: true},
				}).Return([]hive.Hive{
					{ID: 1, AccountID: 3, BeaconPoints: 0, ActiveStatus: true, TrialEndedAt: sql.NullTime{Time: mockTime.AddDate(0, 0, 14), Valid: true}},
				}, nil).Once()

				mockAccountRepo.On("GetMultipleAccounts", mockCtx, []uint64{3, 0}).Return([]accounts.Account{
					{Id: 3, WalletPublicKey: "0x742d35Cc6634C0532925a3b844Bc454e4438f44e", Username: "user3"}, // #gitleaks:allow
				}, nil).Once()

				mockProductTokensRepo.On("GetProductTokens", mockCtx, producttokens.ProductTokenFields{
					OwnerAddress: "0x742d35Cc6634C0532925a3b844Bc454e4438f44e", // #gitleaks:allow
					Products:     []string{"bythen-chip", "bythen-pod", "bythen-card", "bythen-type01"},
					IsTesting:    true,
				}).Return([]producttokens.ProductToken{}, nil).Once()

				mockTransactionsRepo.On("List", mockCtx, transactions.FilterTransaction{
					AccountId: sql.NullInt64{Int64: 3, Valid: true},
					Address:   sql.NullString{String: "0x742d35Cc6634C0532925a3b844Bc454e4438f44e", Valid: true}, // #gitleaks:allow
					SortBy:    "trx_at",
					SortMode:  "DESC",
					Limit:     1,
				}).Return([]transactions.Transaction{}, nil).Once()
			},
			want: entity.GetHiveMemberDetailResponse{
				HiveID:                 1,
				AccountID:              3,
				AccountWalletPublicKey: "0x742d...f44e",
				Username:               "user3",
				ReferrerAccountID:      0,
				ReferrerUsername:       "",
				BeaconPoints:           0,
				Tier:                   1,
				TierName:               "New Bee",
				ActiveStatus:           true,
				IsTrial:                true,
			},
			wantErr: false,
		},
		{
			name: "Failed_HiveID_not_found",
			request: entity.GetHiveMemberDetailRequest{
				HiveID:        1,
				CurrentUserID: 2,
				IsTesting:     true,
			},
			mock: func() {
				mockHiveRepo.On("Get", mockCtx, hive.FilterHive{
					ID:        sql.NullInt64{Int64: 1, Valid: true},
					IsTesting: sql.NullBool{Bool: true, Valid: true},
				}).Return([]hive.Hive{}, nil).Once()
			},
			want:    entity.GetHiveMemberDetailResponse{},
			wantErr: true,
		},
		{
			name: "Failed_HiveID_not_descendant",
			request: entity.GetHiveMemberDetailRequest{
				HiveID:        1,
				CurrentUserID: 2,
				IsTesting:     true,
			},
			mock: func() {
				mockHiveRepo.On("Get", mockCtx, hive.FilterHive{
					ID:        sql.NullInt64{Int64: 1, Valid: true},
					IsTesting: sql.NullBool{Bool: true, Valid: true},
				}).Return([]hive.Hive{
					{ID: 1, AccountID: 3, ReferrerAccountID: 4, BeaconPoints: 100, ActiveStatus: true},
				}, nil).Once()

				mockFlatHiveRepo.On("Get", mockCtx, flathive.FilterFlatHive{
					AccountID:         sql.NullInt64{Int64: 3, Valid: true},
					AncestorAccountID: sql.NullInt64{Int64: 2, Valid: true},
					IsTesting:         sql.NullBool{Bool: true, Valid: true},
				}).Return([]entity.FlatHive{}, nil).Once()
			},
			want:    entity.GetHiveMemberDetailResponse{},
			wantErr: true,
		},
		{
			name: "Failed_error_getting_hive_members",
			request: entity.GetHiveMemberDetailRequest{
				HiveID:        1,
				CurrentUserID: 2,
				IsTesting:     true,
			},
			mock: func() {
				mockHiveRepo.On("Get", mockCtx, hive.FilterHive{
					ID:        sql.NullInt64{Int64: 1, Valid: true},
					IsTesting: sql.NullBool{Bool: true, Valid: true},
				}).Return(nil, errors.New("error in getting hive members")).Once()
			},
			want:    entity.GetHiveMemberDetailResponse{},
			wantErr: true,
		},
		{
			name: "Failed_error_getting_flat_hive",
			request: entity.GetHiveMemberDetailRequest{
				HiveID:        1,
				CurrentUserID: 2,
				IsTesting:     true,
			},
			mock: func() {
				mockHiveRepo.On("Get", mockCtx, hive.FilterHive{
					ID:        sql.NullInt64{Int64: 1, Valid: true},
					IsTesting: sql.NullBool{Bool: true, Valid: true},
				}).Return([]hive.Hive{
					{ID: 1, AccountID: 3, ReferrerAccountID: 4, BeaconPoints: 100, ActiveStatus: true},
				}, nil).Once()

				mockFlatHiveRepo.On("Get", mockCtx, flathive.FilterFlatHive{
					AccountID:         sql.NullInt64{Int64: 3, Valid: true},
					AncestorAccountID: sql.NullInt64{Int64: 2, Valid: true},
					IsTesting:         sql.NullBool{Bool: true, Valid: true},
				}).Return(nil, errors.New("error in getting flat hive")).Once()
			},
			want:    entity.GetHiveMemberDetailResponse{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()

			s := &ServiceHive{
				hiveRepo:          mockHiveRepo,
				flatHiveRepo:      mockFlatHiveRepo,
				accountRepo:       mockAccountRepo,
				transactionsRepo:  mockTransactionsRepo,
				productTokensRepo: mockProductTokensRepo,
			}

			got, err := s.GetHiveMemberDetail(mockCtx, tt.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("ServiceHive.GetHiveMemberDetail() error = %v, wantErr %v", err, tt.wantErr)
			}

			assert.Equal(t, tt.want, got)
		})
	}
}
