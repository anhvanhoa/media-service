package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	pb "github.com/anhvanhoa/sf-proto/gen/media/v1"

	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var serverAddress string

func init() {
	viper.SetConfigFile("dev.config.yaml")
	viper.ReadInConfig()
	// Mặc định sử dụng localhost:50053 nếu không có config
	host := viper.GetString("host_grpc")
	port := viper.GetString("port_grpc")
	if host == "" {
		host = "localhost"
	}
	if port == "" {
		port = "50053"
	}
	serverAddress = fmt.Sprintf("%s:%s", host, port)
}

type MediaServiceClient struct {
	mediaClient pb.MediaServiceClient
	conn        *grpc.ClientConn
}

func NewMediaServiceClient(address string) (*MediaServiceClient, error) {
	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to gRPC server: %v", err)
	}

	return &MediaServiceClient{
		mediaClient: pb.NewMediaServiceClient(conn),
		conn:        conn,
	}, nil
}

func (c *MediaServiceClient) Close() {
	if c.conn != nil {
		c.conn.Close()
	}
}

// --- Helper để làm sạch input ---
func cleanInput(s string) string {
	return strings.ToValidUTF8(strings.TrimSpace(s), "")
}

// ================== Media Service Tests ==================

func (c *MediaServiceClient) TestUploadMedia() {
	fmt.Println("\n=== Test Upload Media (Non-streaming) ===")

	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter file path: ")
	filePath, _ := reader.ReadString('\n')
	filePath = cleanInput(filePath)

	fmt.Print("Enter file name (display name): ")
	fileName, _ := reader.ReadString('\n')
	fileName = cleanInput(fileName)

	fmt.Print("Enter output file: ")
	outputFile, _ := reader.ReadString('\n')
	outputFile = cleanInput(outputFile)

	fmt.Print("Enter created by: ")
	createdBy, _ := reader.ReadString('\n')
	createdBy = cleanInput(createdBy)

	// Đọc file
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		return
	}
	defer file.Close()

	fileData, err := io.ReadAll(file)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := c.mediaClient.UploadMedia(ctx, &pb.UploadMediaRequest{
		FileName:   fileName,
		CreatedBy:  createdBy,
		Metadata:   map[string]string{"tag": "test"},
		FileData:   fileData,
		OutputFile: outputFile,
	})
	if err != nil {
		fmt.Printf("Error calling UploadMedia: %v\n", err)
		return
	}

	fmt.Printf("Upload Media result:\n")
	fmt.Printf("ID: %s\n", resp.Media.Id)
	fmt.Printf("Name: %s\n", resp.Media.Name)
	fmt.Printf("Size: %d bytes\n", resp.Media.Size)
	fmt.Printf("URL: %s\n", resp.Media.Url)
	fmt.Printf("MIME Type: %s\n", resp.Media.MimeType)
	fmt.Printf("Type: %s\n", resp.Media.Type)
	fmt.Printf("Processing Status: %s\n", resp.Media.ProcessingStatus)
}

func (c *MediaServiceClient) TestUploadMediaStream() {
	fmt.Println("\n=== Test Upload Media Stream ===")

	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter file path: ")
	filePath, _ := reader.ReadString('\n')
	filePath = cleanInput(filePath)

	fmt.Print("Enter file name (display name): ")
	fileName, _ := reader.ReadString('\n')
	fileName = cleanInput(fileName)

	fmt.Print("Enter created by: ")
	createdBy, _ := reader.ReadString('\n')
	createdBy = cleanInput(createdBy)

	// Mở file
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		return
	}
	defer file.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	stream, err := c.mediaClient.UploadMediaStream(ctx)
	if err != nil {
		fmt.Printf("Error creating stream: %v\n", err)
		return
	}

	// Gửi info đầu tiên
	info := &pb.UploadMediaChunk{
		Data: &pb.UploadMediaChunk_Info{
			Info: &pb.UploadMediaStreamRequest{
				FileName:  fileName,
				CreatedBy: createdBy,
				Metadata:  map[string]string{"tag": "stream-test"},
			},
		},
	}
	if err := stream.Send(info); err != nil {
		fmt.Printf("Error sending info: %v\n", err)
		return
	}

	// Gửi file chunks
	chunkSize := 1024 * 32 // 32KB
	buffer := make([]byte, chunkSize)
	for {
		n, err := file.Read(buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Printf("Error reading file: %v\n", err)
			return
		}

		if err := stream.Send(&pb.UploadMediaChunk{
			Data: &pb.UploadMediaChunk_Chunk{
				Chunk: buffer[:n],
			},
		}); err != nil {
			fmt.Printf("Error sending chunk: %v\n", err)
			return
		}
	}

	// Nhận response
	resp, err := stream.CloseAndRecv()
	if err != nil {
		fmt.Printf("Error receiving response: %v\n", err)
		return
	}

	fmt.Printf("Upload Media Stream result:\n")
	fmt.Printf("ID: %s\n", resp.Media.Id)
	fmt.Printf("Name: %s\n", resp.Media.Name)
	fmt.Printf("Size: %d bytes\n", resp.Media.Size)
	fmt.Printf("URL: %s\n", resp.Media.Url)
	fmt.Printf("MIME Type: %s\n", resp.Media.MimeType)
	fmt.Printf("Type: %s\n", resp.Media.Type)
	fmt.Printf("Processing Status: %s\n", resp.Media.ProcessingStatus)
}

