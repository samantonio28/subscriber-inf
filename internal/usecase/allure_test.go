//go:build allure

package usecase

// Allure-тесты ЛР1 для 10 юзкейсов. На каждый юзкейс — два сценария:
//   - «Лондон»: доменные репозитории заменены gomock-моками (internal/mocks);
//   - «Классика»: настоящие репозитории internal/service поверх FakePgPool
//     («замоканный PostgreSQL», internal/testutil/pgxstubs.go).
//
// Запуск: make allure-test (go test -tags allure ...). Результаты пишутся
// в каталог ALLURE_OUTPUT_FOLDER, просмотр: make allure-serve.

import (
	"context"
	"testing"
	"time"

	"go.uber.org/mock/gomock"
	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/ozontech/allure-go/pkg/framework/runner"
	"github.com/samantonio28/subscriber-inf/internal/domain"
	mock "github.com/samantonio28/subscriber-inf/internal/mocks"
	"github.com/samantonio28/subscriber-inf/internal/service"
	"github.com/samantonio28/subscriber-inf/internal/testutil"
)

// ---------------------------------------------------------------------------
// 1. CreateSub
// ---------------------------------------------------------------------------

func TestAllureCreateSub(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	runner.Run(t, "CreateSub — Лондонский стиль", func(pt provider.T) {
		pt.Epic("Usecases")
		pt.Feature("CreateSub")
		pt.Tags("london", "unit", "subscriptions")
		pt.Title("Создание подписки через мок SubscriptionRepository")

		ctx := context.Background()
		mockSub := mock.NewMockSubscriptionRepository(ctrl)
		mockSub.EXPECT().StoreSub(ctx, gomock.Any()).Return(domain.SubID(42), nil)

		uc, err := NewCreateSubUC(mockSub, &testutil.NopLogger{})
		pt.Require().NoError(err)

		input := SubscriptionDTO{
			UserId:      testutil.TestUUID(),
			ServiceName: "Netflix",
			Price:       1000,
			SubType:     "usual",
			StartDate:   time.Now(),
			EndDate:     time.Now().AddDate(0, 1, 0),
			PlanID:      1,
		}

		id, err := uc.NewSub(ctx, input)
		pt.Require().NoError(err)
		pt.Assert().Equal(42, id)
	})

	runner.Run(t, "CreateSub — Классический стиль", func(pt provider.T) {
		pt.Epic("Usecases")
		pt.Feature("CreateSub")
		pt.Tags("classic", "unit", "subscriptions", "fake-pg")
		pt.Title("Создание подписки: настоящий SubRepo над FakePgPool")

		ctx := context.Background()
		fakePG := testutil.NewFakePgPool()
		fakePG.Tx().OnQueryRow("RETURNING sub_id", testutil.NewStubRow(42))

		subRepo, err := service.NewSubRepo(fakePG)
		pt.Require().NoError(err)
		uc, err := NewCreateSubUC(subRepo, &testutil.NopLogger{})
		pt.Require().NoError(err)

		input := SubscriptionDTO{
			UserId:      testutil.TestUUID(),
			ServiceName: "Netflix",
			Price:       1000,
			SubType:     "usual",
			StartDate:   time.Now(),
			EndDate:     time.Now().AddDate(0, 1, 0),
			PlanID:      1,
		}

		id, err := uc.NewSub(ctx, input)
		pt.Require().NoError(err)
		pt.Assert().Equal(42, id)
	})
}

// ---------------------------------------------------------------------------
// 2. GetSub
// ---------------------------------------------------------------------------

