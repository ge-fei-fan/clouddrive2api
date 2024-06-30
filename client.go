package clouddrive2api

import (
	"bufio"
	"context"
	"errors"
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
	OfflineFolder     string
	UploadFolder      string
}

func NewClient(addr, username, password string) *Client {
	c := Client{
		addr:              addr,
		conn:              nil,
		cd:                nil,
		contextWithHeader: nil,
		username:          username,
		password:          password,
		OfflineFolder:     "/115/云下载",
		UploadFolder:      "/115/tg",
	}
	return &c
}

func (c *Client) Close() {
	if c.conn != nil {
		_ = c.conn.Close()
	}
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

func (c *Client) Set115Cookie(ck string) error {
	res, err := c.cd.APILogin115Editthiscookie(c.contextWithHeader, &clouddrive.Login115EditthiscookieRequest{EditThiscookieString: ck})
	if err != nil {
		return err
	}
	if !res.Success {
		return errors.New(res.ErrorMessage)
	}
	return nil
}

func (c *Client) AddOfflineFiles(url string) ([]string, error) {
	res, err := c.cd.AddOfflineFiles(c.contextWithHeader, &clouddrive.AddOfflineFileRequest{Urls: url, ToFolder: c.UploadFolder})
	if err != nil {
		return nil, err
	}

	return res.GetResultFilePaths(), nil
}

func (c *Client) Upload(filePath, fileName string) error {
	var createFileResult *clouddrive.CreateFileResult
	var file *os.File
	if fileName == "" {
		fileName = filepath.Base(filePath)
	}
	defer func() {
		if file != nil {
			_ = file.Close()
		}
		if createFileResult != nil {
			_, _ = c.cd.CloseFile(c.contextWithHeader, &clouddrive.CloseFileRequest{FileHandle: createFileResult.FileHandle})
		}
	}()
	createFileResult, err := c.cd.CreateFile(c.contextWithHeader, &clouddrive.CreateFileRequest{ParentPath: c.UploadFolder, FileName: fileName})
	if err != nil {
		return err
	}
	// 打开文件
	file, err = os.Open(filePath)
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
			_, err = c.cd.WriteToFile(c.contextWithHeader, &clouddrive.WriteFileRequest{FileHandle: createFileResult.FileHandle, StartPos: offset, Length: uint64(n), Buffer: data[:n], CloseFile: false})
			if err != nil {
				return err
			}
			//fmt.Println(res.GetBytesWritten())
			offset += uint64(n)
		}
	}
	return nil
}

func (c *Client) GetSubFiles(path string, forceRefresh bool, checkExpires bool) (*clouddrive.SubFilesReply, error) {

	res, err := c.cd.GetSubFiles(c.contextWithHeader, &clouddrive.ListSubFileRequest{Path: path, ForceRefresh: forceRefresh, CheckExpires: &checkExpires})
	if err != nil {
		return nil, err
	}
	subFilesReply, err := res.Recv()
	if err != nil {
		return nil, err
	}
	return subFilesReply, err
}
