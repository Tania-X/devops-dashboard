package infra

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/crypto/ssh"
)

// SSH 远程命令执行与文件上传(Agent 部署/停止的基础设施)。
// 从 Service 抽离:domain/service 不直接依赖 SSH 细节。
type SSH struct{}

// Connect 建立 SSH 连接(密码认证)
//
// 安全取舍(显式说明):HostKeyCallback 使用 InsecureIgnoreHostKey 跳过主机密钥
// 校验,存在 MITM 风险。本部署场景为可信环境 + 密码认证的个人学习项目,接受该
// 风险;生产环境应改为 known_hosts 校验或固定主机指纹(ssh.KnownHosts / 自校验)。
func (s *SSH) Connect(host string, port int, user, password string) (*ssh.Client, error) {
	config := &ssh.ClientConfig{
		User:            user,
		Auth:            []ssh.AuthMethod{ssh.Password(password)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
	}
	return ssh.Dial("tcp", fmt.Sprintf("%s:%d", host, port), config)
}

// Exec 在远端执行一条命令;stderr 并入错误信息便于排查
func (s *SSH) Exec(client *ssh.Client, cmd string) error {
	session, err := client.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()
	var stderr bytes.Buffer
	session.Stderr = &stderr
	if err := session.Run(cmd); err != nil {
		return fmt.Errorf("%s: %w", stderr.String(), err)
	}
	return nil
}

// Upload 通过 scp 协议上传本地文件到远端路径
func (s *SSH) Upload(client *ssh.Client, localPath, remotePath string) error {
	data, err := os.ReadFile(localPath)
	if err != nil {
		return fmt.Errorf("读取本地文件失败: %w", err)
	}
	session, err := client.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()
	go func() {
		w, _ := session.StdinPipe()
		defer w.Close()
		fmt.Fprintf(w, "C%#o %d %s\n", 0755, len(data), filepath.Base(remotePath))
		w.Write(data)
		fmt.Fprint(w, "\x00")
	}()
	return session.Run(fmt.Sprintf("scp -t %s", remotePath))
}