func TestAllureGetSub(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	runner.Run(t, "GetSub — Лондонский стиль", func(pt provider.T) {
		pt.Epic("Usecases")
		pt.Feature("GetSub")
		pt.Tags("london", "unit", "subscriptions")
		pt.Title("Получение подписки по id через мок")

		ctx := context.Background()
		userID := testutil.TestUUID()
		sub := domain.Subscription{
			SubId: 7, UserID: userID, ServiceName: "Netflix", Price: 1000,
			SubType: domain.SubTypeUsual, StartDate: time.Now(), EndDate: time.Now().AddDate(0, 1, 0), PlanID: 1,
		}
		mockSub := mock.NewMockSubscriptionRepository(ctrl)
		mockSub.EXPECT().Sub(ctx, domain.SubID(7)).Return(sub, nil)

		uc, _ := NewGetSubUC(mockSub, &testutil.NopLogger{})
		dto, err := uc.SubById(ctx, 7)
		pt.Require().NoError(err)
		pt.Assert().Equal(7, dto.SubId)
		pt.Assert().Equal("Netflix", dto.ServiceName)
	})

	runner.Run(t, "GetSub — Классический стиль", func(pt provider.T) {
		pt.Epic("Usecases")
		pt.Feature("GetSub")
		pt.Tags("classic", "unit", "subscriptions", "fake-pg")
		pt.Title("Получение подписки: Sub() сканирует 9 полей из StubRow")

		ctx := context.Background()
		userID := testutil.TestUUID()
		start := time.Now()
		end := time.Now().AddDate(0, 1, 0)

		fakePG := testutil.NewFakePgPool()
		fakePG.Tx().OnQueryRow("SELECT s.sub_id", stubSubRow(7, userID, 1, 1000, "usual", start, end))

		subRepo, _ := service.NewSubRepo(fakePG)
		uc, _ := NewGetSubUC(subRepo, &testutil.NopLogger{})
		dto, err := uc.SubById(ctx, 7)
		pt.Require().NoError(err)
		pt.Assert().Equal(7, dto.SubId)
		pt.Assert().Equal(1000, dto.Price)
	})
}

// ---------------------------------------------------------------------------
// 3. GetSubs
// ---------------------------------------------------------------------------

func TestAllureGetSubs(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	runner.Run(t, "GetSubs — Лондонский стиль", func(pt provider.T) {
		pt.Epic("Usecases")
		pt.Feature("GetSubs")
		pt.Tags("london", "unit", "subscriptions")
		pt.Title("Список подписок пользователя через мок")

		ctx := context.Background()
		userID := testutil.TestUUID()
		sub := domain.Subscription{
			SubId: 1, UserID: userID, ServiceName: "Netflix", Price: 500,
			SubType: domain.SubTypeUsual, StartDate: time.Now(), EndDate: time.Now().AddDate(0, 1, 0), PlanID: 1,
		}
		mockSub := mock.NewMockSubscriptionRepository(ctrl)
		mockSub.EXPECT().UserSubs(ctx, userID).Return([]domain.Subscription{sub}, nil)

		uc, _ := NewGetSubsUC(mockSub, &testutil.NopLogger{})
		subs, err := uc.SubsByUserId(ctx, userID)
		pt.Require().NoError(err)
		pt.Assert().Equal(1, len(subs))
		pt.Assert().Equal(1, subs[0].SubId)
	})

	runner.Run(t, "GetSubs — Классический стиль", func(pt provider.T) {
		pt.Epic("Usecases")
		pt.Feature("GetSubs")
		pt.Tags("classic", "unit", "subscriptions", "fake-pg")
		pt.Title("Список подписок: Query(ids) + вложенные Sub()")

		ctx := context.Background()
		userID := testutil.TestUUID()
		start := time.Now()
		end := time.Now().AddDate(0, 1, 0)

		fakePG := testutil.NewFakePgPool()
		fakePG.SetQuery("SELECT sub_id FROM subscriptions", testutil.NewStubRows([]any{1}))
		fakePG.Tx().OnQueryRow("SELECT s.sub_id", stubSubRow(1, userID, 1, 500, "usual", start, end))

		subRepo, _ := service.NewSubRepo(fakePG)
		uc, _ := NewGetSubsUC(subRepo, &testutil.NopLogger{})
		subs, err := uc.SubsByUserId(ctx, userID)
		pt.Require().NoError(err)
		pt.Assert().Equal(1, len(subs))
		pt.Assert().Equal(500, subs[0].Price)
	})
}

// ---------------------------------------------------------------------------
// 4. DeleteSub
// ---------------------------------------------------------------------------

func TestAllureDeleteSub(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	runner.Run(t, "DeleteSub — Лондонский стиль", func(pt provider.T) {
		pt.Epic("Usecases")
		pt.Feature("DeleteSub")
		pt.Tags("london", "unit", "subscriptions")
		pt.Title("Удаление подписки через мок")

		ctx := context.Background()
		mockSub := mock.NewMockSubscriptionRepository(ctrl)
		mockSub.EXPECT().DeleteSub(ctx, domain.SubID(5)).Return(nil)

		uc, _ := NewDeleteSubUC(mockSub, &testutil.NopLogger{})
		err := uc.DeleteSub(ctx, 5)
		pt.Assert().NoError(err)
	})

	runner.Run(t, "DeleteSub — Классический стиль", func(pt provider.T) {
		pt.Epic("Usecases")
		pt.Feature("DeleteSub")
		pt.Tags("classic", "unit", "subscriptions", "fake-pg")
		pt.Title("Удаление: DELETE ... по CommandTag")

		ctx := context.Background()
		fakePG := testutil.NewFakePgPool()
		fakePG.SetExec("DELETE FROM subscriptions", testutil.CommandTagDelete(1))

		subRepo, _ := service.NewSubRepo(fakePG)
		uc, _ := NewDeleteSubUC(subRepo, &testutil.NopLogger{})
		err := uc.DeleteSub(ctx, 5)
		pt.Assert().NoError(err)
	})
}

