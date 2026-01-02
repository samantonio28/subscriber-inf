package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type SubRepo struct {
	p *pgxpool.Pool
}

func NewSubRepo(p *pgxpool.Pool) (*SubRepo, error) {
	if p == nil {
		return nil, fmt.Errorf("invalid pool")
	}
	return &SubRepo{p: p}, nil
}

type Driver struct {
	DriverID    int
	FIO         string
	BirthDate   time.Time
	OnboardedAt time.Time
	Region      string
}

type Route struct {
	DriverID int
	RaceDate time.Time
	RaceTime time.Time
	WeekDay  string
	Event    bool
}

func (r *SubRepo) GetRegionsWithMinExperienceDB() {
	fmt.Println("\n1. Регионы, где есть водители с минимальным стажем (база данных)")

	query := `
		SELECT Region 
		FROM driver 
		WHERE DATE_PART('year', AGE(CURRENT_DATE, OnboardedAt)) = (
			SELECT MIN(DATE_PART('year', AGE(CURRENT_DATE, OnboardedAt))) 
			FROM driver
		)
		GROUP BY Region
		ORDER BY Region`

	rows, err := r.p.Query(context.Background(), query)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var regions []string
	for rows.Next() {
		var region string
		if err := rows.Scan(&region); err != nil {
			log.Fatal(err)
		}
		regions = append(regions, region)
	}

	fmt.Printf("Регионы: %v\n", regions)
}

func (r *SubRepo) GetRegionsWithMinExperienceApp() {
	fmt.Println("\n1. Регионы, где есть водители с минимальным стажем (прога)")
	query := `SELECT DriverID, FIO, BirthDate, OnboardedAt, Region FROM driver`
	rows, err := r.p.Query(context.Background(), query)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var drivers []Driver
	var minExpDays int64 = 365 * 100

	for rows.Next() {
		var d Driver
		if err := rows.Scan(&d.DriverID, &d.FIO, &d.BirthDate, &d.OnboardedAt, &d.Region); err != nil {
			log.Fatal(err)
		}
		drivers = append(drivers, d)

		expDays := time.Since(d.OnboardedAt).Hours() / 24
		if int64(expDays) < minExpDays {
			minExpDays = int64(expDays)
		}
	}

	regionMap := make(map[string]bool)
	for _, d := range drivers {
		expDays := time.Since(d.OnboardedAt).Hours() / 24
		if int64(expDays) == minExpDays {
			regionMap[d.Region] = true
		}
	}

	var regions []string
	for region := range regionMap {
		regions = append(regions, region)
	}

	fmt.Printf("Минимальный стаж: %d дней\n", minExpDays)
	fmt.Printf("Регионы: %v\n", regions)
}

func (r *SubRepo) GetDriversInRouteNovToMayDB() {
	fmt.Println("\n2. Водители в рейсе с ноября по май (фильтрация в базе данных)")
	query := `
		SELECT DISTINCT d.DriverID, d.FIO
		FROM driver d
		INNER JOIN route r ON d.DriverID = r.DriverID
		WHERE r.RaceDate >= '2023-11-01' AND r.RaceDate <= '2024-05-31'
		GROUP BY d.DriverID, d.FIO
		HAVING COUNT(DISTINCT r.RaceDate) = (
			SELECT COUNT(DISTINCT RaceDate) 
			FROM route 
			WHERE RaceDate >= '2023-11-01' AND RaceDate <= '2024-05-31'
		)
		ORDER BY d.DriverID`

	rows, err := r.p.Query(context.Background(), query)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	fmt.Println("Водители в рейсе с ноября по май:")
	for rows.Next() {
		var driverID int
		var fio string
		if err := rows.Scan(&driverID, &fio); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("ID: %d, ФИО: %s\n", driverID, fio)
	}
}

