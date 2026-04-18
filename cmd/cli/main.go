package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/samantonio28/subscriber-inf/internal/domain"
	"github.com/samantonio28/subscriber-inf/internal/logger"
	"github.com/samantonio28/subscriber-inf/internal/redis"
	"github.com/samantonio28/subscriber-inf/internal/service"
	"github.com/samantonio28/subscriber-inf/internal/usecase"
	"github.com/samantonio28/subscriber-inf/pkg/config"
)

func main() {
	if len(os.Args) < 2 {
		printHelp()
		os.Exit(1)
	}

	// Инициализация зависимостей
	ctx := context.Background()
	pool, subRepo, promoRepo, planRepo, statsService, redisClient, logr := initDependencies(ctx)
	defer pool.Close()
	defer redisClient.Close()

	command := os.Args[1]
	args := os.Args[2:]

	switch command {
	case "create-sub":
		handleCreateSub(ctx, subRepo, logr, args)
	case "get-sub":
		handleGetSub(ctx, subRepo, logr, args)
	case "list-subs":
		handleListSubs(ctx, subRepo, logr, args)
	case "delete-sub":
		handleDeleteSub(ctx, subRepo, logr, args)
	case "update-sub":
		handleUpdateSub(ctx, subRepo, logr, args)
	case "create-promocode":
		handleCreatePromocode(ctx, promoRepo, logr, args)
	case "get-promocode":
		handleGetPromocode(ctx, promoRepo, logr, args)
	case "get-promocode-by-code":
		handleGetPromocodeByCode(ctx, promoRepo, logr, args)
	case "list-promocodes":
		handleListPromocodes(ctx, promoRepo, logr, args)
	case "update-promocode":
		handleUpdatePromocode(ctx, promoRepo, logr, args)
	case "delete-promocode":
		handleDeletePromocode(ctx, promoRepo, logr, args)
	case "apply-promocode":
		handleApplyPromocode(ctx, promoRepo, subRepo, logr, args)
	case "total-costs":
		handleTotalCosts(ctx, subRepo, logr, args)
	case "create-plan":
		handleCreatePlan(ctx, planRepo, logr, args)
	case "get-plan":
		handleGetPlan(ctx, planRepo, logr, args)
	case "list-plans":
		handleListPlans(ctx, planRepo, logr, args)
	case "update-plan":
		handleUpdatePlan(ctx, planRepo, logr, args)
	case "delete-plan":
		handleDeletePlan(ctx, planRepo, logr, args)
	case "stats-overview":
		handleStatsOverview(ctx, subRepo, promoRepo, statsService, logr, args)
	case "help":
		printHelp()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printHelp()
		os.Exit(1)
	}
}