// ---------------------------------------------------------------------------
// 5. UpdateSub
// ---------------------------------------------------------------------------

func TestAllureUpdateSub(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	runner.Run(t, "UpdateSub — Лондонский стиль", func(pt provider.T) {
		pt.Epic("Usecases")
		pt.Feature("UpdateSub")
		pt.Tags("london", "unit", "subscriptions")
		pt.Title("Обновление подписки: Sub() + UpdateSub() через моки")

		ctx := context.Background()
		userID := testutil.TestUUID()
		firstOfMonth := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
		end := firstOfMonth.AddDate(0, 1, 0)
		existing := domain.Subscription{
			SubId: 1, UserID: userID, ServiceName: "Netflix", Price: 1000,
			SubType: domain.SubTypeUsual, StartDate: firstOfMonth, EndDate: end, PlanID: 1,
		}
		mockSub := mock.NewMockSubscriptionRepository(ctrl)
		mockSub.EXPECT().Sub(ctx, domain.SubID(1)).Return(existing, nil)
		mockSub.EXPECT().UpdateSub(ctx, gomock.Any()).Return(nil)

		uc, _ := NewUpdateSubUC(mockSub, &testutil.NopLogger{})
		input := SubscriptionDTO{
			UserId: userID, ServiceName: "Netflix", Price: 1100, SubType: "usual",
			StartDate: firstOfMonth, EndDate: end, PlanID: 1,
		}
		err := uc.UpdateSub(ctx, 1, input)
		pt.Assert().NoError(err)
	})

	runner.Run(t, "UpdateSub — Классический стиль", func(pt provider.T) {
		pt.Epic("Usecases")
		pt.Feature("UpdateSub")
		pt.Tags("classic", "unit", "subscriptions", "fake-pg")
		pt.Title("Обновление: две транзакции на одном StubTx (Sub + UPDATE)")

		ctx := context.Background()
		userID := testutil.TestUUID()
		firstOfMonth := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
		end := firstOfMonth.AddDate(0, 1, 0)

		fakePG := testutil.NewFakePgPool()
		fakePG.Tx().OnQueryRow("SELECT s.sub_id", stubSubRow(1, userID, 1, 1000, "usual", firstOfMonth, end))
		fakePG.Tx().OnExec("UPDATE subscriptions", testutil.CommandTagUpdate(1))

		subRepo, _ := service.NewSubRepo(fakePG)
		uc, _ := NewUpdateSubUC(subRepo, &testutil.NopLogger{})
		input := SubscriptionDTO{
			UserId: userID, ServiceName: "Netflix", Price: 1100, SubType: "usual",
			StartDate: firstOfMonth, EndDate: end, PlanID: 1,
		}
		err := uc.UpdateSub(ctx, 1, input)
		pt.Assert().NoError(err)
	})
}

// ---------------------------------------------------------------------------
// 6. CreatePromocode
// ---------------------------------------------------------------------------

