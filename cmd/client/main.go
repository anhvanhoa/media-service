package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	pb "github.com/anhvanhoa/sf-proto/gen/media/v1"

	"google.golang.org/grpc"
)

const chunkSize = 1024 * 32 // 32KB per chunk

func main() {
	// Kết nối tới gRPC server
	conn, err := grpc.Dial("localhost:50053", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Không thể kết nối: %v", err)
	}
	defer conn.Close()

	client := pb.NewMediaServiceClient(conn)

	// Gửi stream
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	uploadMediaStream(client, ctx)
	uploadMedia(client, ctx)

}

func uploadMediaStream(client pb.MediaServiceClient, ctx context.Context) {
	// Mở file để upload
	filePath := "C:/uploads/a.jpg"
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatalf("Không mở được file: %v", err)
	}
	defer file.Close()

	stream, err := client.UploadMediaStream(ctx)
	if err != nil {
		log.Fatalf("Không tạo được stream: %v", err)
	}

	// Gửi info một lần đầu tiên
	info := &pb.UploadMediaChunk{
		Data: &pb.UploadMediaChunk_Info{
			Info: &pb.UploadMediaStreamRequest{
				CreatedBy: "user123",
				Metadata:  map[string]string{"tag": "demo"},
			},
		},
	}
	if err := stream.Send(info); err != nil {
		log.Fatalf("Gửi info thất bại: %v", err)
	}

	// Sau đó gửi từng chunk
	buffer := make([]byte, chunkSize)
	for {
		n, err := file.Read(buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("Lỗi đọc file: %v", err)
		}

		if err := stream.Send(&pb.UploadMediaChunk{
			Data: &pb.UploadMediaChunk_Chunk{
				Chunk: buffer[:n],
			},
		}); err != nil {
			log.Fatalf("Lỗi gửi chunk: %v", err)
		}
	}

	// Đóng stream và nhận response
	res, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatalf("Lỗi nhận response: %v", err)
	}

	fmt.Printf("Upload thành công: %s (size: %d bytes)\n", res.Media.Name, res.Media.Size)
}

func uploadMedia(client pb.MediaServiceClient, ctx context.Context) {
	// Mở file để upload
	filePath := "C:/uploads/i.jpg"
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatalf("Không mở được file: %v", err)
	}
	defer file.Close()

	fileData, err := io.ReadAll(file)
	if err != nil {
		log.Fatalf("Lỗi đọc file: %v", err)
	}
	res, err := client.UploadMedia(ctx, &pb.UploadMediaRequest{
		FileName:  "Xin chào",
		CreatedBy: "user123",
		Metadata:  map[string]string{"tag": "demo"},
		FileData:  fileData,
	})

	if err != nil {
		log.Fatalf("Lỗi gửi media: %v", err)
	}

	fmt.Printf("Upload thành công: %s (size: %d bytes)\n", res.Media.Name, res.Media.Size)
}