func (r *SubRepo) GetDriversInRouteNovToMayApp() {
	fmt.Println("\n2. Водители в рейсе с ноября по май (фильтрация в программе)")

	query := `SELECT d.DriverID, d.FIO, r.RaceDate 
			  FROM driver d
			  INNER JOIN route r ON d.DriverID = r.DriverID
			  WHERE r.RaceDate >= '2023-11-01' AND r.RaceDate <= '2024-05-31'
			  ORDER BY d.DriverID, r.RaceDate`

	rows, err := r.p.Query(context.Background(), query)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	driverDates := make(map[int][]time.Time)
	driverNames := make(map[int]string)

	for rows.Next() {
		var driverID int
		var fio string
		var raceDate time.Time
		if err := rows.Scan(&driverID, &fio, &raceDate); err != nil {
			log.Fatal(err)
		}
		driverDates[driverID] = append(driverDates[driverID], raceDate)
		driverNames[driverID] = fio
	}

	startDate := time.Date(2023, 11, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2024, 5, 31, 0, 0, 0, 0, time.UTC)
	allDays := int(endDate.Sub(startDate).Hours()/24) + 1

	fmt.Println("Водители в рейсе с ноября по май:")
	for driverID, dates := range driverDates {
		daysMap := make(map[string]bool)
		for _, date := range dates {
			daysMap[date.Format("2006-01-02")] = true
		}

		continuous := true
		currentDate := startDate
		for i := 0; i < allDays; i++ {
			if !daysMap[currentDate.Format("2006-01-02")] {
				continuous = false
				break
			}
			currentDate = currentDate.Add(24 * time.Hour)
		}

		if continuous {
			fmt.Printf("ID: %d, ФИО: %s\n", driverID, driverNames[driverID])
		}
	}
}

func (r *SubRepo) GetDriversWithBreakOver100DaysDB() {
	fmt.Println("\n 3. Водители с перерывом > 100 дней (фильтрация в БД)")

	query := `
		SELECT DISTINCT d.DriverID, d.FIO
		FROM driver d
		INNER JOIN (
			SELECT DriverID, RaceDate,
				   LAG(RaceDate) OVER (PARTITION BY DriverID ORDER BY RaceDate) as prev_date
			FROM route
		) r ON d.DriverID = r.DriverID
		WHERE r.prev_date IS NOT NULL 
		  AND (r.RaceDate - r.prev_date) > 100
		ORDER BY d.DriverID`

	rows, err := r.p.Query(context.Background(), query)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	fmt.Println("Водители с перерывом > 100 дней:")
	for rows.Next() {
		var driverID int
		var fio string
		if err := rows.Scan(&driverID, &fio); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("ID: %d, ФИО: %s\n", driverID, fio)
	}
}

func (r *SubRepo) GetDriversWithBreakOver100DaysApp() {
	fmt.Println("\n 3. Водители с перерывом > 100 дней (фильтрация в программе)")

	query := `SELECT DriverID, RaceDate FROM route ORDER BY DriverID, RaceDate`
	rows, err := r.p.Query(context.Background(), query)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	driverInfo := make(map[int]bool)
	var currentDriverID int = -1
	var lastDate time.Time
	firstRecord := true

	for rows.Next() {
		var driverID int
		var raceDate time.Time
		if err := rows.Scan(&driverID, &raceDate); err != nil {
			log.Fatal(err)
		}

		if driverID != currentDriverID {
			currentDriverID = driverID
			lastDate = raceDate
			firstRecord = true
			continue
		}

		if !firstRecord {
			daysDiff := int(raceDate.Sub(lastDate).Hours() / 24)
			if daysDiff > 100 {
				driverInfo[driverID] = true
			}
		}

		lastDate = raceDate
		firstRecord = false
	}

	if len(driverInfo) > 0 {
		var ids []interface{}
		for id := range driverInfo {
			ids = append(ids, id)
		}

		query = `SELECT DriverID, FIO FROM driver WHERE DriverID = ANY($1)`
		rows, err := r.p.Query(context.Background(), query, ids)
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()

		fmt.Println("Водители с перерывом > 100 дней:")
		for rows.Next() {
			var driverID int
			var fio string
			if err := rows.Scan(&driverID, &fio); err != nil {
				log.Fatal(err)
			}
			fmt.Printf("ID: %d, ФИО: %s\n", driverID, fio)
		}
	} else {
		fmt.Println("Водители с перерывом > 100 дней не найдены")
	}
}

func main() {
	connStr := "postgres://postgres:secret@localhost:8000/dev?sslmode=disable"
	pool, err := pgxpool.New(context.Background(), connStr)
	if err != nil {
		log.Fatal("Ошибка подключения к БД:", err)
	}
	defer pool.Close()

	if err := pool.Ping(context.Background()); err != nil {
		log.Fatal("Ошибка ping к БД:", err)
	}
	fmt.Println("Успешное подключение к БД")

	repo, err := NewSubRepo(pool)
	if err != nil {
		log.Fatal(err)
	}

	repo.GetRegionsWithMinExperienceDB()
	repo.GetRegionsWithMinExperienceApp()
	repo.GetDriversInRouteNovToMayDB()
	repo.GetDriversInRouteNovToMayApp()
	repo.GetDriversWithBreakOver100DaysDB()
	repo.GetDriversWithBreakOver100DaysApp()
}