func TestAllureCreatePromocode(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	runner.Run(t, "CreatePromocode — Лондонский стиль", func(pt provider.T) {
		pt.Epic("Usecases")
		pt.Feature("CreatePromocode")
		pt.Tags("london", "unit", "promocodes")
		pt.Title("Создание промокода через мок")

		ctx := context.Background()
		mockPromo := mock.NewMockPromocodeRepository(ctrl)
		mockPromo.EXPECT().Create(ctx, gomock.Any()).Return(domain.PromocodeID(123), nil)

		uc, _ := NewCreatePromocodeUC(mockPromo, &testutil.NopLogger{})
		input := CreatePromocodeInput{
			ServiceID: 1, Value: "SUMMER", Discount: 20, MaxUses: 100,
			ExpiresAt: time.Now().AddDate(0, 0, 7), DurationDays: 7,
		}
		id, err := uc.Create(ctx, input)
		pt.Require().NoError(err)
		pt.Assert().Equal(domain.PromocodeID(123), id)
	})

	runner.Run(t, "CreatePromocode — Классический стиль", func(pt provider.T) {
		pt.Epic("Usecases")
		pt.Feature("CreatePromocode")
		pt.Tags("classic", "unit", "promocodes", "fake-pg")
		pt.Title("Создание: транзакция с RETURNING promocode_id")

		ctx := context.Background()
		fakePG := testutil.NewFakePgPool()
		fakePG.Tx().OnQueryRow("RETURNING promocode_id", testutil.NewStubRow(123))

		promoRepo, err := service.NewPromocodeRepo(fakePG)
		pt.Require().NoError(err)
		uc, _ := NewCreatePromocodeUC(promoRepo, &testutil.NopLogger{})
		input := CreatePromocodeInput{
			ServiceID: 1, Value: "SUMMER", Discount: 20, MaxUses: 100,
			ExpiresAt: time.Now().AddDate(0, 0, 7), DurationDays: 7,
		}
		id, err := uc.Create(ctx, input)
		pt.Require().NoError(err)
		pt.Assert().Equal(domain.PromocodeID(123), id)
	})
}

// ---------------------------------------------------------------------------
// 7. GetPromocode
// ---------------------------------------------------------------------------

func TestAllureGetPromocode(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	runner.Run(t, "GetPromocode — Лондонский стиль", func(pt provider.T) {
		pt.Epic("Usecases")
		pt.Feature("GetPromocode")
		pt.Tags("london", "unit", "promocodes")
		pt.Title("Получение промокода по id через мок репозитория")

		ctx := context.Background()
		promo := domain.Promocode{
			PromocodeID: 7, ServiceID: 1, Value: "SUMMER", Discount: 20,
			MaxUses: 100, Status: domain.PromocodeStatusActive,
		}
		mockPromo := mock.NewMockPromocodeRepository(ctrl)
		cache := mock.NewMockPromocodeCache(ctrl)
		mockPromo.EXPECT().GetByID(ctx, domain.PromocodeID(7)).Return(promo, nil)

		uc, _ := NewGetPromocodeUC(mockPromo, cache, &testutil.NopLogger{})
		pc, err := uc.ByID(ctx, domain.PromocodeID(7))
		pt.Require().NoError(err)
		pt.Assert().Equal("SUMMER", pc.Value)
		pt.Assert().Equal(domain.PromocodeStatusActive, pc.Status)
	})

	runner.Run(t, "GetPromocode — Классический стиль", func(pt provider.T) {
		pt.Epic("Usecases")
		pt.Feature("GetPromocode")
		pt.Tags("classic", "unit", "promocodes", "fake-pg")
		pt.Title("Получение: SELECT сканирует 12 полей; мок только для кэша")

		ctx := context.Background()
		fakePG := testutil.NewFakePgPool()
		fakePG.SetQueryRow("WHERE promocode_id = $1", stubPromoRow(7, "SUMMER", 20, 100, "ACTIVE"))

		promoRepo, _ := service.NewPromocodeRepo(fakePG)
		cache := mock.NewMockPromocodeCache(ctrl)
		uc, _ := NewGetPromocodeUC(promoRepo, cache, &testutil.NopLogger{})
		pc, err := uc.ByID(ctx, domain.PromocodeID(7))
		pt.Require().NoError(err)
		pt.Assert().Equal("SUMMER", pc.Value)
		pt.Assert().Equal(20, pc.Discount)
	})
}

// ---------------------------------------------------------------------------
// 8. ApplyPromocode
// ---------------------------------------------------------------------------

