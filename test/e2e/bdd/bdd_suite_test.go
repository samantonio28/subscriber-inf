//go:build e2e

// Package bdd содержит BDD (Gherkin/godog) E2E-тесты аутентификации ЛР №5.
//
// Сценарии (*.feature) гоняются против РЕАЛЬНО развёрнутого сервиса + PostgreSQL
// + Redis через HTTP — как это делал бы фронтенд. BDD-методология: сначала
// фичи, затем шаги.
//
// Требования ЛР №5, покрываемые сценариями:
//   - двухфакторная аутентификация (второй фактор — email-код);
//   - ограниченное количество попыток ввода (3 неверных пароля → блокировка 423);
//   - восстановление при блокировке (код восстановления → разблокировка + новый пароль);
//   - плановая смена пароля;
//   - технический пользователь: создаётся регистрацией перед сценарием
//     (Background) и утилизируется (удаляется из БД и Redis) после сценария;
//   - пароль НЕ хранится в коде: берётся из переменной окружения
//     E2E_TECH_USER_PASSWORD (в CI — GitHub Actions secret).
//
// Запуск: go test -p 1 -tags e2e ./test/e2e/bdd/...  (см. make e2e-bdd)
package bdd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/cucumber/godog"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/samantonio28/subscriber-inf/internal/redis"
	"github.com/samantonio28/subscriber-inf/internal/service"
	"github.com/samantonio28/subscriber-inf/internal/testutil/testdb"
)

// baseURL — адрес развёрнутого сервиса (переопределяется через BASE_URL).
func baseURL() string {
	if v := os.Getenv("BASE_URL"); v != "" {
		return v
	}
	return "http://localhost:8080"
}

// techPassword — пароль технического пользователя. НЕ должен лежать в коде:
// локально задаётся в .env.e2e (gitignored), в CI — GitHub secret
// E2E_TECH_USER_PASSWORD.
func techPassword() (string, error) {
	if v := os.Getenv("E2E_TECH_USER_PASSWORD"); v != "" {
		return v, nil
	}
	return "", fmt.Errorf("переменная окружения E2E_TECH_USER_PASSWORD не задана: " +
		"локально — скопируйте .env.e2e.example в .env.e2e, в CI — GitHub secret")
}

// redisAddr — адрес Redis тестового стенда (для чтения «почтового ящика»).
func redisAddr() string {
	if v := os.Getenv("TEST_REDIS_ADDR"); v != "" {
		return v
	}
	return "localhost:6380"
}

// bddState — состояние текущего сценария.
type bddState struct {
	email           string
	password        string // начальный пароль (из E2E_TECH_USER_PASSWORD)
	newPassword     string // пароль, установленный сменой/восстановлением
	userID          string
	authCode        string // последний код 2FA из письма
	recoveryCode    string // последний код восстановления из письма
	lastStatus      int
	lastBody        string
	statuses        []int // статусы серии запросов (лимит попыток)
	pool            *pgxpool.Pool
	rdb             *redis.Client
	ctx             context.Context
}

var s = &bddState{}

// --- HTTP helpers --------------------------------------------------------------

// doAuthJSON выполняет HTTP-запрос с JSON-телом и опциональным X-User-ID,
// сохраняет статус и тело в состояние сценария.
func (st *bddState) do(method, path, userHeader string, body map[string]any) {
	var rdr io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			panic(fmt.Sprintf("marshal body: %v", err))
		}
		rdr = bytes.NewReader(b)
		fmt.Printf("▶ %s %s body: %s\n", method, path, b)
	} else {
		fmt.Printf("▶ %s %s\n", method, path)
	}
	req, err := http.NewRequest(method, baseURL()+path, rdr)
	if err != nil {
		panic(fmt.Sprintf("new request: %v", err))
	}
	req.Header.Set("Content-Type", "application/json")
	if userHeader != "" {
		req.Header.Set("X-User-ID", userHeader)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(fmt.Sprintf("do request %s %s: %v", method, path, err))
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	fmt.Printf("◀ %s body: %s\n", resp.Status, strings.TrimSpace(string(raw)))
	st.lastStatus = resp.StatusCode
	st.lastBody = string(raw)
}