func initDependencies(ctx context.Context) (*pgxpool.Pool, domain.SubscriptionRepository, domain.PromocodeRepository, domain.SubscriptionPlanRepository, domain.StatsService, *redis.Client, logger.Logger) {
	cfg, err := config.LoadConfig("configs/postgres.yaml")
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	poolConfig, err := cfg.Postgres.ToPgxPoolConfig()
	if err != nil {
		log.Fatal("Failed to create pool config:", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		log.Fatal("Failed to connect to PostgreSQL:", err)
	}

	logr, err := logger.NewLogrusLogger("logs/cli_logger.log")
	if err != nil {
		log.Fatal("Failed to create logger:", err)
	}

	subRepo, err := service.NewSubRepo(pool)
	if err != nil {
		log.Fatal("Failed to create sub repo:", err)
	}

	promoRepo, err := service.NewPromocodeRepo(pool)
	if err != nil {
		log.Fatal("Failed to create promocode repo:", err)
	}

	planRepo, err := service.NewSubscriptionPlanRepo(pool)
	if err != nil {
		log.Fatal("Failed to create subscription plan repo:", err)
	}

	// Redis client
	redisClient, err := redis.NewRedisClient("localhost:6379")
	if err != nil {
		log.Fatal("Failed to connect to Redis:", err)
	}

	statsService, err := service.NewStatsService(pool, redisClient)
	if err != nil {
		log.Fatal("Failed to create stats service:", err)
	}

	return pool, subRepo, promoRepo, planRepo, statsService, redisClient, logr
}

// ===== Subscription commands =====

func handleCreateSub(ctx context.Context, subRepo domain.SubscriptionRepository, logr logger.Logger, args []string) {
	if len(args) < 5 {
		log.Fatal("Usage: create-sub <service_name> <price> <sub_type> <start_date> <end_date> [user_uuid]")
	}
	serviceName := args[0]
	price, err := strconv.Atoi(args[1])
	if err != nil {
		log.Fatal("Invalid price:", err)
	}
	subType := args[2]
	startDate, err := time.Parse(time.RFC3339, args[3])
	if err != nil {
		log.Fatal("Invalid start_date (RFC3339):", err)
	}
	endDate, err := time.Parse(time.RFC3339, args[4])
	if err != nil {
		log.Fatal("Invalid end_date (RFC3339):", err)
	}
	var userID uuid.UUID
	if len(args) >= 6 {
		userID, err = uuid.Parse(args[5])
		if err != nil {
			log.Fatal("Invalid user_uuid:", err)
		}
	} else {
		userID = uuid.Nil
	}

	uc, err := usecase.NewCreateSubUC(subRepo, logr)
	if err != nil {
		log.Fatal("Failed to create usecase:", err)
	}

	input := usecase.SubscriptionDTO{
		SubId:       0,
		UserId:      userID,
		ServiceName: serviceName,
		Price:       price,
		SubType:     subType,
		StartDate:   startDate,
		EndDate:     endDate,
	}
	subID, err := uc.NewSub(ctx, input)
	if err != nil {
		log.Fatal("Failed to create subscription:", err)
	}
	fmt.Printf("Created subscription with ID: %d\n", subID)
}

func handleGetSub(ctx context.Context, subRepo domain.SubscriptionRepository, logr logger.Logger, args []string) {
	if len(args) < 1 {
		log.Fatal("Usage: get-sub <subscription_id>")
	}
	subID, err := strconv.Atoi(args[0])
	if err != nil {
		log.Fatal("Invalid subscription_id:", err)
	}
	uc, err := usecase.NewGetSubUC(subRepo, logr)
	if err != nil {
		log.Fatal("Failed to create usecase:", err)
	}
	sub, err := uc.SubById(ctx, subID)
	if err != nil {
		log.Fatal("Failed to get subscription:", err)
	}
	fmt.Printf("Subscription: %+v\n", sub)
}

func handleListSubs(ctx context.Context, subRepo domain.SubscriptionRepository, logr logger.Logger, args []string) {
	if len(args) < 1 {
		log.Fatal("Usage: list-subs <user_uuid>")
	}
	userID, err := uuid.Parse(args[0])
	if err != nil {
		log.Fatal("Invalid user_uuid:", err)
	}
	uc, err := usecase.NewGetSubsUC(subRepo, logr)
	if err != nil {
		log.Fatal("Failed to create usecase:", err)
	}
	subs, err := uc.SubsByUserId(ctx, userID)
	if err != nil {
		log.Fatal("Failed to list subscriptions:", err)
	}
	for _, sub := range subs {
		fmt.Printf("%d: UserID=%s, ServiceName=%s, Price=%d, Type=%s, Start=%s, End=%s\n",
			sub.SubId, sub.UserId, sub.ServiceName, sub.Price, sub.SubType, sub.StartDate.Format(time.RFC3339), sub.EndDate.Format(time.RFC3339))
	}
}

func handleDeleteSub(ctx context.Context, subRepo domain.SubscriptionRepository, logr logger.Logger, args []string) {
	if len(args) < 1 {
		log.Fatal("Usage: delete-sub <subscription_id>")
	}
	subID, err := strconv.Atoi(args[0])
	if err != nil {
		log.Fatal("Invalid subscription_id:", err)
	}
	uc, err := usecase.NewDeleteSubUC(subRepo, logr)
	if err != nil {
		log.Fatal("Failed to create usecase:", err)
	}
	err = uc.DeleteSub(ctx, subID)
	if err != nil {
		log.Fatal("Failed to delete subscription:", err)
	}
	fmt.Println("Subscription deleted")
}

// ===== Promocode commands =====

func handleCreatePromocode(ctx context.Context, promoRepo domain.PromocodeRepository, logr logger.Logger, args []string) {
	if len(args) < 7 {
		log.Fatal("Usage: create-promocode <value> <service_id> <plan_id> <sub_id> <expires_at> <discount> <max_uses> <duration_days>")
	}
	value := args[0]
	serviceID, _ := strconv.Atoi(args[1])
	planID, _ := strconv.Atoi(args[2])
	subID, _ := strconv.Atoi(args[3])
	expiresAt, err := time.Parse(time.RFC3339, args[4])
	if err != nil {
		log.Fatal("Invalid expires_at (RFC3339):", err)
	}
	discount, _ := strconv.Atoi(args[5])
	maxUses, _ := strconv.Atoi(args[6])
	durationDays, _ := strconv.Atoi(args[7])

	uc, err := usecase.NewCreatePromocodeUC(promoRepo, logr)
	if err != nil {
		log.Fatal("Failed to create usecase:", err)
	}

	input := usecase.CreatePromocodeInput{
		Value:        value,
		ServiceID:    serviceID,
		PlanID:       &planID,
		SubID:        &subID,
		ExpiresAt:    expiresAt,
		Discount:     discount,
		MaxUses:      maxUses,
		DurationDays: durationDays,
	}
	promoID, err := uc.Create(ctx, input)
	if err != nil {
		log.Fatal("Failed to create promocode:", err)
	}
	fmt.Printf("Created promocode with ID: %d\n", promoID)
}

func handleGetPromocode(ctx context.Context, promoRepo domain.PromocodeRepository, logr logger.Logger, args []string) {
	if len(args) < 1 {
		log.Fatal("Usage: get-promocode <promocode_id>")
	}
	promoID, err := strconv.Atoi(args[0])
	if err != nil {
		log.Fatal("Invalid promocode_id:", err)
	}
	uc, err := usecase.NewGetPromocodeUC(promoRepo, logr)
	if err != nil {
		log.Fatal("Failed to create usecase:", err)
	}
	promo, err := uc.ByID(ctx, domain.PromocodeID(promoID))
	if err != nil {
		log.Fatal("Failed to get promocode:", err)
	}
	fmt.Printf("Promocode: %+v\n", promo)
}

func handleListPromocodes(ctx context.Context, promoRepo domain.PromocodeRepository, logr logger.Logger, args []string) {
	// Используем GetFilteredPromocodesUC без фильтров
	uc, err := usecase.NewGetFilteredPromocodesUC(promoRepo, logr)
	if err != nil {
		log.Fatal("Failed to create usecase:", err)
	}
	promos, err := uc.GetFiltered(ctx, usecase.PromocodeFilter{})
	if err != nil {
		log.Fatal("Failed to list promocodes:", err)
	}
	for _, promo := range promos {
		fmt.Printf("%d: %s, ServiceID=%d, Status=%s, Discount=%d%%, Uses=%d/%d\n",
			promo.PromocodeID, promo.Value, promo.ServiceID, promo.Status, promo.Discount, promo.CurUses, promo.MaxUses)
	}
}


func handleTotalCosts(ctx context.Context, subRepo domain.SubscriptionRepository, logr logger.Logger, args []string) {
	if len(args) < 3 {
		log.Fatal("Usage: total-costs <user_id> <start_date> <end_date>")
	}
	userID, err := uuid.Parse(args[0])
	if err != nil {
		log.Fatal("Invalid user_id (UUID):", err)
	}
	startDate, err := time.Parse(time.RFC3339, args[1])
	if err != nil {
		log.Fatal("Invalid start_date (RFC3339):", err)
	}
	endDate, err := time.Parse(time.RFC3339, args[2])
	if err != nil {
		log.Fatal("Invalid end_date (RFC3339):", err)
	}
	uc, err := usecase.NewTotalCostsUC(subRepo, logr)
	if err != nil {
		log.Fatal("Failed to create usecase:", err)
	}
	total, count, err := uc.TotalCosts(ctx, usecase.SubsFilterDTO{
		UserID:    userID,
		StartDate: startDate,
		EndDate:   endDate,
	})
	if err != nil {
		log.Fatal("Failed to calculate total costs:", err)
	}
	fmt.Printf("Total costs: %d (subscriptions count: %d)\n", total, count)
}

// ===== Update subscription =====

func handleUpdateSub(ctx context.Context, subRepo domain.SubscriptionRepository, logr logger.Logger, args []string) {
	if len(args) < 2 {
		log.Fatal("Usage: update-sub <subscription_id> <service_name> <price> <sub_type> <start_date> <end_date> [user_uuid]")
	}
	subID, err := strconv.Atoi(args[0])
	if err != nil {
		log.Fatal("Invalid subscription_id:", err)
	}
	serviceName := args[1]
	price, err := strconv.Atoi(args[2])
	if err != nil {
		log.Fatal("Invalid price:", err)
	}
	subType := args[3]
	startDate, err := time.Parse(time.RFC3339, args[4])
	if err != nil {
		log.Fatal("Invalid start_date (RFC3339):", err)
	}
	endDate, err := time.Parse(time.RFC3339, args[5])
	if err != nil {
		log.Fatal("Invalid end_date (RFC3339):", err)
	}
	var userID uuid.UUID
	if len(args) >= 7 {
		userID, err = uuid.Parse(args[6])
		if err != nil {
			log.Fatal("Invalid user_uuid:", err)
		}
	} else {
		userID = uuid.Nil
	}

	uc, err := usecase.NewUpdateSubUC(subRepo, logr)
	if err != nil {
		log.Fatal("Failed to create usecase:", err)
	}

	input := usecase.SubscriptionDTO{
		SubId:       subID,
		UserId:      userID,
		ServiceName: serviceName,
		Price:       price,
		SubType:     subType,
		StartDate:   startDate,
		EndDate:     endDate,
	}
	err = uc.UpdateSub(ctx, subID, input)
	if err != nil {
		log.Fatal("Failed to update subscription:", err)
	}
	fmt.Printf("Updated subscription with ID: %d\n", subID)
}

// ===== Update promocode =====

func handleUpdatePromocode(ctx context.Context, promoRepo domain.PromocodeRepository, logr logger.Logger, args []string) {
	if len(args) < 11 {
		log.Fatal("Usage: update-promocode <id> <value> <service_id> <plan_id> <sub_id> <expires_at> <discount> <max_uses> <cur_uses> <status> <duration_days>")
	}
	id, _ := strconv.Atoi(args[0])
	value := args[1]
	serviceID, _ := strconv.Atoi(args[2])
	planID, _ := strconv.Atoi(args[3])
	subID, _ := strconv.Atoi(args[4])
	expiresAt, err := time.Parse(time.RFC3339, args[5])
	if err != nil {
		log.Fatal("Invalid expires_at (RFC3339):", err)
	}
	discount, _ := strconv.Atoi(args[6])
	maxUses, _ := strconv.Atoi(args[7])
	curUses, _ := strconv.Atoi(args[8])
	statusStr := args[9]
	var status domain.PromocodeStatus
	switch statusStr {
	case "ACTIVE":
		status = domain.PromocodeStatusActive
	case "USED":
		status = domain.PromocodeStatusUsed
	case "DISABLED":
		status = domain.PromocodeStatusDisabled
	default:
		log.Fatal("Invalid status, must be ACTIVE, USED, or DISABLED")
	}
	durationDays, _ := strconv.Atoi(args[10])

	uc, err := usecase.NewUpdatePromocodeUC(promoRepo, logr)
	if err != nil {
		log.Fatal("Failed to create usecase:", err)
	}

	input := usecase.UpdatePromocodeInput{
		ID:           domain.PromocodeID(id),
		ServiceID:    serviceID,
		Value:        value,
		PlanID:       &planID,
		SubID:        &subID,
		ExpiresAt:    expiresAt,
		Discount:     discount,
		MaxUses:      maxUses,
		CurUses:      curUses,
		Status:       status,
		DurationDays: durationDays,
	}
	err = uc.Update(ctx, input)
	if err != nil {
		log.Fatal("Failed to update promocode:", err)
	}
	fmt.Printf("Updated promocode with ID: %d\n", id)
}

// ===== Delete promocode =====

func handleDeletePromocode(ctx context.Context, promoRepo domain.PromocodeRepository, logr logger.Logger, args []string) {
	if len(args) < 1 {
		log.Fatal("Usage: delete-promocode <promocode_id>")
	}
	promoID, err := strconv.Atoi(args[0])
	if err != nil {
		log.Fatal("Invalid promocode_id:", err)
	}
	uc, err := usecase.NewDeletePromocodeUC(promoRepo, logr)
	if err != nil {
		log.Fatal("Failed to create usecase:", err)
	}
	err = uc.Delete(ctx, domain.PromocodeID(promoID))
	if err != nil {
		log.Fatal("Failed to delete promocode:", err)
	}
	fmt.Println("Promocode deleted")
}

// ===== Get promocode by code =====

func handleGetPromocodeByCode(ctx context.Context, promoRepo domain.PromocodeRepository, logr logger.Logger, args []string) {
	if len(args) < 1 {
		log.Fatal("Usage: get-promocode-by-code <promocode_value>")
	}
	value := args[0]
	uc, err := usecase.NewGetPromocodeUC(promoRepo, logr)
	if err != nil {
		log.Fatal("Failed to create usecase:", err)
	}
	promo, err := uc.ByCode(ctx, value)
	if err != nil {
		log.Fatal("Failed to get promocode:", err)
	}
	fmt.Printf("Promocode: %+v\n", promo)
}

// ===== Apply promocode (full implementation) =====

func handleApplyPromocode(ctx context.Context, promoRepo domain.PromocodeRepository, subRepo domain.SubscriptionRepository, logr logger.Logger, args []string) {
	if len(args) < 2 {
		log.Fatal("Usage: apply-promocode <promocode_value> <subscription_id>")
	}
	value := args[0]
	subID, err := strconv.Atoi(args[1])
	if err != nil {
		log.Fatal("Invalid subscription_id:", err)
	}
	uc, err := usecase.NewApplyPromocodeUC(subRepo, promoRepo, logr)
	if err != nil {
		log.Fatal("Failed to create usecase:", err)
	}
	input := usecase.ApplyPromocodeInput{
		SubscriptionID: subID,
		PromocodeValue: value,
	}
	output, err := uc.Apply(ctx, input)
	if err != nil {
		log.Fatal("Failed to apply promocode:", err)
	}
	fmt.Printf("Discount applied: %d%%, New price: %d, Message: %s\n", output.DiscountApplied, output.NewPrice, output.Message)
}

// ===== Subscription plan commands =====

func handleCreatePlan(ctx context.Context, planRepo domain.SubscriptionPlanRepository, logr logger.Logger, args []string) {
	if len(args) < 4 {
		log.Fatal("Usage: create-plan <service_id> <name> <duration_days> <price>")
	}
	serviceID, _ := strconv.Atoi(args[0])
	name := args[1]
	durationDays, _ := strconv.Atoi(args[2])
	price, _ := strconv.Atoi(args[3])
	uc, err := usecase.NewCreateSubscriptionPlanUC(planRepo, logr)
	if err != nil {
		log.Fatal("Failed to create usecase:", err)
	}
	input := usecase.CreateSubscriptionPlanInput{
		ServiceID:    serviceID,
		Name:         name,
		DurationDays: durationDays,
		Price:        price,
	}
	planID, err := uc.Create(ctx, input)
	if err != nil {
		log.Fatal("Failed to create subscription plan:", err)
	}
	fmt.Printf("Created subscription plan with ID: %d\n", planID)
}

func handleGetPlan(ctx context.Context, planRepo domain.SubscriptionPlanRepository, logr logger.Logger, args []string) {
	if len(args) < 1 {
		log.Fatal("Usage: get-plan <plan_id>")
	}
	planID, err := strconv.Atoi(args[0])
	if err != nil {
		log.Fatal("Invalid plan_id:", err)
	}
	uc, err := usecase.NewGetSubscriptionPlanUC(planRepo, logr)
	if err != nil {
		log.Fatal("Failed to create usecase:", err)
	}
	plan, err := uc.ByID(ctx, domain.PlanID(planID))
	if err != nil {
		log.Fatal("Failed to get subscription plan:", err)
	}
	fmt.Printf("Subscription plan: %+v\n", plan)
}

func handleListPlans(ctx context.Context, planRepo domain.SubscriptionPlanRepository, logr logger.Logger, args []string) {
	// Используем GetFilteredSubscriptionPlansUC без фильтров
	uc, err := usecase.NewGetFilteredSubscriptionPlansUC(planRepo, logr)
	if err != nil {
		log.Fatal("Failed to create usecase:", err)
	}
	plans, err := uc.GetFiltered(ctx, usecase.SubscriptionPlanFilter{})
	if err != nil {
		log.Fatal("Failed to list subscription plans:", err)
	}
	for _, plan := range plans {
		fmt.Printf("%d: ServiceID=%d, Name=%s, Duration=%d days, Price=%d\n",
			plan.PlanID, plan.ServiceID, plan.Name, plan.DurationDays, plan.Price)
	}
}

func handleUpdatePlan(ctx context.Context, planRepo domain.SubscriptionPlanRepository, logr logger.Logger, args []string) {
	if len(args) < 5 {
		log.Fatal("Usage: update-plan <plan_id> <service_id> <name> <duration_days> <price>")
	}
	planID, _ := strconv.Atoi(args[0])
	serviceID, _ := strconv.Atoi(args[1])
	name := args[2]
	durationDays, _ := strconv.Atoi(args[3])
	price, _ := strconv.Atoi(args[4])
	uc, err := usecase.NewUpdateSubscriptionPlanUC(planRepo, logr)
	if err != nil {
		log.Fatal("Failed to create usecase:", err)
	}
	input := usecase.UpdateSubscriptionPlanInput{
		ID:           domain.PlanID(planID),
		ServiceID:    serviceID,
		Name:         name,
		DurationDays: durationDays,
		Price:        price,
	}
	err = uc.Update(ctx, input)
	if err != nil {
		log.Fatal("Failed to update subscription plan:", err)
	}
	fmt.Printf("Updated subscription plan with ID: %d\n", planID)
}

func handleDeletePlan(ctx context.Context, planRepo domain.SubscriptionPlanRepository, logr logger.Logger, args []string) {
	if len(args) < 1 {
		log.Fatal("Usage: delete-plan <plan_id>")
	}
	planID, err := strconv.Atoi(args[0])
	if err != nil {
		log.Fatal("Invalid plan_id:", err)
	}
	uc, err := usecase.NewDeleteSubscriptionPlanUC(planRepo, logr)
	if err != nil {
		log.Fatal("Failed to create usecase:", err)
	}
	err = uc.Delete(ctx, domain.PlanID(planID))
	if err != nil {
		log.Fatal("Failed to delete subscription plan:", err)
	}
	fmt.Println("Subscription plan deleted")
}

// ===== Stats overview =====

func handleStatsOverview(ctx context.Context, subRepo domain.SubscriptionRepository, promoRepo domain.PromocodeRepository, statsService domain.StatsService, logr logger.Logger, args []string) {
	uc, err := usecase.NewStatsOverviewUC(subRepo, promoRepo, statsService, logr)
	if err != nil {
		log.Fatal("Failed to create usecase:", err)
	}
	output, err := uc.GetOverview(ctx)
	if err != nil {
		log.Fatal("Failed to get stats overview:", err)
	}
	fmt.Printf("Total subscriptions: %d\n", output.TotalSubscriptions)
	fmt.Printf("Active subscriptions: %d\n", output.ActiveSubscriptions)
	fmt.Printf("Total promocodes: %d\n", output.TotalPromocodes)
	fmt.Printf("Active promocodes: %d\n", output.ActivePromocodes)
	fmt.Printf("Total revenue: %d\n", output.TotalRevenue)
	fmt.Printf("Avg subscription price: %.2f\n", output.AvgSubscriptionPrice)
	fmt.Printf("Most popular service: %s\n", output.MostPopularService)
	fmt.Println("Service stats:")
	for _, ss := range output.ServiceStats {
		fmt.Printf("  %s: total=%d, revenue=%d, avg=%.2f\n", ss.ServiceName, ss.TotalSubscriptions, ss.TotalRevenue, ss.AvgSubscriptionCost)
	}
}

func printHelp() {
	fmt.Println(`Subscription Service CLI
Commands:
  create-sub <service_name> <price> <sub_type> <start_date> <end_date> [user_uuid]
  get-sub <subscription_id>
  list-subs <user_uuid>
  delete-sub <subscription_id>
  update-sub <subscription_id> <service_name> <price> <sub_type> <start_date> <end_date> [user_uuid]
  create-promocode <value> <service_id> <plan_id> <sub_id> <expires_at> <discount> <max_uses> <duration_days>
  get-promocode <promocode_id>
  get-promocode-by-code <promocode_value>
  list-promocodes
  update-promocode <id> <value> <service_id> <plan_id> <sub_id> <expires_at> <discount> <max_uses> <cur_uses> <status> <duration_days>
  delete-promocode <promocode_id>
  apply-promocode <promocode_value> <subscription_id>
  total-costs <user_id> <start_date> <end_date>
  create-plan <service_id> <name> <duration_days> <price>
  get-plan <plan_id>
  list-plans
  update-plan <plan_id> <service_id> <name> <duration_days> <price>
  delete-plan <plan_id>
  stats-overview
  help

Dates must be in RFC3339 format (e.g., 2025-01-01T00:00:00Z).
PlanID, SubID can be 0 if not applicable.
Status: ACTIVE, USED, DISABLED.`)
}