func TestAllureApplyPromocode(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	runner.Run(t, "ApplyPromocode — Лондонский стиль", func(pt provider.T) {
		pt.Epic("Usecases")
		pt.Feature("ApplyPromocode")
		pt.Tags("london", "unit", "promocodes")
		pt.Title("Применение промокода: цена 1000 → 800, sub_type=promocode")

		ctx := context.Background()
		now := time.Now()
		sub := domain.Subscription{
			SubId: 1, UserID: testutil.TestUUID(), ServiceName: "Netflix", Price: 1000,
			SubType: domain.SubTypeUsual, StartDate: now.Add(time.Hour),
			EndDate: now.Add(30 * 24 * time.Hour), PlanID: 1,
		}
		promo := domain.Promocode{
			PromocodeID: 5, ServiceID: 1, Value: "SUMMER",
			ExpiresAt: now.Add(7 * 24 * time.Hour), Discount: 20, MaxUses: 100,
			Status: domain.PromocodeStatusActive,
		}
		mockSub := mock.NewMockSubscriptionRepository(ctrl)
		mockPromo := mock.NewMockPromocodeRepository(ctrl)
		mockSub.EXPECT().Sub(ctx, sub.SubId).Return(sub, nil)
		mockPromo.EXPECT().GetByCode(ctx, promo.Value).Return(promo, nil)
		mockSub.EXPECT().UpdateSub(ctx, gomock.Any()).DoAndReturn(
			func(_ context.Context, s domain.Subscription) error {
				pt.Assert().Equal(800, s.Price)
				pt.Assert().Equal("promocode", s.SubType.String())
				return nil
			},
		)
		mockPromo.EXPECT().IncrementUses(ctx, promo.PromocodeID).Return(nil)

		uc, _ := NewApplyPromocodeUC(mockSub, mockPromo, &testutil.NopLogger{})
		out, err := uc.Apply(ctx, ApplyPromocodeInput{SubscriptionID: int(sub.SubId), PromocodeValue: promo.Value})
		pt.Require().NoError(err)
		pt.Assert().Equal(20, out.DiscountApplied)
		pt.Assert().Equal(800, out.NewPrice)
	})

	runner.Run(t, "ApplyPromocode — Классический стиль", func(pt provider.T) {
		pt.Epic("Usecases")
		pt.Feature("ApplyPromocode")
		pt.Tags("classic", "unit", "promocodes", "fake-pg")
		pt.Title("Применение: Sub() → GetByCode() → UPDATE подписки и счётчика")

		ctx := context.Background()
		now := time.Now()
		firstOfNextMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC).AddDate(0, 1, 0)
		subID := domain.SubID(1)
		userID := testutil.TestUUID()
		subRow := stubSubRow(subID, userID, 1, 1000, "usual", firstOfNextMonth, firstOfNextMonth.AddDate(0, 1, 0))
		promoRow := stubPromoRow(5, "SUMMER", 20, 100, "ACTIVE")

		fakePG := testutil.NewFakePgPool()
		fakePG.Tx().OnQueryRow("SELECT s.sub_id", subRow)
		fakePG.SetQueryRow("WHERE promocode = $1", promoRow)
		fakePG.Tx().OnExec("UPDATE subscriptions", testutil.CommandTagUpdate(1))
		fakePG.SetExec("UPDATE promocodes SET", testutil.CommandTagUpdate(1))

		subRepo, _ := service.NewSubRepo(fakePG)
		promoRepo, _ := service.NewPromocodeRepo(fakePG)
		uc, _ := NewApplyPromocodeUC(subRepo, promoRepo, &testutil.NopLogger{})
		out, err := uc.Apply(ctx, ApplyPromocodeInput{SubscriptionID: int(subID), PromocodeValue: "SUMMER"})
		pt.Require().NoError(err)
		pt.Assert().Equal(20, out.DiscountApplied)
		pt.Assert().Equal(800, out.NewPrice)
	})
}

// ---------------------------------------------------------------------------
// 9. PurchaseSubscription
// ---------------------------------------------------------------------------

