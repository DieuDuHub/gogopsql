package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
)

// Define a struct that matches the expected JSON data structure
type MetricmaniaStruct struct {
	Testcase  string `json:"testcase"`
	SAPCode   string `json:"sapCode"`
	Tenant    string `json:"tenant"`
	Project   string `json:"project"`
	App       string `json:"app"`
	Statistic string `json:"statistic"`
}

func readData(conn *pgx.Conn) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")
		intID, err := strconv.Atoi(id)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid ID format",
			})
		}

		var testcase MetricmaniaStruct
		err = conn.QueryRow(context.Background(), `
			SELECT testcase, sapcode, tenant, project, app, statistic
			FROM testcasetable
			WHERE ID = $1`, intID).Scan(
			&testcase.Testcase, &testcase.SAPCode, &testcase.Tenant, &testcase.Project, &testcase.App, &testcase.Statistic)
		if err != nil {
			log.Println(err)
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Testcase not found",
			})
		}

		return c.JSON(testcase)
	}
}

func insertData(conn *pgx.Conn) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var metric MetricmaniaStruct

		// Parse the JSON request body into the MetricmaniaStruct
		if err := c.BodyParser(&metric); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Cannot parse JSON",
			})
		}

		var id int
		// Insert the parsed data into the PostgreSQL table
		err := conn.QueryRow(context.Background(), `
	 INSERT INTO testcasetable (testcase, sapcode, tenant, project, app, statistic)
	 VALUES ($1, $2, $3, $4, $5, $6) RETURNING ID`,
			metric.Testcase, metric.SAPCode, metric.Tenant, metric.Project, metric.App, metric.Statistic).Scan(&id)
		if err != nil {
			log.Println(err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to insert data into database",
			})
		}

		// Return the ID of the inserted row
		return c.JSON(fiber.Map{
			"id": id,
		})
	}
}

func main() {
	app := fiber.New()

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World!")
	})

	// urlExample := "postgres://username:password@localhost:5432/database_name"
	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close(context.Background())

	app.Get("/metricmania/:id", readData(conn))

	app.Post("/metricmania", insertData(conn))

	log.Fatal(app.Listen(":3000"))
}
