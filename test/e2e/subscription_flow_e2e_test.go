//go:build e2e

// Package e2e содержит сквозной (end-to-end) тест демонстрационного сценария
// сервиса подписок. Тест поднимает реальный HTTP-сервер (delivery.BuildHandler)
// с реальными репозиториями и реальными тестовыми PostgreSQL/Redis, и гоняет
// сценарий целиком через HTTP — так, как это делал бы фронтенд.
//
// Сценарий (MVP): пользователь покупает подписку → видит её в списке →
// запрашивает общую стоимость → администратор создаёт промокод →
// пользователь применяет промокод и получает скидку.
package e2e

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/ozontech/allure-go/pkg/framework/runner"
	"github.com/samantonio28/subscriber-inf/internal/delivery"
	"github.com/samantonio28/subscriber-inf/internal/logger"
	"github.com/samantonio28/subscriber-inf/internal/redis"
	"github.com/samantonio28/subscriber-inf/internal/testutil/testdb"
)

func redisAddr() string {
	if v := os.Getenv("TEST_REDIS_ADDR"); v != "" {
		return v
	}
	return "localhost:6380"
}

// doJSON выполняет HTTP-запрос с JSON-телом и опциональным заголовком X-User-ID
// (аутентификация). Возвращает сырой *http.Response.
func doJSON(t testing.TB, url, method, path, userHeader string, body map[string]any) *http.Response {
	t.Helper()
	var rdr io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal body: %v", err)
		}
		rdr = bytes.NewReader(b)
	}
	req, err := http.NewRequest(method, url+path, rdr)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if userHeader != "" {
		req.Header.Set("X-User-ID", userHeader)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do request %s %s: %v", method, path, err)
	}
	return resp
}

func decode(t testing.TB, resp *http.Response, v any) {
	t.Helper()
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(v); err != nil {
		t.Fatalf("decode response (%s): %v", resp.Status, err)
	}
}

// TestSubscriptionFlowE2E — единый демонстрационный сценарий из нескольких
// шагов (arrange один раз в начале, далее последовательные act/assert).
func TestSubscriptionFlowE2E(t *testing.T) {
	runner.Run(t, "Subscription flow E2E", func(pt provider.T) {
		pt.Feature("SubscriptionFlow")
		pt.Tags("e2e", "http", "mvp")

		// --- Arrange: тестовое окружение и HTTP-сервер ---
		pool := testdb.Connect(pt)
		testdb.Reset(pt, pool)

		redisClient, err := redis.NewRedisClient(redisAddr())
		pt.Require().NoError(err)
		pt.Cleanup(func() { _ = redisClient.Close() })

		lg, err := logger.NewLogrusLogger("logs/e2e_access.log")
		pt.Require().NoError(err)

		handler, err := delivery.BuildHandler(pool, redisClient, lg)
		pt.Require().NoError(err)
		srv := httptest.NewServer(handler)
		defer srv.Close()

		// Базовые сущности: сервис, план, покупатель и администратор.
		svc := testdb.SeedService(pt, pool, "Netflix")
		plan := testdb.SeedPlan(pt, pool, svc, "Netflix Basic", 30, 299)
		customer := testdb.SeedUser(pt, pool, "customer@example.com", 1000, "user")
		admin := testdb.SeedUser(pt, pool, "admin@example.com", 0, "admin")

		const price = 299

		// --- Act 1 / Assert 1: покупка подписки ---
		purchaseResp := doJSON(pt, srv.URL, http.MethodPost, "/subscriptions/purchase", customer.String(), map[string]any{
			"user_id":       customer.String(),
			"service_name":  "Netflix",
			"plan_id":       plan,
			"price":         price,
			"duration_days": 30,
		})
		if purchaseResp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(purchaseResp.Body)
			purchaseResp.Body.Close()
			pt.Fatalf("purchase status = %d, body = %s", purchaseResp.StatusCode, body)
		}
		purchaseResp.Body.Close()

		// --- Act 2 / Assert 2: список подписок пользователя ---
		type subItem struct {
			SubID       int64  `json:"sub_id"`
			ServiceName string `json:"service_name"`
			Price       int64  `json:"price"`
			UserID      string `json:"user_id"`
			SubType     string `json:"sub_type"`
		}
		listResp := doJSON(pt, srv.URL, http.MethodGet, "/subscriptions?uuid="+customer.String(), customer.String(), nil)
		var subs []subItem
		decode(pt, listResp, &subs)
		if len(subs) != 1 {
			pt.Fatalf("expected 1 subscription, got %d", len(subs))
		}
		if subs[0].ServiceName != "Netflix" {
			pt.Errorf("service_name: got %q want Netflix", subs[0].ServiceName)
		}
		if subs[0].Price != price {
			pt.Errorf("price: got %d want %d", subs[0].Price, price)
		}
		subID := subs[0].SubID

		// --- Act 3 / Assert 3: общая стоимость за период ---
		now := time.Now()
		startDate := now.Format("01-2006")
		endDate := now.AddDate(0, 1, 0).Format("01-2006")
		totalResp := doJSON(pt, srv.URL, http.MethodPost, "/total_costs", customer.String(), map[string]any{
			"start_date": startDate,
			"end_date":   endDate,
			"filter": map[string]any{
				"service_name": "Netflix",
				"user_id":      customer.String(),
			},
		})
		var total struct {
			TotalSum int64 `json:"total_sum"`
		}
		decode(pt, totalResp, &total)
		if total.TotalSum != price {
			pt.Errorf("total_sum: got %d want %d", total.TotalSum, price)
		}

		// --- Act 4 / Assert 4: администратор создаёт промокод ---
		promoResp := doJSON(pt, srv.URL, http.MethodPost, "/promocodes", admin.String(), map[string]any{
			"service_id":    svc,
			"value":         "SAVE20",
			"discount":      20,
			"max_uses":      5,
			"duration_days": 30,
		})
		if promoResp.StatusCode != http.StatusCreated {
			body, _ := io.ReadAll(promoResp.Body)
			promoResp.Body.Close()
			pt.Fatalf("create promocode status = %d, body = %s", promoResp.StatusCode, body)
		}
		promoResp.Body.Close()

		// --- Act 5 / Assert 5: применение промокода к подписке ---
		applyResp := doJSON(pt, srv.URL, http.MethodPost, "/subscriptions/"+strconv.FormatInt(subID, 10)+"/apply-promocode", customer.String(), map[string]any{
			"promocode": "SAVE20",
		})
		var applied struct {
			DiscountApplied int `json:"discount_applied"`
			NewPrice        int `json:"new_price"`
		}
		decode(pt, applyResp, &applied)
		if applied.DiscountApplied != 20 {
			pt.Errorf("discount_applied: got %d want 20", applied.DiscountApplied)
		}
		wantNewPrice := price * 80 / 100
		if applied.NewPrice != wantNewPrice {
			pt.Errorf("new_price: got %d want %d", applied.NewPrice, wantNewPrice)
		}
	})
}