func TestAllurePurchaseSubscription(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	runner.Run(t, "PurchaseSubscription — Лондонский стиль", func(pt provider.T) {
		pt.Epic("Usecases")
		pt.Feature("PurchaseSubscription")
		pt.Tags("london", "unit", "payments")
		pt.Title("Покупка подписки: баланс, StoreSub, StorePayment, UpdateUser")

		ctx := context.Background()
		userID := testutil.TestUUID()
		mockUser := mock.NewMockUserRepository(ctrl)
		mockSub := mock.NewMockSubscriptionRepository(ctrl)
		mockPayment := mock.NewMockPaymentRepository(ctrl)

		mockUser.EXPECT().GetUser(ctx, userID).Return(domain.User{UserID: userID, Balance: 1000}, nil)
		mockSub.EXPECT().StoreSub(ctx, gomock.Any()).Return(domain.SubID(42), nil)
		mockPayment.EXPECT().StorePayment(ctx, gomock.Any()).Return(nil)
		mockUser.EXPECT().UpdateUser(ctx, gomock.Any()).Return(nil)

		uc, err := NewPurchaseSubscriptionUC(mockUser, mockSub, mockPayment, &testutil.NopLogger{})
		pt.Require().NoError(err)

		input := PurchaseSubscriptionDTO{
			UserID: userID, ServiceName: "Netflix", PlanID: 1, Price: 300, DurationDays: 30,
		}
		id, err := uc.Purchase(ctx, input)
		pt.Require().NoError(err)
		pt.Assert().Equal(42, id)
	})

	runner.Run(t, "PurchaseSubscription — Классический стиль", func(pt provider.T) {
		pt.Epic("Usecases")
		pt.Feature("PurchaseSubscription")
		pt.Tags("classic", "unit", "payments", "fake-pg")
		pt.Title("Покупка: GetUser → StoreSub → StorePayment → UPDATE users")

		ctx := context.Background()
		userID := testutil.TestUUID()
		fakePG := testutil.NewFakePgPool()
		fakePG.SetQueryRow("FROM users", testutil.NewStubRow(
			userID, "alice@example.com", "secret", "Alice", 25, 1000, "REF123", "user",
		))
		fakePG.Tx().OnQueryRow("RETURNING sub_id", testutil.NewStubRow(42))
		fakePG.SetQueryRow("RETURNING paym_id", testutil.NewStubRow(1))
		fakePG.SetExec("UPDATE users", testutil.CommandTagUpdate(1))

		userRepo, _ := service.NewUserRepo(fakePG)
		subRepo, _ := service.NewSubRepo(fakePG)
		paymentRepo, _ := service.NewPaymentRepo(fakePG)
		uc, _ := NewPurchaseSubscriptionUC(userRepo, subRepo, paymentRepo, &testutil.NopLogger{})

		input := PurchaseSubscriptionDTO{
			UserID: userID, ServiceName: "Netflix", PlanID: 1, Price: 300, DurationDays: 30,
		}
		id, err := uc.Purchase(ctx, input)
		pt.Require().NoError(err)
		pt.Assert().Equal(42, id)
	})
}

// ---------------------------------------------------------------------------
// 10. TotalCosts
// ---------------------------------------------------------------------------

func TestAllureTotalCosts(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	runner.Run(t, "TotalCosts — Лондонский стиль", func(pt provider.T) {
		pt.Epic("Usecases")
		pt.Feature("TotalCosts")
		pt.Tags("london", "unit", "subscriptions")
		pt.Title("Сумма затрат за период через мок SubsTotalCosts")

		ctx := context.Background()
		userID := testutil.TestUUID()
		jan1 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
		mar1 := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
		subs := []domain.Subscription{{
			SubId: 1, UserID: userID, ServiceName: "Netflix", Price: 100,
			SubType: domain.SubTypeUsual, StartDate: jan1, EndDate: jan1.AddDate(0, 1, 0), PlanID: 1,
		}}
		mockSub := mock.NewMockSubscriptionRepository(ctrl)
		mockSub.EXPECT().SubsTotalCosts(ctx, gomock.Any()).Return(100, subs, nil)

		uc, _ := NewTotalCostsUC(mockSub, &testutil.NopLogger{})
		input := SubsFilterDTO{StartDate: jan1, EndDate: mar1, UserID: userID, ServiceName: "Netflix", SubType: "usual"}
		sum, result, err := uc.TotalCosts(ctx, input)
		pt.Require().NoError(err)
		pt.Assert().Equal(100, sum)
		pt.Assert().Equal(1, len(result))
	})

	runner.Run(t, "TotalCosts — Классический стиль", func(pt provider.T) {
		pt.Epic("Usecases")
		pt.Feature("TotalCosts")
		pt.Tags("classic", "unit", "subscriptions", "fake-pg")
		pt.Title("Сумма затрат: Query(ids) + Sub() — 100 × 1 месяц")

		ctx := context.Background()
		userID := testutil.TestUUID()
		jan1 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
		feb1 := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
		mar1 := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)

		fakePG := testutil.NewFakePgPool()
		fakePG.SetQuery("SELECT sub_id FROM subscriptions", testutil.NewStubRows([]any{1}))
		fakePG.Tx().OnQueryRow("SELECT s.sub_id", stubSubRow(1, userID, 1, 100, "usual", jan1, feb1))

		subRepo, _ := service.NewSubRepo(fakePG)
		uc, _ := NewTotalCostsUC(subRepo, &testutil.NopLogger{})
		input := SubsFilterDTO{StartDate: jan1, EndDate: mar1, UserID: userID, ServiceName: "Netflix", SubType: "usual"}
		sum, result, err := uc.TotalCosts(ctx, input)
		pt.Require().NoError(err)
		pt.Assert().Equal(1, len(result))
		pt.Assert().Equal(100, sum)
	})
}
