package main

import (
	"context"
	"flag"
	"github.com/ge-fei-fan/clouddrive2api/clouddrive"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"log"
)

//const DEFAULT_BUFFER_SIZE = 8192 // 默认缓冲区大小
//
//func upload(client *CloudDriveClient, path string, file io.Reader) error {
//	// 分割路径，获取目录和文件名
//	dir, name := filepath.Split(path)
//
//	// 创建一个文件并获取文件句柄
//	fh, err := client.CreateFile(dir, name)
//	if err != nil {
//		return err
//	}
//
//	defer func() {
//		// 在函数返回前确保关闭文件
//		client.CloseFile(fh)
//	}()
//
//	// 如果传入了文件对象
//	if file != nil {
//		offset := int64(0)
//		buffer := make([]byte, DEFAULT_BUFFER_SIZE)
//		// 循环读取文件内容并写入到云端文件
//		for {
//			n, err := file.Read(buffer)
//			if err != nil && err != io.EOF {
//				return err
//			}
//			if n == 0 {
//				break
//			}
//			// 将文件内容写入到云端文件
//			err = client.WriteToFile(fh, offset, buffer[:n])
//			if err != nil {
//				return err
//			}
//			offset += int64(n)
//		}
//	}
//
//	// 关闭并上传文件
//	return client.CloseFile(fh)
//}

const (
	defaultName = "world"
)

var (
	addr = flag.String("addr", "127.0.0.1:19798", "the address to connect to")
	name = flag.String("name", defaultName, "Name to greet")
)

func main() {
	flag.Parse()
	// 连接到server端，此处禁用安全传输
	conn, err := grpc.Dial(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := clouddrive.NewCloudDriveFileSrvClient(conn)

	// 执行RPC调用并打印收到的响应数据
	//ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	//defer cancel()

	header := metadata.New(map[string]string{
		"authorization": "Bearer b20ce9fd-010e-4355-8115-f7e8de50e229",
	})
	//md := metadata.Pairs("authorization", "Bearer b20ce9fd-010e-4355-8115-f7e8de50e229")

	// 创建一个带有请求头的 context
	ctx := metadata.NewOutgoingContext(context.Background(), header)

	//r, err := c.GetToken(ctx, &clouddrive.GetTokenRequest{UserName: "13276022144@163.com", Password: "987lxgff."})
	//if err != nil {
	//	log.Fatalf("could not greet: %v", err)
	//}
	//log.Printf("Greeting: %s", r.GetToken())
	res, err := c.AddOfflineFiles(ctx, &clouddrive.AddOfflineFileRequest{Urls: "magnet:?xt=urn:btih:50e1314f59bd943237435bde9733cb5bd5066669", ToFolder: "/115/tg"})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	log.Printf("Greeting: %s", res.GetResultFilePaths())
}
