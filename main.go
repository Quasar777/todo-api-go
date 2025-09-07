package main

import (
	"fmt"
	"strings"
	"sync"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type Task struct {
	ID uuid.UUID `json:"id"`
	Title string `json:"title"`
	Body string `json:"body"`
	Status string `json:"status"`
}

var (
	tasks = make([]Task, 0)
	mu = &sync.RWMutex{}
)

func main() {
	app := fiber.New()

	startConf()

	// получить список всех задач
	app.Get("/api/tasks", func (c *fiber.Ctx) error  {
		mu.RLock()
		out := make([]Task, len(tasks))
		copy(out, tasks)
		mu.RUnlock()

		return c.Status(fiber.StatusOK).JSON(out)
	})

	// получить задачу по ID
	app.Get("/api/tasks/:id", func (c *fiber.Ctx) error {
		id, err := uuid.Parse(c.Params("id"))
		if err != nil {
			return invalidIDError(c)
		}

		mu.RLock()
		defer mu.RUnlock()

		for _, t := range tasks {
			if t.ID == id {
				return c.Status(fiber.StatusOK).JSON(t)
			}
		}

		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "task not found",
		})
	})	

	// создать задачу
	app.Post("/api/tasks", func (c *fiber.Ctx) error {
		if !strings.HasPrefix(c.Get("Content-Type"), "application/json") {
			return c.Status(fiber.StatusUnsupportedMediaType).JSON(fiber.Map{
				"error": "Content-Type must be application/json",
			})
		}

		var newTask Task
		if err := c.BodyParser(&newTask); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid JSON body",
			})
		}

		if newTask.Body == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "field 'body' is required",
			})
		}
		if newTask.Title == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "field 'title' is required",
			})
		}

		newTask.ID = uuid.New()
		newTask.Status = "no"

		mu.Lock()
		defer mu.Unlock()
		tasks = append(tasks, newTask)

		c.Location(fmt.Sprintf("/api/tasks/%s", newTask.ID))
		return c.Status(fiber.StatusCreated).JSON(newTask)
	})

	// пометить задачу как выполненную
	app.Patch("/api/tasks/:id", func (c *fiber.Ctx) error {
		id, err := uuid.Parse(c.Params("id"))
		if err != nil {
			return invalidIDError(c)
		}

		mu.Lock()
		defer mu.Unlock()

		for i, e := range tasks {
			if e.ID == id {
				tasks[i].Status = "yes"
				return c.Status(fiber.StatusOK).JSON(tasks[i])
			}
		}

		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "task not found",
		})
	})

	// обновить задачу
	app.Put("/api/tasks/:id", func (c *fiber.Ctx) error {
		if !strings.HasPrefix(c.Get("Content-Type"), "application/json") {
			return c.Status(fiber.StatusUnsupportedMediaType).JSON(fiber.Map{
				"error": "Content-Type must be application/json",
			})
		}

		id, err := uuid.Parse(c.Params("id"))
		if err != nil {
			return invalidIDError(c)
		}

		var newTask Task
		if err := c.BodyParser(&newTask); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid JSON body",
			})
		}

		if newTask.Body == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "field 'body' is required",
			})
		}
		if newTask.Title == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "field 'title' is required",
			})
		}

		mu.Lock()
		defer mu.Unlock()

		for i, t := range tasks {
			if t.ID == id {
				newTask.ID = tasks[i].ID
				newTask.Status = tasks[i].Status
				tasks[i] = newTask
				return c.Status(fiber.StatusOK).JSON(tasks[i])
			}
		}

		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "task not found",
		})
	})

	// удалить задачу
	app.Delete("/api/tasks/:id", func (c *fiber.Ctx) error {
		id, err := uuid.Parse(c.Params("id"))
		if err != nil {
			return invalidIDError(c)
		}

		isExist := false 
		index := -1

		mu.RLock()
		for i, e := range tasks {
			if e.ID == id {
				isExist = true
				index = i
				break
			}
		}
		mu.RUnlock()

		if isExist {
			mu.Lock()
			defer mu.Unlock()

			tasks = append(tasks[:index], tasks[index + 1:]... )
			return c.Status(fiber.StatusOK).JSON(fiber.Map{
				"message": "deleted task with id " + id.String(),
			})
		} else {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "task not found",
			})
		}
		
	})

	app.Listen(":8080")
}

func invalidIDError(c *fiber.Ctx) error {
	return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
		"error": "invalid id",
	})
}

func startConf() {
	toAdd := []Task{
		{
			ID: uuid.New(),
			Title: "buy groceries",
			Body: "list: banana, apple, bread",
			Status: "no",
		},
		{
			ID: uuid.New(),
			Title: "go to gym",
			Body: "It's a leg-day today",
			Status: "no",
		},
	}

	tasks = append(tasks, toAdd...)
}