func (c *MediaServiceClient) TestGetMedia() {
	fmt.Println("\n=== Test Get Media ===")

	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter media ID: ")
	id, _ := reader.ReadString('\n')
	id = cleanInput(id)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := c.mediaClient.GetMedia(ctx, &pb.GetMediaRequest{
		Id: id,
	})
	if err != nil {
		fmt.Printf("Error calling GetMedia: %v\n", err)
		return
	}

	fmt.Printf("Get Media result:\n")
	fmt.Printf("ID: %s\n", resp.Media.Id)
	fmt.Printf("Name: %s\n", resp.Media.Name)
	fmt.Printf("Size: %d bytes\n", resp.Media.Size)
	fmt.Printf("URL: %s\n", resp.Media.Url)
	fmt.Printf("MIME Type: %s\n", resp.Media.MimeType)
	fmt.Printf("Type: %s\n", resp.Media.Type)
	fmt.Printf("Width: %d\n", resp.Media.Width)
	fmt.Printf("Height: %d\n", resp.Media.Height)
	fmt.Printf("Duration: %d\n", resp.Media.Duration)
	fmt.Printf("Processing Status: %s\n", resp.Media.ProcessingStatus)
	fmt.Printf("Created By: %s\n", resp.Media.CreatedBy)
	if resp.Media.CreatedAt != nil {
		fmt.Printf("Created At: %s\n", resp.Media.CreatedAt.AsTime().Format(time.RFC3339))
	}
	if resp.Media.UpdatedAt != nil {
		fmt.Printf("Updated At: %s\n", resp.Media.UpdatedAt.AsTime().Format(time.RFC3339))
	}
	if len(resp.Media.Metadata) > 0 {
		fmt.Printf("Metadata:\n")
		for k, v := range resp.Media.Metadata {
			fmt.Printf("  %s: %s\n", k, v)
		}
	}
}

func (c *MediaServiceClient) TestListMedia() {
	fmt.Println("\n=== Test List Media ===")

	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter page (default 1): ")
	pageStr, _ := reader.ReadString('\n')
	pageStr = cleanInput(pageStr)
	page := int32(1)
	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil {
			page = int32(p)
		}
	}

	fmt.Print("Enter limit (default 10): ")
	limitStr, _ := reader.ReadString('\n')
	limitStr = cleanInput(limitStr)
	limit := int32(10)
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = int32(l)
		}
	}

	fmt.Print("Enter created by (optional): ")
	createdBy, _ := reader.ReadString('\n')
	createdBy = cleanInput(createdBy)

	fmt.Print("Enter type filter (optional): ")
	typeFilter, _ := reader.ReadString('\n')
	typeFilter = cleanInput(typeFilter)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := c.mediaClient.ListMedia(ctx, &pb.ListMediaRequest{
		Limit:     limit,
		Offset:    (page - 1) * limit,
		CreatedBy: createdBy,
		Type:      typeFilter,
	})
	if err != nil {
		fmt.Printf("Error calling ListMedia: %v\n", err)
		return
	}

	fmt.Printf("List Media result:\n")
	fmt.Printf("Total: %d\n", resp.Total)
	fmt.Printf("Page: %d\n", page)
	fmt.Printf("Limit: %d\n", limit)
	fmt.Printf("Media items:\n")
	for i, media := range resp.Media {
		fmt.Printf("  [%d] ID: %s, Name: %s, Size: %d, Type: %s, Status: %s\n",
			i+1, media.Id, media.Name, media.Size, media.Type, media.ProcessingStatus)
	}
}