func (st *bddState) login(email, secret string) {
	st.do(http.MethodPost, "/auth/login", "", map[string]any{
		"email":        email,
		"login_secret": secret,
	})
}

// outboxLetter читает «почтовый ящик» пользователя (Redis outbox) —
// роль получателя письма в сценарии.
func (st *bddState) outboxLetter() (string, error) {
	env, err := service.ReadOutbox(st.ctx, st.rdb, st.email)
	if err != nil {
		return "", err
	}
	return env.Body, nil
}

// latestCode извлекает 6-значный код из последнего письма.
func (st *bddState) latestCode() (string, error) {
	body, err := st.outboxLetter()
	if err != nil {
		return "", err
	}
	re := regexp.MustCompile(`\b\d{6}\b`)
	code := re.FindString(body)
	if code == "" {
		return "", fmt.Errorf("в письме нет 6-значного кода: %s", body)
	}
	return code, nil
}

// outboxEmpty проверяет, что письма ещё не было.
func (st *bddState) outboxEmpty() bool {
	_, err := st.outboxLetter()
	return err != nil
}

// registerTechUser регистрирует технического пользователя с уникальным email.
func (st *bddState) registerTechUser() error {
	secret, err := techPassword()
	if err != nil {
		return err
	}
	st.password = secret
	st.email = fmt.Sprintf("bdd-tech-%s@example.com", uuid.New().String()[:8])
	st.do(http.MethodPost, "/auth/register", "", map[string]any{
		"email":        st.email,
		"user_name":    "BDD Tech User",
		"age":          30,
		"balance":      1000,
		"login_secret": st.password,
	})
	if st.lastStatus != http.StatusCreated {
		return fmt.Errorf("регистрация техпользователя: ожидается 201, получено %d", st.lastStatus)
	}
	var resp struct {
		UserID string `json:"user_id"`
	}
	if err := json.Unmarshal([]byte(st.lastBody), &resp); err != nil || resp.UserID == "" {
		return fmt.Errorf("нет user_id в ответе регистрации: %s", st.lastBody)
	}
	st.userID = resp.UserID
	st.newPassword = ""
	st.authCode = ""
	st.recoveryCode = ""
	st.statuses = nil
	return nil
}

// utilizeTechUser утилизирует техпользователя: удаляет его из БД
// и очищает почтовый ящик / состояние аутентификации в Redis.
func (st *bddState) utilizeTechUser() {
	if st.pool != nil && st.email != "" {
		_, _ = st.pool.Exec(st.ctx, "DELETE FROM users WHERE email = $1", st.email)
	}
	if st.rdb != nil {
		if userID := st.userID; userID != "" {
			_ = st.rdb.Delete(st.ctx, "auth:failures:"+userID)
			_ = st.rdb.Delete(st.ctx, "auth:lock:"+userID)
			_ = st.rdb.Delete(st.ctx, "auth:code:"+userID)
		}
		_ = st.rdb.Delete(st.ctx, "auth:recovery:"+st.email)
		_ = st.rdb.Delete(st.ctx, "mail:outbox:"+st.email)
	}
	fmt.Printf("🧹 техпользователь %s утилизирован (удалён из БД и Redis)\n", st.email)
}

// fullLogin выполняет полный вход: пароль → код из письма → подтверждение.
func (st *bddState) fullLogin(email, secret string) error {
	st.login(email, secret)
	if st.lastStatus != http.StatusOK {
		return fmt.Errorf("первый фактор: ожидается 200, получено %d", st.lastStatus)
	}
	code, err := st.latestCode()
	if err != nil {
		return err
	}
	st.do(http.MethodPost, "/auth/verify", "", map[string]any{
		"user_id": st.userID,
		"code":    code,
	})
	if st.lastStatus != http.StatusOK {
		return fmt.Errorf("второй фактор: ожидается 200, получено %d", st.lastStatus)
	}
	return nil
}

