/**
 * MQTT 客户端配置与连接（基于 Eclipse Paho，Go 原生库）。
 * 供内置 mqttIn / mqttOut 及服务端子订阅管理复用。
 */
package mqtt

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"strings"
	"time"

	paho "github.com/eclipse/paho.mqtt.golang"
)

// Config 连接 Broker 的通用参数。
type Config struct {
	Server                 string
	Username               string
	Password               string
	ClientID               string
	CleanSession           bool
	MaxReconnectInterval   time.Duration
	CAFile                 string
	CertFile               string
	CertKeyFile            string
}

// NormalizeServer 补全 tcp:// / ssl:// 前缀。
func NormalizeServer(server string) string {
	s := strings.TrimSpace(server)
	if s == "" {
		return s
	}
	lower := strings.ToLower(s)
	if strings.HasPrefix(lower, "tcp://") ||
		strings.HasPrefix(lower, "ssl://") ||
		strings.HasPrefix(lower, "tls://") ||
		strings.HasPrefix(lower, "ws://") ||
		strings.HasPrefix(lower, "wss://") {
		return s
	}
	return "tcp://" + s
}

// NewPahoOptions 构建 Paho 客户端选项（含自动重连）。
func NewPahoOptions(cfg Config) (*paho.ClientOptions, error) {
	opts := paho.NewClientOptions()
	server := NormalizeServer(cfg.Server)
	if server == "" {
		return nil, fmt.Errorf("mqtt: server required")
	}
	opts.AddBroker(server)
	if cfg.Username != "" {
		opts.SetUsername(cfg.Username)
	}
	if cfg.Password != "" {
		opts.SetPassword(cfg.Password)
	}
	if cfg.ClientID != "" {
		opts.SetClientID(cfg.ClientID)
	}
	opts.SetCleanSession(cfg.CleanSession)
	opts.SetAutoReconnect(true)
	opts.SetConnectRetry(true)
	if cfg.MaxReconnectInterval > 0 {
		opts.SetMaxReconnectInterval(cfg.MaxReconnectInterval)
	} else {
		opts.SetMaxReconnectInterval(60 * time.Second)
	}
	opts.SetKeepAlive(30 * time.Second)
	opts.SetPingTimeout(10 * time.Second)
	opts.SetWriteTimeout(10 * time.Second)

	tlsCfg, err := buildTLS(cfg)
	if err != nil {
		return nil, err
	}
	if tlsCfg != nil {
		opts.SetTLSConfig(tlsCfg)
	}
	return opts, nil
}

// Connect 创建并连接客户端；超时由 connectTimeout 控制。
func Connect(cfg Config, connectTimeout time.Duration) (paho.Client, error) {
	opts, err := NewPahoOptions(cfg)
	if err != nil {
		return nil, err
	}
	c := paho.NewClient(opts)
	if connectTimeout <= 0 {
		connectTimeout = 15 * time.Second
	}
	token := c.Connect()
	if !token.WaitTimeout(connectTimeout) {
		c.Disconnect(250)
		return nil, fmt.Errorf("mqtt: connect timeout to %s", NormalizeServer(cfg.Server))
	}
	if err := token.Error(); err != nil {
		return nil, fmt.Errorf("mqtt: connect %s: %w", NormalizeServer(cfg.Server), err)
	}
	return c, nil
}

func buildTLS(cfg Config) (*tls.Config, error) {
	hasFiles := cfg.CAFile != "" || cfg.CertFile != "" || cfg.CertKeyFile != ""
	server := strings.ToLower(NormalizeServer(cfg.Server))
	needTLS := hasFiles || strings.HasPrefix(server, "ssl://") ||
		strings.HasPrefix(server, "tls://") || strings.HasPrefix(server, "wss://")
	if !needTLS {
		return nil, nil
	}
	tlsCfg := &tls.Config{MinVersion: tls.VersionTLS12}
	if cfg.CAFile != "" {
		pem, err := os.ReadFile(cfg.CAFile)
		if err != nil {
			return nil, fmt.Errorf("mqtt: read caFile: %w", err)
		}
		pool := x509.NewCertPool()
		if !pool.AppendCertsFromPEM(pem) {
			return nil, fmt.Errorf("mqtt: invalid caFile PEM")
		}
		tlsCfg.RootCAs = pool
	}
	if cfg.CertFile != "" && cfg.CertKeyFile != "" {
		cert, err := tls.LoadX509KeyPair(cfg.CertFile, cfg.CertKeyFile)
		if err != nil {
			return nil, fmt.Errorf("mqtt: load client cert: %w", err)
		}
		tlsCfg.Certificates = []tls.Certificate{cert}
	}
	return tlsCfg, nil
}

// ClampQoS 将 qos 限制在 0..2。
func ClampQoS(q int) byte {
	if q < 0 {
		return 0
	}
	if q > 2 {
		return 2
	}
	return byte(q)
}