func (c *MediaServiceClient) TestUpdateMedia() {
	fmt.Println("\n=== Test Update Media ===")

	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter media ID: ")
	id, _ := reader.ReadString('\n')
	id = cleanInput(id)

	fmt.Print("Enter new name (optional): ")
	name, _ := reader.ReadString('\n')
	name = cleanInput(name)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := c.mediaClient.UpdateMedia(ctx, &pb.UpdateMediaRequest{
		Id:       id,
		Name:     name,
		Metadata: map[string]string{"updated": "true"},
	})
	if err != nil {
		fmt.Printf("Error calling UpdateMedia: %v\n", err)
		return
	}

	fmt.Printf("Update Media result:\n")
	fmt.Printf("ID: %s\n", resp.Media.Id)
	fmt.Printf("Name: %s\n", resp.Media.Name)
	fmt.Printf("Size: %d bytes\n", resp.Media.Size)
	fmt.Printf("Processing Status: %s\n", resp.Media.ProcessingStatus)
	if resp.Media.UpdatedAt != nil {
		fmt.Printf("Updated At: %s\n", resp.Media.UpdatedAt.AsTime().Format(time.RFC3339))
	}
}

func (c *MediaServiceClient) TestDeleteMedia() {
	fmt.Println("\n=== Test Delete Media ===")

	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter media ID to delete: ")
	id, _ := reader.ReadString('\n')
	id = cleanInput(id)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := c.mediaClient.DeleteMedia(ctx, &pb.DeleteMediaRequest{
		Id: id,
	})
	if err != nil {
		fmt.Printf("Error calling DeleteMedia: %v\n", err)
		return
	}

	fmt.Printf("Delete Media result:\n")
	fmt.Printf("Success: %t\n", resp.Success)
}

func (c *MediaServiceClient) TestGetMediaVariants() {
	fmt.Println("\n=== Test Get Media Variants ===")

	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter media ID: ")
	mediaId, _ := reader.ReadString('\n')
	mediaId = cleanInput(mediaId)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := c.mediaClient.GetMediaVariants(ctx, &pb.GetMediaVariantsRequest{
		MediaId: mediaId,
	})
	if err != nil {
		fmt.Printf("Error calling GetMediaVariants: %v\n", err)
		return
	}

	fmt.Printf("Get Media Variants result:\n")
	fmt.Printf("Media ID: %s\n", mediaId)
	fmt.Printf("Variants:\n")
	for i, variant := range resp.Variants {
		fmt.Printf("  [%d] ID: %s, Type: %s, Size: %s, Format: %s\n",
			i+1, variant.Id, variant.Type, variant.Size, variant.Format)
		fmt.Printf("      URL: %s\n", variant.Url)
		fmt.Printf("      File Size: %d bytes, Dimensions: %dx%d, Quality: %d\n",
			variant.FileSize, variant.Width, variant.Height, variant.Quality)
	}
}

func (c *MediaServiceClient) TestProcessMedia() {
	fmt.Println("\n=== Test Process Media ===")

	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter media ID: ")
	mediaId, _ := reader.ReadString('\n')
	mediaId = cleanInput(mediaId)

	fmt.Print("Enter process type (resize, thumbnail, compress, etc.): ")
	processType, _ := reader.ReadString('\n')
	processType = cleanInput(processType)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := c.mediaClient.ProcessMedia(ctx, &pb.ProcessMediaRequest{
		MediaId: mediaId,
	})
	if err != nil {
		fmt.Printf("Error calling ProcessMedia: %v\n", err)
		return
	}

	fmt.Printf("Process Media result:\n")
	fmt.Printf("Success: %t\n", resp.Success)
}

// ================== Menu Functions ==================

func printMainMenu() {
	fmt.Println("\n=== gRPC Media Service Test Client ===")
	fmt.Println("1. Upload Media (Non-streaming)")
	fmt.Println("2. Upload Media Stream")
	fmt.Println("3. Get Media")
	fmt.Println("4. List Media")
	fmt.Println("5. Update Media")
	fmt.Println("6. Delete Media")
	fmt.Println("7. Get Media Variants")
	fmt.Println("8. Process Media")
	fmt.Println("0. Exit")
	fmt.Print("Enter your choice: ")
}

func main() {
	fmt.Printf("Connecting to gRPC server at %s...\n", serverAddress)
	client, err := NewMediaServiceClient(serverAddress)
	if err != nil {
		log.Fatalf("Failed to create gRPC client: %v", err)
	}
	defer client.Close()

	fmt.Println("Connected successfully!")

	reader := bufio.NewReader(os.Stdin)

	for {
		printMainMenu()
		choice, _ := reader.ReadString('\n')
		choice = cleanInput(choice)

		switch choice {
		case "1":
			client.TestUploadMedia()
		case "2":
			client.TestUploadMediaStream()
		case "3":
			client.TestGetMedia()
		case "4":
			client.TestListMedia()
		case "5":
			client.TestUpdateMedia()
		case "6":
			client.TestDeleteMedia()
		case "7":
			client.TestGetMediaVariants()
		case "8":
			client.TestProcessMedia()
		case "0":
			fmt.Println("Goodbye!")
			return
		default:
			fmt.Println("Invalid choice. Please try again.")
		}
	}
}