// --- Шаги ----------------------------------------------------------------------

func InitializeScenario(ctx *godog.ScenarioContext) {
	ctx.Before(func(sctx context.Context, sc *godog.Scenario) (context.Context, error) {
		s.ctx = sctx
		return sctx, nil
	})

	ctx.After(func(sctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
		s.utilizeTechUser()
		return sctx, err
	})

	// --- Background / подготовка -------------------------------------------
	ctx.Given(`^зарегистрирован технический пользователь с паролем из переменной окружения$`, func() error {
		return s.registerTechUser()
	})

	ctx.Given(`^аккаунт заблокирован после 3 неверных попыток входа$`, func() error {
		for i := 0; i < 3; i++ {
			s.login(s.email, s.password+"-WRONG")
		}
		return nil
	})

	ctx.Given(`^пользователь полностью прошёл аутентификацию с двумя факторами$`, func() error {
		return s.fullLogin(s.email, s.password)
	})

	// --- Вход ---------------------------------------------------------------
	ctx.When(`^пользователь входит с верным паролем$`, func() error {
		s.login(s.email, s.password)
		return nil
	})

	ctx.When(`^пользователь входит с неверным паролем$`, func() error {
		s.login(s.email, s.password+"-WRONG")
		return nil
	})

	ctx.When(`^пользователь (\d+) раза? подряд входит с неверным паролем$`, func(n int) error {
		s.statuses = nil
		for i := 0; i < n; i++ {
			s.login(s.email, s.password+"-WRONG")
			s.statuses = append(s.statuses, s.lastStatus)
		}
		return nil
	})

	ctx.When(`^пользователь входит с паролем (\S+)$`, func(kind string) error {
		switch kind {
		case "старым":
			s.login(s.email, s.password)
		case "новым":
			s.login(s.email, s.newPassword)
		default:
			return fmt.Errorf("неизвестный вид пароля: %s", kind)
		}
		return nil
	})

	// --- 2FA ---------------------------------------------------------------
	ctx.Then(`^сервер отвечает (\d+) и сообщает, что код отправлен$`, func(status int) error {
		if s.lastStatus != status {
			return fmt.Errorf("ожидается %d, получено %d", status, s.lastStatus)
		}
		if !strings.Contains(s.lastBody, `"code_sent"`) {
			return fmt.Errorf("ожидается статус code_sent, тело: %s", s.lastBody)
		}
		return nil
	})

	// ctx.Step: шаг используется и после Then, и после When — godog резолвит
	// And по предыдущему ключевому слову.
	ctx.Step(`^пользователю на email отправлен одноразовый код подтверждения$`, func() error {
		code, err := s.latestCode()
		if err != nil {
			return err
		}
		s.authCode = code
		return nil
	})

	ctx.Then(`^код подтверждения на email не отправлялся$`, func() error {
		if !s.outboxEmpty() {
			return fmt.Errorf("в почтовом ящике %s есть письмо, а его быть не должно", s.email)
		}
		return nil
	})

	ctx.When(`^пользователь подтверждает вход этим кодом$`, func() error {
		s.do(http.MethodPost, "/auth/verify", "", map[string]any{
			"user_id": s.userID,
			"code":    s.authCode,
		})
		return nil
	})

	ctx.When(`^пользователь повторно подтверждает вход тем же кодом$`, func() error {
		s.do(http.MethodPost, "/auth/verify", "", map[string]any{
			"user_id": s.userID,
			"code":    s.authCode,
		})
		return nil
	})

	ctx.When(`^пользователь подтверждает вход неверным кодом$`, func() error {
		wrong := "000000"
		if s.authCode == wrong {
			wrong = "111111"
		}
		s.do(http.MethodPost, "/auth/verify", "", map[string]any{
			"user_id": s.userID,
			"code":    wrong,
		})
		return nil
	})

	ctx.Then(`^сервер отвечает 200 и вход подтверждён$`, func() error {
		if s.lastStatus != http.StatusOK {
			return fmt.Errorf("ожидается 200, получено %d", s.lastStatus)
		}
		return nil
	})

	// --- Общие проверки статуса ----------------------------------------------
	ctx.Then(`^сервер отвечает (\d+) invalid credentials$`, func(status int) error {
		if s.lastStatus != status {
			return fmt.Errorf("ожидается %d, получено %d", status, s.lastStatus)
		}
		return nil
	})

	ctx.Then(`^сервер отвечает (\d+) invalid code$`, func(status int) error {
		if s.lastStatus != status {
			return fmt.Errorf("ожидается %d, получено %d", status, s.lastStatus)
		}
		return nil
	})

	ctx.Then(`^сервер отвечает (\d+) locked$`, func(status int) error {
		if s.lastStatus != status {
			return fmt.Errorf("ожидается %d, получено %d", status, s.lastStatus)
		}
		return nil
	})

	ctx.Then(`^сервер отвечает (\d+) authentication required$`, func(status int) error {
		if s.lastStatus != status {
			return fmt.Errorf("ожидается %d, получено %d", status, s.lastStatus)
		}
		return nil
	})

	ctx.Then(`^каждый ответ сервера — (\d+) invalid credentials$`, func(status int) error {
		for i, st := range s.statuses {
			if st != status {
				return fmt.Errorf("попытка %d: ожидается %d, получено %d", i+1, status, st)
			}
		}
		return nil
	})

	// --- Ограничение попыток / блокировка ------------------------------------
	ctx.Then(`^третья неверная попытка отвечает (\d+) locked$`, func(status int) error {
		s.login(s.email, s.password+"-WRONG")
		if s.lastStatus != status {
			return fmt.Errorf("ожидается %d, получено %d", status, s.lastStatus)
		}
		return nil
	})

	ctx.Then(`^попытка входа с верным паролем отвечает (\d+) locked$`, func(status int) error {
		s.login(s.email, s.password)
		if s.lastStatus != status {
			return fmt.Errorf("ожидается %d, получено %d", status, s.lastStatus)
		}
		return nil
	})

	ctx.Then(`^аккаунт по-прежнему заблокирован$`, func() error {
		s.login(s.email, s.password)
		if s.lastStatus != http.StatusLocked {
			return fmt.Errorf("ожидается 423 locked, получено %d", s.lastStatus)
		}
		return nil
	})

	// --- Восстановление --------------------------------------------------------
	ctx.When(`^пользователь запрашивает код восстановления$`, func() error {
		s.do(http.MethodPost, "/auth/recover", "", map[string]any{"email": s.email})
		if s.lastStatus != http.StatusOK {
			return fmt.Errorf("запрос восстановления: ожидается 200, получено %d", s.lastStatus)
		}
		code, err := s.latestCode()
		if err != nil {
			return err
		}
		s.recoveryCode = code
		return nil
	})

	ctx.Then(`^сервер отвечает 200 и код восстановления отправлен на email$`, func() error {
		if s.lastStatus != http.StatusOK {
			return fmt.Errorf("ожидается 200, получено %d", s.lastStatus)
		}
		return nil
	})

	ctx.When(`^пользователь подтверждает восстановление неверным кодом$`, func() error {
		wrong := "000000"
		if s.recoveryCode == wrong {
			wrong = "111111"
		}
		s.do(http.MethodPost, "/auth/recover/confirm", "", map[string]any{
			"email":           s.email,
			"code":            wrong,
			"new_login_secret": s.password + "-NEW",
		})
		return nil
	})

	ctx.When(`^пользователь подтверждает восстановление кодом из письма и новым паролем$`, func() error {
		s.newPassword = s.password + "-RECOVERED"
		s.do(http.MethodPost, "/auth/recover/confirm", "", map[string]any{
			"email":            s.email,
			"code":             s.recoveryCode,
			"new_login_secret": s.newPassword,
		})
		if s.lastStatus != http.StatusOK {
			return fmt.Errorf("подтверждение восстановления: ожидается 200, получено %d", s.lastStatus)
		}
		return nil
	})

	ctx.Then(`^сервер отвечает 200 и аккаунт разблокирован$`, func() error {
		if s.lastStatus != http.StatusOK {
			return fmt.Errorf("ожидается 200, получено %d", s.lastStatus)
		}
		return nil
	})

	ctx.Then(`^вход со старым паролем отвечает (\d+)$`, func(status int) error {
		s.login(s.email, s.password)
		if s.lastStatus != status {
			return fmt.Errorf("ожидается %d, получено %d", status, s.lastStatus)
		}
		return nil
	})

	ctx.Then(`^вход с новым паролем и кодом из письма подтверждён$`, func() error {
		return s.fullLogin(s.email, s.newPassword)
	})

	// --- Плановая смена пароля ---------------------------------------------------
	ctx.When(`^пользователь меняет пароль с текущим на новый$`, func() error {
		s.newPassword = s.password + "-CHANGED"
		s.do(http.MethodPut, "/auth/password", s.userID, map[string]any{
			"user_id":             s.userID,
			"current_login_secret": s.password,
			"new_login_secret":    s.newPassword,
		})
		return nil
	})

	ctx.When(`^пользователь меняет пароль, указав неверный текущий$`, func() error {
		s.do(http.MethodPut, "/auth/password", s.userID, map[string]any{
			"user_id":             s.userID,
			"current_login_secret": s.password + "-WRONG",
			"new_login_secret":    s.password + "-CHANGED",
		})
		return nil
	})

	ctx.When(`^неаутентифицированный пользователь меняет пароль$`, func() error {
		// Нет заголовка X-User-ID — аутентификация не пройдена.
		s.do(http.MethodPut, "/auth/password", "", map[string]any{
			"user_id":             s.userID,
			"current_login_secret": s.password,
			"new_login_secret":    s.password + "-CHANGED",
		})
		return nil
	})

	ctx.Then(`^сервер отвечает 200 и пароль изменён$`, func() error {
		if s.lastStatus != http.StatusOK {
			return fmt.Errorf("ожидается 200, получено %d", s.lastStatus)
		}
		return nil
	})
}

