//go:build e2e

// Package e2e содержит сквозной (end-to-end) тест демонстрационного сценария
// сервиса подписок. Тест работает с РЕАЛЬНО развёрнутым сервисом (отдельный
// процесс/контейнер на порту 8080) и реальными PostgreSQL/Redis: сбрасывает и
// сидирует тестовую БД, после чего гоняет сценарий целиком через HTTP — так,
// как это делал бы фронтенд.
//
// Адрес сервиса задаётся переменной BASE_URL (по умолчанию http://localhost:8080).
package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/ozontech/allure-go/pkg/framework/runner"
	"github.com/samantonio28/subscriber-inf/internal/testutil/testdb"
)

// baseURL возвращает адрес развёрнутого сервиса (переопределяется через BASE_URL).
func baseURL() string {
	if v := os.Getenv("BASE_URL"); v != "" {
		return v
	}
	return "http://localhost:8080"
}

// doJSON выполняет HTTP-запрос с JSON-телом и опциональным заголовком X-User-ID
// (аутентификация). Логирует запрос в curl-подобном виде и возвращает *http.Response.
func doJSON(t testing.TB, url, method, path, userHeader string, body map[string]any) *http.Response {
	t.Helper()
	var bodyBytes []byte
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal body: %v", err)
		}
		bodyBytes = b
	}

	// --- лог запроса (curl-подобный) ---
	fmt.Printf("\n▶ %s %s%s\n", method, url, path)
	if userHeader != "" {
		fmt.Printf("   X-User-ID: %s\n", userHeader)
	}
	if len(bodyBytes) > 0 {
		fmt.Printf("   body: %s\n", bodyBytes)
	}

	var rdr io.Reader
	if bodyBytes != nil {
		rdr = bytes.NewReader(bodyBytes)
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

// decode читает и логирует тело ответа, затем декодирует его в v.
func decode(t testing.TB, resp *http.Response, v any) {
	t.Helper()
	defer resp.Body.Close()
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read response body: %v", err)
	}
	fmt.Printf("◀ %s\n   body: %s\n", resp.Status, bodyBytes)
	if err := json.Unmarshal(bodyBytes, v); err != nil {
		t.Fatalf("decode response (%s): %v", resp.Status, err)
	}
}

// requireStatus проверяет HTTP-статус ответа; при расхождении читает тело и
// валит тест с подробным сообщением.
func requireStatus(s provider.StepCtx, resp *http.Response, want int, desc string) {
	if resp.StatusCode == want {
		resp.Body.Close()
		fmt.Printf("◀ %s\n", resp.Status)
		return
	}
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	fmt.Printf("◀ %s (unexpected)\n   body: %s\n", resp.Status, body)
	s.Require().Equal(want, resp.StatusCode, "%s: got status %d, body=%s", desc, resp.StatusCode, body)
}

// TestSubscriptionFlowE2E — единый демонстрационный сценарий из нескольких
// шагов (arrange один раз в начале, далее последовательные act/assert).
func TestSubscriptionFlowE2E(t *testing.T) {
	runner.Run(t, "Subscription flow E2E", func(pt provider.T) {
		pt.Feature("SubscriptionFlow")
		pt.Description("Сквозной HTTP-сценарий MVP против реально развёрнутого сервиса: пользователь покупает подписку, " +
			"видит её в списке, запрашивает общую стоимость; администратор создаёт промокод; пользователь применяет его и получает скидку.")

		// --- Arrange: одной строкой подключаемся к тестовой БД и сбрасываем её
		// (сервис подключён к этой же БД и увидит сидированные ниже данные). ---
		pool := testdb.Setup(pt)

		svc := testdb.SeedService(pt, pool, "Netflix")
		plan := testdb.SeedPlan(pt, pool, svc, "Netflix Basic", 30, 299)
		customer := testdb.SeedUser(pt, pool, "customer@example.com", 1000, "user")
		admin := testdb.SeedUser(pt, pool, "admin@example.com", 0, "admin")

		const price = 299
		var subID int64
		url := baseURL()

		pt.WithNewStep("1. Покупка подписки (customer)", func(s provider.StepCtx) {
			resp := doJSON(pt, url, http.MethodPost, "/subscriptions/purchase", customer.String(), map[string]any{
				"user_id":       customer.String(),
				"service_name":  "Netflix",
				"plan_id":       plan,
				"price":         price,
				"duration_days": 30,
			})
			requireStatus(s, resp, http.StatusOK, "покупка подписки должна вернуть 200 OK")
		})

		pt.WithNewStep("2. Подписка видна в списке пользователя", func(s provider.StepCtx) {
			type subItem struct {
				SubID       int64  `json:"sub_id"`
				ServiceName string `json:"service_name"`
				Price       int64  `json:"price"`
				UserID      string `json:"user_id"`
				SubType     string `json:"sub_type"`
			}
			resp := doJSON(pt, url, http.MethodGet, "/subscriptions?uuid="+customer.String(), customer.String(), nil)
			var subs []subItem
			decode(pt, resp, &subs)
			s.Require().Equal(1, len(subs), "у покупателя должна быть ровно одна подписка")
			s.Assert().Equal("Netflix", subs[0].ServiceName, "имя сервиса должно быть Netflix")
			s.Assert().Equal(int64(price), subs[0].Price, "цена подписки должна быть 299")
			subID = subs[0].SubID
		})

		pt.WithNewStep("3. Общая стоимость за период", func(s provider.StepCtx) {
			now := time.Now()
			startDate := now.Format("01-2006")
			endDate := now.AddDate(0, 1, 0).Format("01-2006")
			resp := doJSON(pt, url, http.MethodPost, "/total_costs", customer.String(), map[string]any{
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
			decode(pt, resp, &total)
			s.Assert().Equal(int64(price), total.TotalSum, "сумма за месяц должна быть равна цене подписки")
		})

		pt.WithNewStep("4. Администратор создаёт промокод SAVE20", func(s provider.StepCtx) {
			resp := doJSON(pt, url, http.MethodPost, "/promocodes", admin.String(), map[string]any{
				"service_id":    svc,
				"value":         "SAVE20",
				"discount":      20,
				"max_uses":      5,
				"duration_days": 30,
			})
			requireStatus(s, resp, http.StatusCreated, "создание промокода должно вернуть 201 Created")
		})

		pt.WithNewStep("5. Применение промокода со скидкой 20%", func(s provider.StepCtx) {
			resp := doJSON(pt, url, http.MethodPost, "/subscriptions/"+strconv.FormatInt(subID, 10)+"/apply-promocode", customer.String(), map[string]any{
				"promocode": "SAVE20",
			})
			var applied struct {
				DiscountApplied int `json:"discount_applied"`
				NewPrice        int `json:"new_price"`
			}
			decode(pt, resp, &applied)
			s.Assert().Equal(20, applied.DiscountApplied, "скидка должна быть 20%")
			s.Assert().Equal(price*80/100, applied.NewPrice, "новая цена должна быть 299 * 80% = 239")
		})
	})
}
