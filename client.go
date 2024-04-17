package clouddrive2api

import (
	"bufio"
	"context"
	"fmt"
	"github.com/ge-fei-fan/clouddrive2api/clouddrive"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"io"
	"os"
	"path/filepath"
	"time"
)

const DEFAULT_BUFFER_SIZE = 8192 // 默认缓冲区大小

type Client struct {
	addr              string
	conn              *grpc.ClientConn
	cd                clouddrive.CloudDriveFileSrvClient
	contextWithHeader context.Context
	username          string
	password          string
	offlineFolder     string
	uploadFolder      string
}

func NewClient(addr, username, password string) *Client {
	c := Client{
		addr:              addr,
		conn:              nil,
		cd:                nil,
		contextWithHeader: nil,
		username:          username,
		password:          password,
		offlineFolder:     "/115/云下载",
		uploadFolder:      "/115/tg",
	}
	return &c
}

func (c *Client) Close() {
	c.conn.Close()
}

func (c *Client) Login() error {
	conn, err := grpc.Dial(c.addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}
	c.conn = conn
	c.cd = clouddrive.NewCloudDriveFileSrvClient(c.conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		cancel()
	}()
	r, err := c.cd.GetToken(ctx, &clouddrive.GetTokenRequest{UserName: c.username, Password: c.password})
	if err != nil {
		return err
	}
	header := metadata.New(map[string]string{
		"authorization": "Bearer " + r.GetToken(),
	})
	c.contextWithHeader = metadata.NewOutgoingContext(context.Background(), header)
	return nil
}

func (c *Client) AddOfflineFiles(url string) ([]string, error) {
	res, err := c.cd.AddOfflineFiles(c.contextWithHeader, &clouddrive.AddOfflineFileRequest{Urls: url, ToFolder: c.offlineFolder})
	if err != nil {
		return nil, err
	}

	return res.GetResultFilePaths(), nil
}

func (c *Client) Upload(FilePath string) error {
	var createFileResult *clouddrive.CreateFileResult
	var file *os.File
	fileName := filepath.Base(FilePath)
	defer func() {
		if file != nil {
			file.Close()
		}
		if createFileResult != nil {
			_, _ = c.cd.CloseFile(c.contextWithHeader, &clouddrive.CloseFileRequest{FileHandle: createFileResult.FileHandle})
		}
	}()
	createFileResult, err := c.cd.CreateFile(c.contextWithHeader, &clouddrive.CreateFileRequest{ParentPath: c.uploadFolder, FileName: fileName})
	if err != nil {
		return err
	}
	// 打开文件
	file, err = os.Open(FilePath)
	if err != nil {
		return err
	}
	// 如果传入了文件对象
	if file != nil {
		offset := uint64(0)
		// 循环读取文件内容并写入到云端文件
		for {
			reader := bufio.NewReader(file)
			data := make([]byte, DEFAULT_BUFFER_SIZE)
			n, err := reader.Read(data)
			if err != nil && err != io.EOF {
				return err
			}
			if n == 0 {
				break
			}
			// 将文件内容写入到云端文件
			res, err := c.cd.WriteToFile(c.contextWithHeader, &clouddrive.WriteFileRequest{FileHandle: createFileResult.FileHandle, StartPos: offset, Length: uint64(n), Buffer: data[:n], CloseFile: false})
			if err != nil {
				return err
			}
			fmt.Println(res.GetBytesWritten())
			offset += uint64(n)
		}
	}
	return nil
}
