package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
	apiv2 "github.com/samantonio28/subscriber-inf/pkg/clients/apiv2"
)

func main() {
	if len(os.Args) < 2 {
		printHelp()
		os.Exit(1)
	}

	ctx := context.Background()

	baseURL := os.Getenv("API_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}
	client, err := apiv2.NewClientWithResponses(baseURL)
	if err != nil {
		log.Fatalf("failed to create API client: %v", err)
	}

	command := os.Args[1]
	args := os.Args[2:]

	switch command {
	// users
	case "create-user":
		handleCreateUser(ctx, client, args)
	case "get-user":
		handleGetUser(ctx, client, args)
	// services
	case "list-services":
		handleListServices(ctx, client)
	case "create-service":
		handleCreateService(ctx, client, args)
	case "get-service":
		handleGetService(ctx, client, args)
	case "delete-service":
		handleDeleteService(ctx, client, args)
	// plans
	case "list-plans":
		handleListPlans(ctx, client)
	case "create-plan":
		handleCreatePlan(ctx, client, args)
	case "get-plan":
		handleGetPlan(ctx, client, args)
	case "delete-plan":
		handleDeletePlan(ctx, client, args)
	// subscriptions
	case "list-subs":
		handleListSubs(ctx, client, args)
	case "create-sub":
		handleCreateSub(ctx, client, args)
	case "get-sub":
		handleGetSub(ctx, client, args)
	case "delete-sub":
		handleDeleteSub(ctx, client, args)
	case "patch-sub":
		handlePatchSub(ctx, client, args)
	case "apply-promocode":
		handleApplyPromocode(ctx, client, args)
	// promocodes
	case "list-promocodes":
		handleListPromocodes(ctx, client)
	case "create-promocode":
		handleCreatePromocode(ctx, client, args)
	case "get-promocode":
		handleGetPromocode(ctx, client, args)
	case "delete-promocode":
		handleDeletePromocode(ctx, client, args)
	// analytics
	case "total-costs":
		handleTotalCosts(ctx, client, args)
	case "stats-overview":
		handleStatsOverview(ctx, client)
	case "help":
		printHelp()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printHelp()
		os.Exit(1)
	}
}

// ---------- users ----------

func handleCreateUser(ctx context.Context, c *apiv2.ClientWithResponses, args []string) {
	if len(args) < 5 {
		log.Fatal("Usage: create-user <email> <password> <user_name> <age> <balance>")
	}
	age, _ := strconv.Atoi(args[3])
	balance, _ := strconv.Atoi(args[4])
	pass := args[1]

	resp, err := c.CreateUserWithResponse(ctx, apiv2.UserCreate{
		Email:    openapi_types.Email(args[0]),
		Password: &pass,
		UserName: args[2],
		Age:      age,
		Balance:  balance,
	})
	if err != nil {
		log.Fatal(err)
	}
	if resp.JSON201 != nil {
		fmt.Printf("created user: %+v\n", *resp.JSON201)
		return
	}
	printErr(resp.HTTPResponse)
}

func handleGetUser(ctx context.Context, c *apiv2.ClientWithResponses, args []string) {
	if len(args) < 1 {
		log.Fatal("Usage: get-user <user_id>")
	}
	userID, err := uuid.Parse(args[0])
	if err != nil {
		log.Fatal("invalid user_id")
	}
	resp, err := c.GetUserByIdWithResponse(ctx, openapi_types.UUID(userID))
	if err != nil {
		log.Fatal(err)
	}
	if resp.JSON200 != nil {
		fmt.Printf("user: %+v\n", *resp.JSON200)
		return
	}
	printErr(resp.HTTPResponse)
}

// ---------- services ----------

func handleListServices(ctx context.Context, c *apiv2.ClientWithResponses) {
	resp, err := c.ListServicesWithResponse(ctx)
	if err != nil {
		log.Fatal(err)
	}
	if resp.JSON200 != nil {
		for _, s := range *resp.JSON200 {
			fmt.Printf("%+v\n", s)
		}
		return
	}
	printErr(resp.HTTPResponse)
}

func handleCreateService(ctx context.Context, c *apiv2.ClientWithResponses, args []string) {
	if len(args) < 1 {
		log.Fatal("Usage: create-service <service_name>")
	}
	resp, err := c.CreateServiceWithResponse(ctx, apiv2.ServiceCreate{ServiceName: args[0]})
	if err != nil {
		log.Fatal(err)
	}
	if resp.JSON201 != nil {
		fmt.Printf("created service: %+v\n", *resp.JSON201)
		return
	}
	printErr(resp.HTTPResponse)
}

func handleGetService(ctx context.Context, c *apiv2.ClientWithResponses, args []string) {
	if len(args) < 1 {
		log.Fatal("Usage: get-service <service_id>")
	}
	id, _ := strconv.Atoi(args[0])
	resp, err := c.GetServiceByIdWithResponse(ctx, id)
	if err != nil {
		log.Fatal(err)
	}
	if resp.JSON200 != nil {
		fmt.Printf("service: %+v\n", *resp.JSON200)
		return
	}
	printErr(resp.HTTPResponse)
}

func handleDeleteService(ctx context.Context, c *apiv2.ClientWithResponses, args []string) {
	if len(args) < 1 {
		log.Fatal("Usage: delete-service <service_id>")
	}
	id, _ := strconv.Atoi(args[0])
	resp, err := c.DeleteServiceWithResponse(ctx, id)
	if err != nil {
		log.Fatal(err)
	}
	if resp.HTTPResponse != nil && resp.HTTPResponse.StatusCode == 204 {
		fmt.Println("service deleted")
		return
	}
	printErr(resp.HTTPResponse)
}

// ---------- plans ----------

func handleListPlans(ctx context.Context, c *apiv2.ClientWithResponses) {
	resp, err := c.ListSubscriptionPlansWithResponse(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}
	if resp.JSON200 != nil {
		for _, p := range *resp.JSON200 {
			fmt.Printf("%+v\n", p)
		}
		return
	}
	printErr(resp.HTTPResponse)
}

func handleCreatePlan(ctx context.Context, c *apiv2.ClientWithResponses, args []string) {
	if len(args) < 4 {
		log.Fatal("Usage: create-plan <service_id> <name> <duration_days> <price>")
	}
	serviceID, _ := strconv.Atoi(args[0])
	dur, _ := strconv.Atoi(args[2])
	price, _ := strconv.Atoi(args[3])
	resp, err := c.CreateSubscriptionPlanWithResponse(ctx, apiv2.SubscriptionPlanCreate{
		ServiceId:    serviceID,
		Name:         args[1],
		DurationDays: dur,
		Price:        price,
	})
	if err != nil {
		log.Fatal(err)
	}
	if resp.JSON201 != nil {
		fmt.Printf("created plan: %+v\n", *resp.JSON201)
		return
	}
	printErr(resp.HTTPResponse)
}

func handleGetPlan(ctx context.Context, c *apiv2.ClientWithResponses, args []string) {
	if len(args) < 1 {
		log.Fatal("Usage: get-plan <plan_id>")
	}
	id, _ := strconv.Atoi(args[0])
	resp, err := c.GetSubscriptionPlanByIdWithResponse(ctx, id)
	if err != nil {
		log.Fatal(err)
	}
	if resp.JSON200 != nil {
		fmt.Printf("plan: %+v\n", *resp.JSON200)
		return
	}
	printErr(resp.HTTPResponse)
}

func handleDeletePlan(ctx context.Context, c *apiv2.ClientWithResponses, args []string) {
	if len(args) < 1 {
		log.Fatal("Usage: delete-plan <plan_id>")
	}
	id, _ := strconv.Atoi(args[0])
	resp, err := c.DeleteSubscriptionPlanWithResponse(ctx, id)
	if err != nil {
		log.Fatal(err)
	}
	if resp.HTTPResponse != nil && resp.HTTPResponse.StatusCode == 204 {
		fmt.Println("plan deleted")
		return
	}
	printErr(resp.HTTPResponse)
}

// ---------- subscriptions ----------

func handleListSubs(ctx context.Context, c *apiv2.ClientWithResponses, args []string) {
	if len(args) < 1 {
		log.Fatal("Usage: list-subs <user_id>")
	}
	userID, err := uuid.Parse(args[0])
	if err != nil {
		log.Fatal("invalid user_id")
	}
	resp, err := c.ListSubscriptionsWithResponse(ctx, &apiv2.ListSubscriptionsParams{UserId: apiv2.UserIdQuery(userID)})
	if err != nil {
		log.Fatal(err)
	}
	if resp.JSON200 != nil {
		for _, s := range *resp.JSON200 {
			fmt.Printf("%+v\n", s)
		}
		return
	}
	printErr(resp.HTTPResponse)
}

func handleCreateSub(ctx context.Context, c *apiv2.ClientWithResponses, args []string) {
	if len(args) < 2 {
		log.Fatal("Usage: create-sub <user_id> <plan_id>")
	}
	userID, err := uuid.Parse(args[0])
	if err != nil {
		log.Fatal("invalid user_id")
	}
	planID, _ := strconv.Atoi(args[1])
	resp, err := c.CreateSubscriptionWithResponse(ctx, apiv2.SubscriptionCreate{
		UserId: openapi_types.UUID(userID),
		PlanId: &planID,
	})
	if err != nil {
		log.Fatal(err)
	}
	if resp.JSON201 != nil {
		fmt.Printf("created subscription: %+v\n", *resp.JSON201)
		return
	}
	printErr(resp.HTTPResponse)
}

func handleGetSub(ctx context.Context, c *apiv2.ClientWithResponses, args []string) {
	if len(args) < 1 {
		log.Fatal("Usage: get-sub <sub_id>")
	}
	id, _ := strconv.Atoi(args[0])
	resp, err := c.GetSubscriptionByIdWithResponse(ctx, id)
	if err != nil {
		log.Fatal(err)
	}
	if resp.JSON200 != nil {
		fmt.Printf("subscription: %+v\n", *resp.JSON200)
		return
	}
	printErr(resp.HTTPResponse)
}

func handleDeleteSub(ctx context.Context, c *apiv2.ClientWithResponses, args []string) {
	if len(args) < 1 {
		log.Fatal("Usage: delete-sub <sub_id>")
	}
	id, _ := strconv.Atoi(args[0])
	resp, err := c.DeleteSubscriptionWithResponse(ctx, id)
	if err != nil {
		log.Fatal(err)
	}
	if resp.HTTPResponse != nil && resp.HTTPResponse.StatusCode == 204 {
		fmt.Println("subscription deleted")
		return
	}
	printErr(resp.HTTPResponse)
}

func handlePatchSub(ctx context.Context, c *apiv2.ClientWithResponses, args []string) {
	if len(args) < 3 {
		log.Fatal("Usage: patch-sub <sub_id> <field> <value>  (field: price|sub_type|start_date|end_date)")
	}
	id, _ := strconv.Atoi(args[0])
	field := args[1]
	value := args[2]

	var body apiv2.SubscriptionPatch
	switch field {
	case "price":
		p, _ := strconv.Atoi(value)
		body.Price = &p
	case "sub_type":
		st := apiv2.SubscriptionType(value)
		body.SubType = &st
	case "start_date":
		body.StartDate = &value
	case "end_date":
		body.EndDate = &value
	default:
		log.Fatalf("unknown field %q", field)
	}

	resp, err := c.PatchSubscriptionWithResponse(ctx, id, body)
	if err != nil {
		log.Fatal(err)
	}
	if resp.JSON200 != nil {
		fmt.Printf("updated subscription: %+v\n", *resp.JSON200)
		return
	}
	printErr(resp.HTTPResponse)
}

func handleApplyPromocode(ctx context.Context, c *apiv2.ClientWithResponses, args []string) {
	if len(args) < 2 {
		log.Fatal("Usage: apply-promocode <sub_id> <code>")
	}
	id, _ := strconv.Atoi(args[0])
	resp, err := c.PatchSubscriptionWithResponse(ctx, id, apiv2.SubscriptionPatch{Promocode: &args[1]})
	if err != nil {
		log.Fatal(err)
	}
	if resp.JSON200 != nil {
		fmt.Printf("promocode applied: %+v\n", *resp.JSON200)
		return
	}
	printErr(resp.HTTPResponse)
}

// ---------- promocodes ----------

func handleListPromocodes(ctx context.Context, c *apiv2.ClientWithResponses) {
	resp, err := c.ListPromocodesWithResponse(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}
	if resp.JSON200 != nil {
		for _, p := range *resp.JSON200 {
			fmt.Printf("%+v\n", p)
		}
		return
	}
	printErr(resp.HTTPResponse)
}

func handleCreatePromocode(ctx context.Context, c *apiv2.ClientWithResponses, args []string) {
	if len(args) < 4 {
		log.Fatal("Usage: create-promocode <service_id> <value> <discount> <max_uses>")
	}
	serviceID, _ := strconv.Atoi(args[0])
	discount, _ := strconv.Atoi(args[2])
	maxUses, _ := strconv.Atoi(args[3])
	resp, err := c.CreatePromocodeWithResponse(ctx, apiv2.PromocodeCreate{
		ServiceId: serviceID,
		Value:     args[1],
		Discount:  discount,
		MaxUses:   maxUses,
	})
	if err != nil {
		log.Fatal(err)
	}
	if resp.JSON201 != nil {
		fmt.Printf("created promocode: %+v\n", *resp.JSON201)
		return
	}
	printErr(resp.HTTPResponse)
}

func handleGetPromocode(ctx context.Context, c *apiv2.ClientWithResponses, args []string) {
	if len(args) < 1 {
		log.Fatal("Usage: get-promocode <promocode_id>")
	}
	id, _ := strconv.Atoi(args[0])
	resp, err := c.GetPromocodeByIdWithResponse(ctx, id)
	if err != nil {
		log.Fatal(err)
	}
	if resp.JSON200 != nil {
		fmt.Printf("promocode: %+v\n", *resp.JSON200)
		return
	}
	printErr(resp.HTTPResponse)
}

func handleDeletePromocode(ctx context.Context, c *apiv2.ClientWithResponses, args []string) {
	if len(args) < 1 {
		log.Fatal("Usage: delete-promocode <promocode_id>")
	}
	id, _ := strconv.Atoi(args[0])
	resp, err := c.DeletePromocodeWithResponse(ctx, id)
	if err != nil {
		log.Fatal(err)
	}
	if resp.HTTPResponse != nil && resp.HTTPResponse.StatusCode == 204 {
		fmt.Println("promocode deleted")
		return
	}
	printErr(resp.HTTPResponse)
}

// ---------- analytics ----------

func handleTotalCosts(ctx context.Context, c *apiv2.ClientWithResponses, args []string) {
	if len(args) < 3 {
		log.Fatal("Usage: total-costs <user_id> <start_date> <end_date>  (MM-YYYY)")
	}
	userID, err := uuid.Parse(args[0])
	if err != nil {
		log.Fatal("invalid user_id")
	}
	end := args[2]
	resp, err := c.GetTotalCostsWithResponse(ctx, &apiv2.GetTotalCostsParams{
		UserId:    apiv2.UserIdQuery(userID),
		StartDate: args[1],
		EndDate:   &end,
	})
	if err != nil {
		log.Fatal(err)
	}
	if resp.JSON200 != nil {
		fmt.Printf("total costs: %+v\n", *resp.JSON200)
		return
	}
	printErr(resp.HTTPResponse)
}

func handleStatsOverview(ctx context.Context, c *apiv2.ClientWithResponses) {
	resp, err := c.GetStatsOverviewWithResponse(ctx)
	if err != nil {
		log.Fatal(err)
	}
	if resp.JSON200 != nil {
		fmt.Printf("stats overview: %+v\n", *resp.JSON200)
		return
	}
	printErr(resp.HTTPResponse)
}

// ---------- helpers ----------

func printErr(resp *http.Response) {
	if resp != nil {
		fmt.Printf("error: HTTP %d\n", resp.StatusCode)
	} else {
		fmt.Println("error: empty response")
	}
}

func printHelp() {
	fmt.Println(`Subscription Service CLI (v2 API)
Commands:
  create-user <email> <password> <user_name> <age> <balance>
  get-user <user_id>
  list-services
  create-service <service_name>
  get-service <service_id>
  delete-service <service_id>
  list-plans
  create-plan <service_id> <name> <duration_days> <price>
  get-plan <plan_id>
  delete-plan <plan_id>
  list-subs <user_id>
  create-sub <user_id> <plan_id>
  get-sub <sub_id>
  patch-sub <sub_id> <price|sub_type|start_date|end_date> <value>
  apply-promocode <sub_id> <code>
  delete-sub <sub_id>
  list-promocodes
  create-promocode <service_id> <value> <discount> <max_uses>
  get-promocode <promocode_id>
  delete-promocode <promocode_id>
  total-costs <user_id> <start_date> <end_date>
  stats-overview
  help

Set API_URL to point to the backend (default http://localhost:8080).`)
}
