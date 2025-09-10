package main

import (
	"context"
	"io"
	"log"
	"os"
	"sync"

	"github.com/NXWeb-Group/vnc-containers/utils"
	"github.com/docker/docker/api/types/build"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func main() {

	port := "8080"
	networkName := "chrome-vnc-network"

	// Create a new Docker client using the default configuration
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Fatalf("Failed to create Docker client: %v", err)
	}

	buildContext, err := utils.CreateTarArchive("./docker/chrome")
	if err != nil {
		log.Fatalf("Failed to create build context: %v", err)
	}
	// Build options
	buildOptions := build.ImageBuildOptions{
		Dockerfile: "Dockerfile",
		Tags:       []string{"chrome-instance:latest"},
		Remove:     true,
	}

	ctx := context.Background()

	// Build the image
	response, err := cli.ImageBuild(ctx, buildContext, buildOptions)
	if err != nil {
		log.Fatalf("Failed to build image: %v", err)
	}
	defer response.Body.Close()

	// Stream the build output
	_, err = io.Copy(os.Stdout, response.Body)
	if err != nil {
		log.Fatalf("Failed to read build output: %v", err)
	}

	log.Println("Docker image built successfully!")

	app := fiber.New()

	app.Static("/", "./frontend/dist")

	unusedContainer := map[string]bool{}
	var mutex sync.Mutex

	app.Get("/api/createContainer", func(c *fiber.Ctx) error {
		id := uuid.NewString()
		containerName := "chrome-instance-" + id

		resp, err := cli.ContainerCreate(context.Background(), &container.Config{
			Image: "chrome-instance:latest",
		}, nil, &network.NetworkingConfig{
			EndpointsConfig: map[string]*network.EndpointSettings{
				networkName: {},
			},
		}, nil, containerName)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Failed to create container: " + err.Error())
		}

		err = cli.ContainerStart(context.Background(), resp.ID, container.StartOptions{})
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Failed to start container: " + err.Error())
		}

		mutex.Lock()
		unusedContainer[containerName] = true
		mutex.Unlock()

		go utils.StartContainerTimer(cli, containerName, unusedContainer, &mutex)

		return c.JSON(fiber.Map{"id": id})
	})

	app.Get("/websockify/:id", websocket.New(func(c *websocket.Conn) {
		id := c.Params("id")
		log.Println("WebSocket connection established")
		utils.HandleWebSocket(c, id, cli, unusedContainer, &mutex)
	}))

	log.Println("Server starting on", port)
	log.Fatal(app.Listen(":" + port))

}