// TestMain запускает godog-сьют: фичи из ./features, отчёт — Cucumber JSON
// (конвертируется в Allure в CI; см. tools/cucumber2allure).
func TestMain(m *testing.M) {
	cucumberPath := os.Getenv("CUCUMBER_JSON_PATH")
	if cucumberPath == "" {
		cucumberPath = "cucumber.json"
	}
	if dir := filepath.Dir(cucumberPath); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			fmt.Printf("cannot mkdir %s: %v\n", dir, err)
			os.Exit(1)
		}
	}
	f, err := os.Create(cucumberPath)
	if err != nil {
		fmt.Printf("cannot create %s: %v\n", cucumberPath, err)
		os.Exit(1)
	}
	defer f.Close()

	opts := godog.Options{
		Paths:     []string{"features"},
		Format:    "cucumber",
		Output:    f,
		Strict:    true,
	}

	connectStands()
	defer func() {
		if s.pool != nil {
			s.pool.Close()
		}
		if s.rdb != nil {
			_ = s.rdb.Close()
		}
	}()

	status := godog.TestSuite{
		Name:                "auth-bdd-e2e",
		ScenarioInitializer: InitializeScenario,
		Options:             &opts,
	}.Run()

	if st := m.Run(); st > status {
		status = st
	}
	os.Exit(status)
}

// connectStands открывает соединения к тестовому стенду: БД (утилизация
// техпользователя) и Redis (чтение «почтового ящика»).
func connectStands() {
	var err error
	s.pool, err = pgxpool.New(context.Background(), testdb.DSN())
	if err != nil {
		fmt.Printf("⚠ не удалось подключиться к тестовой БД (%v): утилизация техпользователя будет неполной\n", err)
	}
	s.rdb, err = redis.NewRedisClient(redisAddr())
	if err != nil {
		fmt.Printf("⚠ не удалось подключиться к Redis (%v): чтение почтового ящика не сработает\n", err)
	}
}
